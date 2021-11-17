package discord

import (
	"github.com/bwmarrin/discordgo"
)

func (d *Discord) onReady(_ *discordgo.Session, e *discordgo.Ready) {
	d.logger.Sugar().Infof("Logged in Discord API as %s.", e.User)
}

func (d *Discord) onMessageUpdate(_ *discordgo.Session, e *discordgo.MessageUpdate) {
	if ee, exists := d.embedEditSched[e.ID]; exists {
		if len(e.Embeds) != len(ee.message.Embeds) {
			d.logger.Sugar().Debugf("Received embed addition update event for message %s.", ee.message.ID)
			ee.message.Embeds = e.Embeds
			d.maybeCreatePost(e.Message)
		}

		ee.timer.Stop()
		ee.stopChan <- struct{}{}
	}
}

func (d *Discord) onMessageCreate(_ *discordgo.Session, e *discordgo.MessageCreate) {
	d.maybeCreatePost(e.Message)
}

func (d *Discord) onMessageDelete(_ *discordgo.Session, e *discordgo.MessageDelete) {
	d.maybeDeletePost(e.Message)
}

func (d *Discord) onMessageDeleteBulk(_ *discordgo.Session, e *discordgo.MessageDeleteBulk) {
	d.logger.Sugar().Debugf("Bulk-deleting posts in channel %s of guild %s.", e.ChannelID, e.GuildID)
	for _, m := range e.Messages {
		d.maybeDeletePost(&discordgo.Message{ID: m})
	}
}
