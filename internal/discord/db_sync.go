package discord

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
	"pkg.mon.icu/monicu/internal/storage/entity"
)

func (d *Discord) maybeCreatePost(e *discordgo.MessageCreate) {
	if len(e.Attachments) == 0 && len(e.Embeds) == 0 {
		d.logger.Sugar().Infof("Ignoring attachmentless and embedless message %s.", e.ID)
		return
	}

	d.logger.Sugar().Debugf("Parsing guild ID %s.", e.GuildID)
	guildID, err := strconv.ParseUint(e.GuildID, 10, 64)
	if err != nil {
		d.logger.Sugar().Error("Couldn't parse guild ID %s.", e.GuildID)
		return
	}
	if !d.config.guilds.Contains(guildID) {
		return
	}

	d.logger.Sugar().Debugf("Parsing channel ID %s.", e.ChannelID)
	chanID, err := strconv.ParseUint(e.ChannelID, 10, 64)
	if err != nil {
		d.logger.Sugar().Error("Couldn't parse channel ID %s.", e.ChannelID)
		return
	}
	if !d.config.chans.Contains(chanID) {
		return
	}

	d.logger.Sugar().Debugf("Parsing message ID %s.", e.ID)
	mesID, err := strconv.ParseUint(e.ID, 10, 64)
	if err != nil {
		d.logger.Sugar().Error("Couldn't parse message ID %s.", e.ID)
		return
	}

	d.logger.Sugar().Debugf("Parsing user ID %s.", e.ChannelID)
	userID, err := strconv.ParseUint(e.Author.ID, 10, 64)
	if err != nil {
		d.logger.Sugar().Error("Couldn't parse user ID %s.", e.Author.ID)
		return
	}

	d.logger.Sugar().Debugf("Creating post for message %s.", e.ID)
	if err := d.storage.Begin(d.ctx, func(tx pgx.Tx) error {
		d.logger.Sugar().Debugf("Creating guild ID %s.", e.GuildID)
		guild := entity.NewGuild(0, guildID)
		if err := entity.FindOrCreateGuild(d.ctx, tx, guild); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating channel ID %s.", e.ChannelID)
		chan_ := entity.NewChannel(0, chanID, guild.ID)
		if err := entity.FindOrCreateChannel(d.ctx, tx, chan_); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating user ID %s.", e.Author.ID)
		user := entity.NewUser(0, userID)
		if err := entity.FindOrCreateUser(d.ctx, tx, user); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating post ID %s.", e.ID)
		post := entity.NewPost(0, mesID, chan_.ID, user.ID, e.Content)
		if err := entity.CreatePost(d.ctx, tx, post); err != nil {
			return err
		}

		d.logger.Sugar().Debugf("Creating attachments for post %s.", e.ID)
		for _, at := range e.Attachments {
			if at.Width != 0 && at.Height != 0 {
				im := entity.NewImageFromAttachment(at, post.ID)
				if err := entity.CreateImage(d.ctx, tx, im); err != nil {
					return err
				}
			} else {
				d.logger.Sugar().Debug("Ignoring non-image attachment %s.", at.ID)
			}
		}
		for _, em := range e.Embeds {
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
		d.logger.Sugar().Infof("Finished creating post for message %s.", e.ID)
	}
}
