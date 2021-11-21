package discord

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
	"pkg.mon.icu/monicu/internal/storage/model"
)

func (d *Discord) createChannelsAndGuilds() {
	for g := range d.config.guilds {
		gm := model.WrapGuildID(strconv.FormatUint(g, 10))
		if err := d.createGuild(gm); err != nil {
			d.logger.Errorf("Failed to create guild: %s.", err)
		}

		c, err := d.session.GuildChannels(strconv.FormatUint(g, 10))
		if err != nil {
			d.logger.Errorf("Failed to retrieve guild channels: %s.", err)
		}
		for _, ch := range c {
			if d.config.chans.Contains(model.MustParseSnowflake(ch.ID)) {
				cm := model.WrapChannelID(ch.ID)
				cm.GuildID = gm.ID
				if err := d.createChannel(cm); err != nil {
					d.logger.Errorf("Failed to create channel: %s.", err)
				}
			}
		}
	}
}

func (d *Discord) createGuild(gm *model.Guild) error {
	return d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		return model.FindOrCreateGuild(d.ctx, tx, gm)
	})
}

func (d *Discord) createChannel(cm *model.Channel) error {
	return d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		return model.FindOrCreateChannel(d.ctx, tx, cm)
	})
}

func isValidPost(m *discordgo.Message) bool {
	return !(len(m.Attachments) == 0 && len(m.Embeds) == 0)
}

func (d *Discord) isSyncRequired(ID string) (bool, error) {
	var empty bool
	return empty, d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		cm := model.WrapChannelID(ID)
		if err := model.FindChannel(d.ctx, tx, cm); err != nil {
			return fmt.Errorf("failed to find or create channel: %w", err)
		}
		if cm.ID == 0 {
			return errors.New("channel is not in database")
		}

		var err error
		if empty, err = model.IsChannelEmpty(d.ctx, tx, cm); err != nil {
			return fmt.Errorf("failed to check if channel is empty: %w", err)
		}
		return nil
	})
}

func (d *Discord) syncChannel(ID string) {
	var beforeID string
	for {
		ms, err := d.session.ChannelMessages(ID, 100, beforeID, "", "")
		if err != nil {
			d.logger.Errorf("Failed to fetch messages: %s.", err)
			return
		}

		if len(ms) == 0 {
			break
		}

		for _, m := range ms {
			if isValidPost(m) {
				if err := d.createPost(m); err != nil {
					d.logger.Errorf("Failed to create post: %s.", err)
				}
			}
		}

		beforeID = ms[len(ms)-1].ID
	}
}

func (d *Discord) syncChannels() {
	for c := range d.config.chans {
		cID := strconv.FormatUint(c, 10)

		sync, err := d.isSyncRequired(cID)
		if err != nil {
			d.logger.Errorf("Failed to check if channel %d should be synchronized: %s.", c, err)
			continue
		}

		if sync {
			go d.syncChannel(cID)
		}
	}
}

func (d *Discord) createPostImages(tx pgx.Tx, m *discordgo.Message, pm *model.Post) error {
	for _, at := range m.Attachments {
		if at.Width != 0 || at.Height != 0 {
			im := model.WrapDiscordAttachment(at)
			im.PostID = pm.ID
			if err := model.CreateImage(d.ctx, tx, im); err != nil {
				return fmt.Errorf("failed to create image: %w", err)
			}
		} else {
			// not an image attachment
		}
	}

	for _, e := range m.Embeds {
		if e.Image != nil {
			im, err := model.WrapDiscordEmbed(d.ctx, e)
			if err != nil {
				return fmt.Errorf("failed to wrap embed: %w", err)
			}

			im.PostID = pm.ID
			if err := model.CreateImage(d.ctx, tx, im); err != nil {
				return fmt.Errorf("failed to create image: %w", err)
			}
		} else {
			// not an image attachment
		}
	}

	return nil
}

