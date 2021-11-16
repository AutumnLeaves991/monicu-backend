package discord

import (
	"github.com/bwmarrin/discordgo"
)

func (d *Discord) onReady(_ *discordgo.Session, e *discordgo.Ready) {
	d.logger.Sugar().Infof("Logged in Discord API as %s.", e.User)
}

func (d *Discord) onMessageCreate(_ *discordgo.Session, e *discordgo.MessageCreate) {
	d.maybeCreatePost(e.Message)
}

func (d *Discord) onMessageDelete(_ *discordgo.Session, e *discordgo.MessageDelete) {
	d.maybeDeletePost(e.Message)
}
