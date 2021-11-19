package discord

import (
	"github.com/bwmarrin/discordgo"
)

func isGuildMessage(m *discordgo.Message) bool {
	return m.GuildID != ""
}

func isGuildReaction(r *discordgo.MessageReaction) bool {
	return r.GuildID != ""
}

func (d *Discord) onReady(_ *discordgo.Session, e *discordgo.Ready) {
	d.logger.Infof("Logged in Discord API as %s.", e.User)
	d.buildChannelGuildCache()
	d.maybeSyncChannels()
}

func (d *Discord) onMessageUpdate(_ *discordgo.Session, e *discordgo.MessageUpdate) {
	if !isGuildMessage(e.Message) {
		return
	}

	d.maybeUpdatePost(e.Message)
}

func (d *Discord) onMessageCreate(_ *discordgo.Session, e *discordgo.MessageCreate) {
	if !isGuildMessage(e.Message) {
		return
	}

	d.maybeCreatePost(e.Message)
}

func (d *Discord) onMessageDelete(_ *discordgo.Session, e *discordgo.MessageDelete) {
	if !isGuildMessage(e.Message) {
		return
	}

	d.maybeDeletePost(e.Message)
}

func (d *Discord) onMessageDeleteBulk(_ *discordgo.Session, e *discordgo.MessageDeleteBulk) {
	d.logger.Debugf("Bulk-deleting posts in channel %s of guild %s.", e.ChannelID, e.GuildID)
	for _, m := range e.Messages {
		if e.GuildID == "" {
			continue
		}

		d.maybeDeletePost(&discordgo.Message{ID: m})
	}
}

func (d *Discord) onMessageReactionAdd(_ *discordgo.Session, e *discordgo.MessageReactionAdd) {
	if !isGuildReaction(e.MessageReaction) {
		return
	}

	d.maybeAddReaction(e.MessageReaction)
}

func (d *Discord) onMessageReactionRemove(_ *discordgo.Session, e *discordgo.MessageReactionRemove) {
	if !isGuildReaction(e.MessageReaction) {
		return
	}

	d.maybeRemoveReaction(e.MessageReaction)
}

func (d *Discord) onMessageReactionRemoveAll(_ *discordgo.Session, e *discordgo.MessageReactionRemoveAll) {
	if !isGuildReaction(e.MessageReaction) {
		return
	}

	d.maybeRemoveAllReactions(e.MessageReaction)
}
