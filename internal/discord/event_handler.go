package discord

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func (d *Discord) onReady(_ *discordgo.Session, e *discordgo.Ready) {
	d.logger.Sugar().Infof("Logged in Discord API as %s.", e.User)
}

func (d *Discord) onMessageCreate(_ *discordgo.Session, e *discordgo.MessageCreate) {
	guildID, err := strconv.ParseUint(e.GuildID, 10, 64)
	if err != nil {
		d.logger.Sugar().Errorf("Failed to parse guild ID: %s.", err)
		return
	}
	chanID, err := strconv.ParseUint(e.ChannelID, 10, 64)
	if err != nil {
		d.logger.Sugar().Errorf("Failed to parse channel ID: %s.", err)
		return
	}
	if !d.config.guilds.Contains(guildID) || !d.config.chans.Contains(chanID) {
		return
	}
}
