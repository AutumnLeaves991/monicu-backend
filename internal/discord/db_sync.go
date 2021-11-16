package discord

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
	"pkg.mon.icu/monicu/internal/storage/entity"
)

func (d *Discord) maybeCreatePost(m *discordgo.Message) {
	if len(m.Attachments) == 0 && len(m.Embeds) == 0 {
		d.logger.Sugar().Infof("Ignoring attachmentless and embedless message %s.", m.ID)
		return
	}

	d.logger.Sugar().Debugf("Parsing guild ID %s.", m.GuildID)
	guildID, err := strconv.ParseUint(m.GuildID, 10, 64)
	if err != nil {
		d.logger.Sugar().Error("Couldn't parse guild ID %s.", m.GuildID)
		return
	}
	if !d.config.guilds.Contains(guildID) {
		return
	}

	d.logger.Sugar().Debugf("Parsing channel ID %s.", m.ChannelID)
	chanID, err := strconv.ParseUint(m.ChannelID, 10, 64)
	if err != nil {
		d.logger.Sugar().Error("Couldn't parse channel ID %s.", m.ChannelID)
		return
	}
	if !d.config.chans.Contains(chanID) {
		return
	}

	d.logger.Sugar().Debugf("Parsing message ID %s.", m.ID)
	mesID, err := strconv.ParseUint(m.ID, 10, 64)
	if err != nil {
		d.logger.Sugar().Error("Couldn't parse message ID %s.", m.ID)
		return
	}

	d.logger.Sugar().Debugf("Parsing user ID %s.", m.ChannelID)
	userID, err := strconv.ParseUint(m.Author.ID, 10, 64)
	if err != nil {
		d.logger.Sugar().Error("Couldn't parse user ID %s.", m.Author.ID)
		return
	}

	d.logger.Sugar().Debugf("Creating post for message %s.", m.ID)
	if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		d.logger.Sugar().Debugf("Creating guild ID %s.", m.GuildID)
		guild := entity.NewGuild(0, guildID)
		if err := entity.FindOrCreateGuild(d.ctx, tx, guild); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating channel ID %s.", m.ChannelID)
		chan_ := entity.NewChannel(0, chanID, guild.ID)
		if err := entity.FindOrCreateChannel(d.ctx, tx, chan_); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating user ID %s.", m.Author.ID)
		user := entity.NewUser(0, userID)
		if err := entity.FindOrCreateUser(d.ctx, tx, user); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating post ID %s.", m.ID)
		post := entity.NewPost(0, mesID, chan_.ID, user.ID, m.Content)
		if err := entity.CreatePost(d.ctx, tx, post); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating attachments for post %s.", m.ID)
		for _, at := range m.Attachments {
			if at.Width != 0 && at.Height != 0 {
				im := entity.NewImageFromAttachment(at, post.ID)
				if err := entity.CreateImage(d.ctx, tx, im); err != nil {
					return err
				}
			} else {
				d.logger.Sugar().Debug("Ignoring non-image attachment %s.", at.ID)
			}
		}
		for _, em := range m.Embeds {
			if em.Image != nil {
				im, err := entity.NewImageFromEmbed(d.ctx, em, post.ID)
				if err != nil {
					return err
				}
				if err := entity.CreateImage(d.ctx, tx, im); err != nil {
					return err
				}
			} else {
				d.logger.Sugar().Debug("Ignoring non-image embed.")
			}
		}

		// todo: reactions

		return nil
	}); err != nil {
		d.logger.Sugar().Errorf("Failed to complete post creation transaction: %s.", err)
	} else {
		d.logger.Sugar().Infof("Finished creating post for message %s.", m.ID)
	}
}

func (d *Discord) maybeDeletePost(m *discordgo.Message) {
	d.logger.Sugar().Debugf("Parsing message ID %s.", m.ID)
	mesID, err := strconv.ParseUint(m.ID, 10, 64)
	if err != nil {
		d.logger.Sugar().Error("Couldn't parse message ID %s.", m.GuildID)
		return
	}

	if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		if ok, err := entity.DeletePost(d.ctx, tx, entity.NewPost(0, mesID, 0, 0, "")); err != nil {
			return err
		} else if ok {
			d.logger.Sugar().Debugf("Deleted post %d.", mesID)
		}

		return nil
	}); err != nil {
		d.logger.Sugar().Errorf("Failed to delete post %d: %s", mesID, err)
	}
}