func (d *Discord) createPost(m *discordgo.Message) error {
	return d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		if m.GuildID == "" {
			// In cases such as channel synchronization message will likely lack GuildID
			// So we pull it from the cache
			m.GuildID = strconv.FormatUint(d.channelGuildRelations[model.MustParseSnowflake(m.ChannelID)], 10)
		}

		gm := model.WrapGuildID(m.GuildID) // guild model
		if err := model.FindGuild(d.ctx, tx, gm); err != nil {
			return fmt.Errorf("failed to find or create guild: %w", err)
		}
		if gm.ID == 0 {
			return errors.New("guild is not in database")
		}

		cm := model.WrapChannelID(m.ChannelID) // channel model
		cm.GuildID = gm.ID
		if err := model.FindChannel(d.ctx, tx, cm); err != nil {
			return fmt.Errorf("failed to find or create channel: %w", err)
		}
		if cm.ID == 0 {
			return errors.New("channel is not in database")
		}

		um := model.WrapUserID(m.Author.ID) // user model
		if err := model.FindOrCreateUser(d.ctx, tx, um); err != nil {
			return fmt.Errorf("failed to find or create user: %w", err)
		}

		pm := model.WrapDiscordMessage(m) // post model
		pm.ChannelID, pm.UserID = cm.ID, um.ID
		if err := model.CreatePost(d.ctx, tx, pm); err != nil {
			return fmt.Errorf("failed to find or create channel: %w", err)
		}

		for i, at := range m.Attachments {
			if at.Width == 0 && at.Height == 0 { // not an image attachment
				d.logger.Debugf("Skipping non-image attachment %s (index %d).", at.ID, i)
			} else {
				d.logger.Debugf("Creating image for attachmend ID %s (index %d).", at.ID, i)
				im := model.WrapDiscordAttachment(at)
				im.PostID = pm.ID
				if err := model.CreateImage(d.ctx, tx, im); err != nil {
					return fmt.Errorf("failed to create image: %w", err)
				}
			}
		}

		for i, e := range m.Embeds {
			if e.Image == nil { // not an image attachment
				d.logger.Debugf("Skipping non-image embed index %d.", i)
			} else {
				d.logger.Debugf("Creating image for embed index %d.", i)
				im, err := model.WrapDiscordEmbed(d.ctx, e)
				if err != nil {
					return fmt.Errorf("failed to wrap embed: %w", err)
				}

				im.PostID = pm.ID
				if err := model.CreateImage(d.ctx, tx, im); err != nil {
					return fmt.Errorf("failed to create image: %w", err)
				}
			}
		}

		for i, mr := range m.Reactions {
			d.logger.Debugf("Creating reactions for emoji %s (index %d).", mr.Emoji.Name, i)

			em := model.WrapDiscordEmoji(mr.Emoji) // emoji model
			if err := model.FindOrCreateEmoji(d.ctx, tx, em); err != nil {
				return fmt.Errorf("failed to find or create emoji: %w", err)
			}

			rm := model.NewReaction() // reaction model
			rm.PostID, rm.EmojiID = pm.ID, em.ID
			if err := model.CreateReaction(d.ctx, tx, rm); err != nil {
				return fmt.Errorf("failed to create reaction: %w", err)
			}

			var afterID string
			for {
				ur, err := d.session.MessageReactions(m.ChannelID, m.ID, mr.Emoji.APIName(), 100, "", afterID)
				if err != nil {
					return fmt.Errorf("failed to fetch user reactions: %w", err)
				}

				if len(ur) == 0 {
					break
				}

				for _, u := range ur {
					rum := model.WrapUserID(u.ID) // reacted user model
					if err := model.FindOrCreateUser(d.ctx, tx, rum); err != nil {
						return fmt.Errorf("failed to find or create user for Discord ID %s: %w", u.ID, err)
					}

					urm := model.NewUserReaction() // user reaction model
					urm.ReactionID, urm.UserID = rm.ID, rum.ID
					if err := model.CreateUserReaction(d.ctx, tx, urm); err != nil {
						return fmt.Errorf("failed to create user reaction: %w", err)
					}
				}

				afterID = ur[len(ur) - 1].ID
			}
		}

		return nil
	})
}

