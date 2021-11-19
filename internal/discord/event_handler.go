package discord

import (
	"fmt"
	"reflect"

	"github.com/bwmarrin/discordgo"
)

func isGuildEvent(e interface{}) bool {
	switch e := e.(type) {
	case *discordgo.MessageUpdate:
		return e.GuildID != ""
	case *discordgo.MessageCreate:
		return e.GuildID != ""
	case *discordgo.MessageDelete:
		return e.GuildID != ""
	case *discordgo.MessageDeleteBulk:
		return e.GuildID != ""
	case *discordgo.MessageReactionAdd:
		return e.GuildID != ""
	case *discordgo.MessageReactionRemove:
		return e.GuildID != ""
	case *discordgo.MessageReactionRemoveAll:
		return e.GuildID != ""
	default:
		panic(fmt.Errorf("unknown event type %s", reflect.TypeOf(e).Name()))
	}
}

func (d *Discord) onReady(_ *discordgo.Session, e *discordgo.Ready) {
	d.logger.Infof("Logged in Discord API as %s.", e.User)
	d.buildChannelGuildCache()
	d.maybeSyncChannels()
}

func (d *Discord) onMessageUpdate(_ *discordgo.Session, e *discordgo.MessageUpdate) {
	if !isGuildEvent(e) {
		return
	}
	d.maybeUpdatePost(e.Message)
}

func (d *Discord) onMessageCreate(_ *discordgo.Session, e *discordgo.MessageCreate) {
	if !isGuildEvent(e) {
		return
	}
	d.maybeCreatePost(e.Message)
}

func (d *Discord) onMessageDelete(_ *discordgo.Session, e *discordgo.MessageDelete) {
	if !isGuildEvent(e) {
		return
	}
	d.maybeDeletePost(e.Message)
}

func (d *Discord) onMessageDeleteBulk(_ *discordgo.Session, e *discordgo.MessageDeleteBulk) {
	if !isGuildEvent(e) {
		return
	}
	d.logger.Debugf("Bulk-deleting posts in channel %s of guild %s.", e.ChannelID, e.GuildID)
	for _, m := range e.Messages {
		d.maybeDeletePost(&discordgo.Message{ID: m})
	}
}

func (d *Discord) onMessageReactionAdd(_ *discordgo.Session, e *discordgo.MessageReactionAdd) {
	if !isGuildEvent(e) {
		return
	}
	d.maybeAddReaction(e.MessageReaction)
}

func (d *Discord) onMessageReactionRemove(_ *discordgo.Session, e *discordgo.MessageReactionRemove) {
	if !isGuildEvent(e) {
		return
	}
	d.maybeRemoveReaction(e.MessageReaction)
}

func (d *Discord) onMessageReactionRemoveAll(_ *discordgo.Session, e *discordgo.MessageReactionRemoveAll) {
	if !isGuildEvent(e) {
		return
	}
	d.maybeRemoveAllReactions(e.MessageReaction)
}
