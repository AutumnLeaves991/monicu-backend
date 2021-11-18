package discord

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
	"pkg.mon.icu/monicu/internal/storage/entity"
)

func (d *Discord) maybeSyncChannels() {
	d.logger.Info("Synchronizing channels.")
	for chanID := range d.config.chans {
		d.logger.Sugar().Infof("Synchronizing channel %d.", chanID)

		var empty bool
		if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
			ec := entity.NewChannel(0, chanID, 0)
			if err := entity.FindChannel(d.ctx, tx, ec); err != nil {
				return err
			}

			if chanEmpty, err := entity.IsChannelEmpty(d.ctx, tx, ec); err != nil {
				return err
			} else if chanEmpty {
				empty = true
			}

			return nil
		}); err != nil {
			d.logger.Sugar().Errorf("Failed to synchronize channel %d: %s.", chanID, err)
			return
		}

		if !empty {
			d.logger.Sugar().Debugf("Channel %d is not empty in database, skipping first sync.", chanID)
			return
		}

		chanObj, err := d.session.Channel(strconv.FormatUint(chanID, 10))
		if err != nil {
			d.logger.Sugar().Errorf("Failed to fetch channel %d: %s.", chanID, err)
			return
		}

		chanID := chanID
		go func() {
			var beforeID string
			for {
				msg, err := d.session.ChannelMessages(strconv.FormatUint(chanID, 10), 100, beforeID, "", "")
				if err != nil {
					d.logger.Sugar().Errorf("Failed to request channel messages from channel %d: %s.", chanID, err)
					return
				}

				if len(msg) == 0 {
					break
				}

				for _, m := range msg {
					m.GuildID = chanObj.GuildID
					d.maybeCreatePost(m)
				}

				beforeID = msg[len(msg)-1].ID

				if len(msg) < 100 {
					break
				}
			}
		}()
	}
}

func (d *Discord) maybeCreatePost(m *discordgo.Message) {
	eg := entity.NewGuildFromSnowflakeID(m.GuildID)
	if !d.config.guilds.Contains(eg.DiscordID) {
		d.logger.Sugar().Debugf("Ignoring message %s from ignored guild %s.", m.ID, m.GuildID)
		return
	}

	ec := entity.NewChannelFromSnowflakeID(m.ChannelID, 0)
	if !d.config.chans.Contains(ec.DiscordID) {
		d.logger.Sugar().Debugf("Ignoring message %s from ignored channel %s.", m.ID, m.ChannelID)
		return
	}

	if d.config.ignoreRegexp.MatchString(m.Content) {
		d.logger.Sugar().Debugf("Ignoring message %s because it matches ignore regexp.", m.ID)
		return
	}

	if len(m.Attachments) == 0 && len(m.Embeds) == 0 {
		d.logger.Sugar().Debugf("Skipping attachmentless/embedless message %s.", m.ID)
		return
	}

	eu := entity.NewUserFromSnowflakeID(m.Author.ID)
	ep := entity.NewPostFromSnowflakeID(m.ID, 0, 0, m.Content)

	d.logger.Sugar().Debugf("Creating post for message %s.", m.ID)
	if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		d.logger.Sugar().Debugf("Creating guild ID %s.", m.GuildID)
		if err := entity.FindOrCreateGuild(d.ctx, tx, eg); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating channel ID %s.", m.ChannelID)
		ec.GuildID = eg.ID
		if err := entity.FindOrCreateChannel(d.ctx, tx, ec); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating user ID %s.", m.Author.ID)
		if err := entity.FindOrCreateUser(d.ctx, tx, eu); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating post ID %s.", m.ID)
		ep.ChannelID, ep.UserID = ec.ID, eu.ID
		if err := entity.CreatePost(d.ctx, tx, ep); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating attachments for post %s.", m.ID)
		for i, at := range m.Attachments {
			if at.Width != 0 && at.Height != 0 {
				im := entity.NewImageFromAttachment(at, ep.ID)
				if err := entity.CreateImage(d.ctx, tx, im); err != nil {
					return err
				}
			} else {
				d.logger.Sugar().Debugf("Ignoring non-image attachment #%d %s.", i, at.ID)
			}
		}
		for i, em := range m.Embeds {
			if em.Image != nil {
				d.logger.Sugar().Debugf("Creating image struct from embed #%d.", i)
				im, err := entity.NewImageFromEmbed(d.ctx, em, ep.ID)
				if err != nil {
					return err
				}
				if err := entity.CreateImage(d.ctx, tx, im); err != nil {
					return err
				}
			} else {
				d.logger.Sugar().Debugf("Ignoring non-image embed #%d.", i)
			}
		}

		for _, r := range m.Reactions {
			em := entity.NewEmojiFromDiscord(r.Emoji)
			d.logger.Sugar().Debugf("Creating emoji %s.", r.Emoji.APIName())
			if err := entity.FindOrCreateEmoji(d.ctx, tx, em); err != nil {
				return err
			}

			er := entity.NewReaction(0, ep.ID, em.ID)
			d.logger.Sugar().Debugf("Creating reaction for emoji %s.", r.Emoji.APIName())
			if err := entity.CreateReaction(d.ctx, tx, er); err != nil {
				return err
			}

			var beforeID string
			for {
				re, err := d.session.MessageReactions(m.ChannelID, m.ID, r.Emoji.APIName(), 100, beforeID, "")
				if err != nil {
					return err
				}

				if len(re) == 0 {
					break
				}

				for _, u := range re {
					eru := entity.NewUserFromSnowflakeID(u.ID)
					d.logger.Sugar().Debugf("Creating user ID %s.", u.ID)
					if err := entity.FindOrCreateUser(d.ctx, tx, eru); err != nil {
						return err
					}

					eur := entity.NewUserReaction(0, er.ID, eru.ID)
					d.logger.Sugar().Debugf("Creating reaction for emoji %s.", r.Emoji.APIName())
					if err := entity.CreateUserReaction(d.ctx, tx, eur); err != nil {
						return err
					}
				}

				beforeID = re[len(re)-1].ID

				if len(re) < 100 {
					break
				}
			}
		}

		return nil
	}); err != nil {
		d.logger.Sugar().Errorf("Failed to complete post creation transaction: %s.", err)
	} else {
		d.logger.Sugar().Infof("Finished creating post for message %s.", m.ID)
	}
}

