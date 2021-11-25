package discord

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
	"pkg.mon.icu/monicu/internal/storage/model"
	"pkg.mon.icu/monicu/internal/util"
)

// Guilds/channels

// createChannelsAndGuilds populates database with entries for guilds and channels that are
// defined in config.
func (d *Discord) createChannelsAndGuilds() {
	for _, g := range d.config.guilds.Values() {
		// Guild
		gm := model.ForGuildID(util.FormatSnowflake(g))
		if err := d.createGuild(gm); err != nil {
			d.logger.Errorf("Failed to create guild: %s.", err)
			continue
		}

		// Channels
		chs, err := d.session.GuildChannels(util.FormatSnowflake(g))
		if err != nil {
			d.logger.Errorf("Failed to retrieve messages: %s.", err)
			continue
		}
		for _, ch := range chs {
			if d.config.chans.Contains(util.MustParseSnowflake(ch.ID)) {
				cm := model.ForChannelID(ch.ID)
				cm.GuildID = gm.ID
				if err := d.createChannel(cm); err != nil {
					d.logger.Errorf("Failed to create channel: %s.", err)
				}
			}
		}
	}
}

// createGuild creates (or finds, effectively creating one if it did not exist) guild entry in database
// for the specified guild _model.
func (d *Discord) createGuild(gm *model.Guild) error {
	return d.storage.Transaction(func(db *gorm.DB) error {
		return db.WithContext(d.ctx).FirstOrCreate(gm, "discord_id = ?", gm.DiscordID).Error
	})
}

// createChannel created (or finds, effectively creating one if it did not exist) channel entry in database
// for the specified channel _model.
func (d *Discord) createChannel(cm *model.Channel) error {
	return d.storage.Transaction(func(db *gorm.DB) error {
		return db.WithContext(d.ctx).FirstOrCreate(cm, "discord_id = ?", cm.DiscordID).Error
	})
}

// Channel sync

// isSyncRequired checks if channel with the specified ID has no posts and requires initial synchronization.
func (d *Discord) isSyncRequired(DiscordID uint64) (bool, error) {
	var posts int64
	if err := d.storage.Transaction(func(db *gorm.DB) error {
		cm := model.ForChannelIDUint(DiscordID)
		if err := db.WithContext(d.ctx).First(cm, "discord_id = ?", DiscordID).Error; err != nil {
			return err
		}
		if err := db.WithContext(d.ctx).Model(&model.Post{}).Where("channel_id = ?", cm.ID).Count(&posts).Error; err != nil {
			return err
		}
		return nil
	}); err != nil && !errors.Is(err, context.Canceled) {
		return false, err
	}
	return posts == 0, nil
}

// syncChannel performs initial synchronization of channel with the specified ID.
func (d *Discord) syncChannel(ID string) {
	var beforeID string
	for {
		ms, err := d.session.ChannelMessages(ID, 100, beforeID, "", "")
		if err != nil {
			return
		}
		if len(ms) == 0 {
			break
		}
		for _, m := range ms {
			if d.ctx.Err() != nil {
				return
			}

			// todo concurrency
			d.createPost(m)
		}

		beforeID = ms[len(ms)-1].ID
	}
}

// syncChannel attempts synchronization of all channels defined in config in parallel skipping channels
// that do not require initial synchronization.
func (d *Discord) syncChannels() {
	for _, c := range d.config.chans.Values() {
		sync, err := d.isSyncRequired(c)
		if err != nil && !errors.Is(err, context.Canceled) {
			d.logger.Errorf("Failed to check if channel %d should be synchronized: %s.", c, err)
			continue
		}

		if sync {
			// todo concurrency
			d.syncChannel(util.FormatSnowflake(c))
		}
	}
}

// Posts

// isValidPost checks is message makes a valid post with images, returning false for messages that have no
// image attachments or embeds.
func (d *Discord) isValidPost(m *discordgo.Message) bool {
	if len(m.Attachments) == 0 && len(m.Embeds) == 0 {
		return false
	}
	if d.config.ignoreRegexp.MatchString(m.Content) {
		return false
	}
	return true
}

func (d *Discord) createPostReactions(db *gorm.DB, m *discordgo.Message, pm *model.Post) error {
	for _, r := range m.Reactions {
		em := model.ForEmoji(r.Emoji)
		var err error
		if em.IsGuild() {
			err = db.WithContext(d.ctx).FirstOrCreate(em, "discord_id = ?", em.DiscordID).Error
		} else {
			err = db.WithContext(d.ctx).FirstOrCreate(em, "name = ?", em.Name).Error
		}
		if err != nil {
			return err
		}

		rm := model.NewReaction(pm.ID, em.ID)
		if err := db.WithContext(d.ctx).Create(rm).Error; err != nil {
			return err
		}

		var afterID string
		for {
			us, err := d.session.MessageReactions(m.ChannelID, m.ID, r.Emoji.APIName(), 100, "", afterID)
			if err != nil {
				return err
			}
			if len(us) == 0 {
				break
			}
			for _, u := range us {
				um := model.ForUserID(u.ID)
				if err := db.WithContext(d.ctx).FirstOrCreate(um, "discord_id = ?", um.DiscordID).Error; err != nil {
					return err
				}

				urm := model.NewUserReaction(rm.ID, um.ID)
				if err := db.WithContext(d.ctx).Create(urm).Error; err != nil {
					return err
				}
			}
			afterID = us[len(us)-1].ID
		}
	}
	return nil
}