func (d *Discord) addReaction(r *discordgo.MessageReaction) error {
	return d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		pm := model.WrapMessageID(r.MessageID)
		if err := model.FindPost(d.ctx, tx, pm); err != nil {
			return fmt.Errorf("failed to find post: %w", err)
		}
		if pm.ID == 0 {
			return errors.New("post is not in database")
		}

		em := model.WrapDiscordEmoji(&r.Emoji)
		if err := model.FindOrCreateEmoji(d.ctx, tx, em); err != nil {
			return fmt.Errorf("failed to find or create emoji: %w", err)
		}

		rm := model.NewReaction()
		rm.PostID, rm.EmojiID = pm.ID, em.ID
		if err := model.FindOrCreateReaction(d.ctx, tx, rm); err != nil {
			return fmt.Errorf("failed to find or create reaction: %w", err)
		}

		um := model.WrapUserID(r.UserID)
		if err := model.FindOrCreateUser(d.ctx, tx, um); err != nil {
			return fmt.Errorf("failed to find or create user: %w", err)
		}

		urm := model.NewUserReaction()
		urm.ReactionID, urm.UserID = rm.ID, um.ID
		if err := model.CreateUserReaction(d.ctx, tx, urm); err != nil {
			return fmt.Errorf("failed to create user reaction: %w", err)
		}

		return nil
	})
}

func (d *Discord) removeReaction(r *discordgo.MessageReaction) error {
	return d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		pm := model.WrapMessageID(r.MessageID)
		if err := model.FindPost(d.ctx, tx, pm); err != nil {
			return fmt.Errorf("failed to find post: %w", err)
		}
		if pm.ID == 0 {
			return errors.New("post is not in database")
		}

		em := model.WrapDiscordEmoji(&r.Emoji)
		if err := model.FindOrCreateEmoji(d.ctx, tx, em); err != nil {
			return fmt.Errorf("failed to find or create emoji: %w", err)
		}

		rm := model.NewReaction()
		rm.PostID, rm.EmojiID = pm.ID, em.ID
		if err := model.FindOrCreateReaction(d.ctx, tx, rm); err != nil {
			return fmt.Errorf("failed to find or create reaction: %w", err)
		}

		um := model.WrapUserID(r.UserID)
		if err := model.FindOrCreateUser(d.ctx, tx, um); err != nil {
			return fmt.Errorf("failed to find or create user: %w", err)
		}

		urm := model.NewUserReaction()
		urm.ReactionID, urm.UserID = rm.ID, um.ID
		if _, err := model.DeleteUserReaction(d.ctx, tx, urm); err != nil {
			return fmt.Errorf("failed to create user reaction: %w", err)
		}

		return nil
	})
}

func (d *Discord) removeReactionsBulk(r *discordgo.MessageReaction) error {
	return d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		pm := model.WrapMessageID(r.MessageID)
		if err := model.FindPost(d.ctx, tx, pm); err != nil {
			return fmt.Errorf("failed to find post: %w", err)
		}
		if pm.ID == 0 {
			return errors.New("post is not in database")
		}

		if _, err := model.DeleteAllReactions(d.ctx, tx, pm); err != nil {
			return fmt.Errorf("failed to delete all reactions: %w", err)
		}

		return nil
	})
}

func (d *Discord) updatePost(m *discordgo.Message) error {
	return d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		pm := model.WrapDiscordMessage(m)
		if err := model.FindPost(d.ctx, tx, pm); err != nil {
			return fmt.Errorf("failed to find post: %w", err)
		}
		if pm.ID == 0 {
			return nil
		}

		if _, err := model.DeletePostImages(d.ctx, tx, pm); err != nil {
			return fmt.Errorf("failed to delete post images: %w", err)
		}
		if err := d.createPostImages(tx, m, pm); err != nil {
			return fmt.Errorf("failed to create post images: %w", err)
		}
		if _, err := model.UpdatePost(d.ctx, tx, pm); err != nil {
			return fmt.Errorf("failed to update post: %w", err)
		}

		return nil
	})
}

func (d *Discord) deletePost(m *discordgo.Message) error {
	return d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		pm := model.WrapDiscordMessage(m)
		if _, err := model.DeletePost(d.ctx, tx, pm); err != nil {
			return fmt.Errorf("failed to find post: %w", err)
		}

		return nil
	})
}