func (d *Discord) maybeDeletePost(m *discordgo.Message) {
	if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		if ok, err := entity.DeletePost(d.ctx, tx, entity.NewPostFromSnowflakeID(m.ID, 0, 0, "")); err != nil {
			return err
		} else if ok {
			d.logger.Sugar().Infof("Deleted post %s.", m.ID)
		} /*else {
			d.logger.Sugar().Debugf("Attempted to delete post %s but SQL query returned zero affected rows.", m.ID)
		}*/

		return nil
	}); err != nil {
		d.logger.Sugar().Errorf("Failed to delete post %s: %s", m.ID, err)
	}
}

func (d *Discord) maybeUpdatePost(m *discordgo.Message) {
	eg := entity.NewGuildFromSnowflakeID(m.GuildID)
	if !d.config.guilds.Contains(eg.DiscordID) {
		d.logger.Sugar().Debugf("Ignoring message %s from ignored guild %s.", m.ID, m.GuildID)
		return
	}

	ec := entity.NewChannelFromSnowflakeID(m.ChannelID, 0)
	if !d.config.chans.Contains(ec.DiscordID) {
		d.logger.Sugar().Debugf("Ignoring message %s from ignored channel %s.", m.ID, m.ChannelID)
		return
	}

	if d.config.ignoreRegexp.MatchString(m.Content) {
		d.logger.Sugar().Debugf("Ignoring message %s because it matches ignore regexp.", m.ID)
		return
	}

	if len(m.Attachments) == 0 && len(m.Embeds) == 0 {
		d.logger.Sugar().Debugf("Skipping attachmentless/embedless message %s.", m.ID)
		return
	}

	ep := entity.NewPostFromSnowflakeID(m.ID, 0, 0, m.Content)
	if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		if err := entity.FindPost(d.ctx, tx, ep); err != nil {
			return err
		} else if ep.ID > 0 {
			// found
			// todo as editing involves potential embed/attachment removal, need to make
			// todo a routine that will 1) wipe all existing images attached 2) rescan attachements and embeds
		}

		return nil
	}); err != nil {
		d.logger.Sugar().Errorf("Failed to retrieve post %s: %s.", m.ID, err)
	} else if ep.ID == 0 {
		// not found, create it
		om, err := d.session.ChannelMessage(m.ChannelID, m.ID)
		if err != nil {
			d.logger.Sugar().Errorf("Failed to retrieve post %s: %s.", m.ID, err)
			return
		}

		m.Author = om.Author
		m.GuildID = strconv.FormatUint(d.guildChannelCache[entity.MustParseSnowflake(m.ChannelID)], 10)
		d.maybeCreatePost(m)
	}
}