// createPostImages creates images for a post.
func (d *Discord) createPostImages(db *gorm.DB, m *discordgo.Message, pm *model.Post) error {
	for _, at := range m.Attachments {
		if at.Width != 0 || at.Height != 0 {
			im := model.ForDiscordAttachment(at)
			im.Post = pm
			if err := db.WithContext(d.ctx).Model(pm).Association("Images").Append(im); err != nil {
				return fmt.Errorf("failed to create image: %w", err)
			}
		} else {
			// not an image attachment
		}
	}
	for _, e := range m.Embeds {
		if e.Image != nil {
			im, err := model.ForDiscordEmbed(d.ctx, e)
			if err != nil {
				return fmt.Errorf("failed to wrap embed: %w", err)
			}

			im.Post = pm
			if err := db.WithContext(d.ctx).Model(pm).Association("Images").Append(im); err != nil {
				return fmt.Errorf("failed to create image: %w", err)
			}
		} else {
			// not an image attachment
		}
	}
	return nil
}

func (d *Discord) createPostBase(db *gorm.DB, m *discordgo.Message) (*model.Post, error) {
	if m.GuildID == "" {
		// In cases such as channel synchronization message will likely lack GuildID
		// So we pull it from the cache
		m.GuildID = util.FormatSnowflake(d.channelGuildRelations[util.MustParseSnowflake(m.ChannelID)])
	}

	gm := model.ForGuildID(m.GuildID)
	if err := db.WithContext(d.ctx).First(gm, "discord_id = ?", gm.DiscordID).Error; err != nil {
		return nil, err
	}

	cm := model.ForChannelID(m.ChannelID)
	cm.Guild = gm
	if err := db.WithContext(d.ctx).First(cm, "discord_id = ?", cm.DiscordID).Error; err != nil {
		return nil, err
	}

	um := model.ForUserID(m.Author.ID)
	if err := db.WithContext(d.ctx).FirstOrCreate(um, "discord_id = ?", um.DiscordID).Error; err != nil {
		return nil, err
	}

	pm := model.ForMessage(m)
	pm.Channel, pm.User = cm, um
	if err := db.WithContext(d.ctx).Create(pm).Error; err != nil {
		return nil, err
	}

	return pm, nil
}

// createPost creates a post from Discord message.
func (d *Discord) createPost(m *discordgo.Message) {
	if !d.isValidPost(m) {
		return
	}

	if err := d.storage.Transaction(func(db *gorm.DB) error {
		pm, err := d.createPostBase(db, m)
		if err != nil {
			return err
		}
		if err := d.createPostImages(db, m, pm); err != nil {
			return err
		}
		if err := d.createPostReactions(db, m, pm); err != nil {
			return err
		}
		d.logger.Infof("Created post for message %s.", m.ID)
		return nil
	}); err != nil && !errors.Is(err, context.Canceled) {
		d.logger.Errorf("Failed to create reaction: %s.", err)
	}
}

// updatePost updates a post (or creates one if an attachment- and embed-less message contained a link
// and was updated automatically server-side with attachment/embed) from Discord message.
func (d *Discord) updatePost(m *discordgo.Message) {
	if err := d.storage.Transaction(func(db *gorm.DB) error {
		pm := model.ForMessage(m)
		if err := db.WithContext(d.ctx).Find(pm, "discord_id = ?", pm.DiscordID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ogm, err := d.session.ChannelMessage(m.ChannelID, m.ID)
				if err != nil {
					return err
				}
				m.Author = ogm.Author
				d.createPost(m)
				return nil
			}
			return err
		}
		if len(m.Attachments) == 0 && len(m.Embeds) == 0 {
			d.deletePost(m)
			return nil
		}
		if err := db.WithContext(d.ctx).Delete(&model.Image{}, "post_id = ?", pm.ID).Error; err != nil {
			return err
		}
		if err := d.createPostImages(db, m, pm); err != nil {
			return err
		}
		if err := db.WithContext(d.ctx).Save(pm).Error; err != nil {
			return err
		}
		d.logger.Infof("Updated post for message %s.", m.ID)
		return nil
	}); err != nil && !errors.Is(err, context.Canceled) {
		d.logger.Errorf("Failed to update post: %s.", err)
	}
}

