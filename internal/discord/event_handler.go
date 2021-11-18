package discord

import (
	"github.com/bwmarrin/discordgo"
)

func (d *Discord) onReady(_ *discordgo.Session, e *discordgo.Ready) {
	d.logger.Sugar().Infof("Logged in Discord API as %s.", e.User)
	d.buildChannelGuildCache()
	d.maybeSyncChannels()
}

func (d *Discord) onMessageUpdate(_ *discordgo.Session, e *discordgo.MessageUpdate) {
	d.maybeUpdatePost(e.Message)
}

func (d *Discord) onMessageCreate(_ *discordgo.Session, e *discordgo.MessageCreate) {
	if e.GuildID == "" {
		return
	}
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

func (d *Discord) onMessageReactionAdd(_ *discordgo.Session, e *discordgo.MessageReactionAdd) {
	d.maybeAddReaction(e.MessageReaction)
}

func (d *Discord) onMessageReactionRemove(_ *discordgo.Session, e *discordgo.MessageReactionRemove) {
	d.maybeRemoveReaction(e.MessageReaction)
}

func (d *Discord) onMessageReactionRemoveAll(_ *discordgo.Session, e *discordgo.MessageReactionRemoveAll) {
	d.maybeRemoveAllReactions(e.MessageReaction)
}