func (d *Discord) maybeAddReaction(r *discordgo.MessageReaction) {
	eg := entity.NewGuildFromSnowflakeID(r.GuildID)
	if !d.config.guilds.Contains(eg.DiscordID) {
		d.logger.Sugar().Debugf("Ignoring message %s from ignored guild %s.", r.MessageID, r.GuildID)
		return
	}

	ec := entity.NewChannelFromSnowflakeID(r.ChannelID, 0)
	if !d.config.chans.Contains(ec.DiscordID) {
		d.logger.Sugar().Debugf("Ignoring message %s from ignored channel %s.", r.MessageID, r.ChannelID)
		return
	}

	eu := entity.NewUserFromSnowflakeID(r.UserID)
	ep := entity.NewPostFromSnowflakeID(r.MessageID, 0, 0, "")

	em := entity.NewEmojiFromDiscord(&r.Emoji)
	er := entity.NewReaction(0, 0, 0)

	if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		d.logger.Sugar().Debugf("Finding guild ID %s.", r.GuildID)
		if err := entity.FindGuild(d.ctx, tx, eg); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating user ID %s.", r.UserID)
		if err := entity.FindOrCreateUser(d.ctx, tx, eu); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Finding post ID %s.", r.MessageID)
		if err := entity.FindPost(d.ctx, tx, ep); err != nil {
			return err
		} else if ep.ID == 0 {
			d.logger.Sugar().Debugf("Ignoring untracked post ID %s.", r.MessageID)
			return nil
		}

		d.logger.Sugar().Debugf("Finding emoji %s.", r.Emoji.APIName())
		if err := entity.FindOrCreateEmoji(d.ctx, tx, em); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Finding reaction for emoji %s.", r.Emoji.APIName())
		er.PostID, er.EmojiID = ep.ID, em.ID
		if err := entity.FindOrCreateReaction(d.ctx, tx, er); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating reaction from user %s with emoji %s.", r.UserID, r.Emoji.APIName())
		if err := entity.CreateUserReaction(d.ctx, tx, entity.NewUserReaction(0, er.ID, eu.ID)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		d.logger.Sugar().Errorf("Failed to complete reaction creation transaction: %s.", err)
	} else {
		d.logger.Sugar().Infof("Finished creating reaction for message %s.", r.MessageID)
	}
}

func (d *Discord) maybeRemoveReaction(r *discordgo.MessageReaction) {
	eu := entity.NewUserFromSnowflakeID(r.UserID)
	ep := entity.NewPostFromSnowflakeID(r.MessageID, 0, 0, "")

	em := entity.NewEmojiFromDiscord(&r.Emoji)
	er := entity.NewReaction(0, 0, 0)

	if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		d.logger.Sugar().Debugf("Finding post ID %s.", r.MessageID)
		if err := entity.FindPost(d.ctx, tx, ep); err != nil {
			return err
		} else if ep.ID == 0 {
			d.logger.Sugar().Debugf("Ignoring reaction removal from post %s that is not in the database.", r.MessageID)
			return nil
		}

		d.logger.Sugar().Debugf("Finding user ID %s.", r.UserID)
		if err := entity.FindUser(d.ctx, tx, eu); err != nil {
			return err
		} else if eu.ID == 0 {
			d.logger.Sugar().Debugf("There is no user with ID %s.", r.UserID)
			return nil
		}

		d.logger.Sugar().Debugf("Finding emoji %s.", r.Emoji.APIName())
		if err := entity.FindEmoji(d.ctx, tx, em); err != nil {
			return err
		} else if eu.ID == 0 {
			d.logger.Sugar().Debugf("There is no emoji %s.", r.Emoji.APIName())
			return nil
		}

		d.logger.Sugar().Debugf("Finding reaction for emoji %s.", r.Emoji.APIName())
		er.PostID, er.EmojiID = ep.ID, em.ID
		if err := entity.FindReaction(d.ctx, tx, er); err != nil {
			return err
		} else if eu.ID == 0 {
			d.logger.Sugar().Debugf("There is no reaction for emoji %s.", r.Emoji.APIName())
			return nil
		}

		d.logger.Sugar().Debugf("Deleting reaction from user %s with emoji %s.", r.UserID, r.Emoji.APIName())
		if ok, err := entity.DeleteUserReaction(d.ctx, tx, entity.NewUserReaction(0, er.ID, eu.ID)); err != nil {
			return err
		} else if !ok {
			d.logger.Sugar().Debugf("There is no reaction from user %s with emoji %s.", r.UserID, r.Emoji.APIName())
		}

		return nil
	}); err != nil {
		d.logger.Sugar().Errorf("Failed to complete reaction creation transaction: %s.", err)
	} else {
		d.logger.Sugar().Infof("Finished deleting reaction for message %s.", r.MessageID)
	}
}

func (d *Discord) maybeRemoveAllReactions(r *discordgo.MessageReaction) {
	ep := entity.NewPostFromSnowflakeID(r.MessageID, 0, 0, "")

	if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		d.logger.Sugar().Debugf("Finding post ID %s.", r.MessageID)
		if err := entity.FindPost(d.ctx, tx, ep); err != nil {
			return err
		} else if ep.ID == 0 {
			d.logger.Sugar().Debugf("Ignoring reaction removal from post %s that is not in the database.", r.MessageID)
			return nil
		}

		d.logger.Sugar().Debugf("Finding reaction for emoji %s.", r.Emoji.APIName())
		if ok, err := entity.DeleteAllReactions(d.ctx, tx, ep); err != nil {
			return err
		} else if ok {
			d.logger.Sugar().Infof("Deleted all reactions for post %s.", r.MessageID)
		} /*else {
			d.logger.Sugar().Debugf("There were no reactions for post %s.", r.MessageID)
			return nil
		}*/

		return nil
	}); err != nil {
		d.logger.Sugar().Errorf("Failed to complete reaction deletion transaction: %s.", err)
	} else {
		d.logger.Sugar().Infof("Finished deleting reactions for message %s.", r.MessageID)
	}
}