// deletePost deletes a post from Discord message.
func (d *Discord) deletePost(m *discordgo.Message) {
	if err := d.storage.Transaction(func(db *gorm.DB) error {
		pm := model.ForMessage(m)
		if err := db.WithContext(d.ctx).Delete(pm, "discord_id = ?", pm.DiscordID).Error; err != nil {
			return err
		}
		d.logger.Infof("Deleted post for message %s.", m.ID)
		return nil
	}); err != nil && !errors.Is(err, context.Canceled) {
		d.logger.Errorf("Failed to delete post: %s.", err)
	}
}

// deletePostsBulk deletes a number of posts from array of Discord IDs.
func (d *Discord) deletePostsBulk(messages []string) {
	for _, m := range messages {
		d.deletePost(&discordgo.Message{ID: m})
	}
}

// Reactions

// addReaction adds reaction to post loaded from the database for the message that is tied to the specified reaction.
func (d *Discord) addReaction(r *discordgo.MessageReaction) {
	if err := d.storage.Transaction(func(db *gorm.DB) error {
		pm := model.ForMessageID(r.MessageID)
		if err := db.WithContext(d.ctx).First(pm).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		em := model.ForEmoji(&r.Emoji)
		var err error
		if em.IsGuild() {
			err = db.WithContext(d.ctx).FirstOrCreate(em, "discord_id = ?", em.DiscordID).Error
		} else {
			err = db.WithContext(d.ctx).FirstOrCreate(em, "name = ?", em.Name).Error
		}
		if err != nil {
			return err
		}

		rm := model.NewReaction(pm.ID, em.ID)
		if err := db.WithContext(d.ctx).FirstOrCreate(rm, "post_id = ? and emoji_id = ?", pm.ID, em.ID).Error; err != nil {
			return err
		}

		um := model.ForUserID(r.UserID)
		if err := db.WithContext(d.ctx).FirstOrCreate(um, "discord_id = ?", um.DiscordID).Error; err != nil {
			return err
		}

		urm := model.NewUserReaction(rm.ID, um.ID)
		if err := db.WithContext(d.ctx).Model(rm).Association("UserReactions").Append(urm); err != nil {
			return err
		}

		d.logger.Infof("Added reaction to post for for message %s.", r.MessageID)
		return nil
	}); err != nil && !errors.Is(err, context.Canceled) {
		d.logger.Errorf("Failed to add reaction: %s.", err)
	}
}

// removeReaction removes reaction from post loaded from the database for the message that is tied to the specified reaction.
func (d *Discord) removeReaction(r *discordgo.MessageReaction) {
	if err := d.storage.Transaction(func(db *gorm.DB) error {
		pm := model.ForMessageID(r.MessageID)
		if err := db.WithContext(d.ctx).First(pm, "discord_id = ?", pm.DiscordID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}

			return err
		}

		em := model.ForEmoji(&r.Emoji)
		var err error
		if em.IsGuild() {
			err = db.WithContext(d.ctx).FirstOrCreate(em, "discord_id = ?", em.DiscordID).Error
		} else {
			err = db.WithContext(d.ctx).FirstOrCreate(em, "name = ?", em.Name).Error
		}
		if err != nil {
			return err
		}

		rm := model.NewReaction(pm.ID, em.ID)
		if err := db.WithContext(d.ctx).FirstOrCreate(rm, "post_id = ? and emoji_id = ?", pm.ID, em.ID).Error; err != nil {
			return err
		}

		um := model.ForUserID(r.UserID)
		if err := db.WithContext(d.ctx).FirstOrCreate(um, "discord_id = ?", um.DiscordID).Error; err != nil {
			return err
		}

		if err := db.WithContext(d.ctx).Delete(&model.UserReaction{}, "reaction_id = ? and user_id = ?", rm.ID, um.ID).Error; err != nil {
			return err
		}

		d.logger.Infof("Removed reaction from post for for message %s.", r.MessageID)
		return nil
	}); err != nil {
		d.logger.Errorf("Failed to remove reaction: %s.", err)
	}
}

// removeReactionsBulk removes all reactions from post loaded from the database for the message that is tied to the specified reaction.
func (d *Discord) removeReactionsBulk(r *discordgo.MessageReaction) {
	if err := d.storage.Transaction(func(db *gorm.DB) error {
		pm := model.ForMessageID(r.MessageID)
		if err := db.WithContext(d.ctx).First(pm, "discord_id = ?", pm.DiscordID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}

			return err
		}

		if err := db.WithContext(d.ctx).Delete(&model.Reaction{}, "post_id = ?", pm.ID).Error; err != nil {
			return err
		}

		d.logger.Infof("Removed all reactions from post for for message %s.", r.MessageID)
		return nil
	}); err != nil {
		d.logger.Errorf("Failed to remove all reactions: %s.", err)
	}
}
