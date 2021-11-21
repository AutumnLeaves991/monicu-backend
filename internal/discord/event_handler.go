package discord

import (
	"fmt"
	"reflect"

	"github.com/bwmarrin/discordgo"
	"pkg.mon.icu/monicu/internal/storage/model"
)

// Util functions

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

func (d *Discord) shouldIgnoreEvent(e interface{}) bool {
	var gID, cID model.Snowflake
	switch e := e.(type) {
	case *discordgo.MessageUpdate:
		gID, cID = model.MustParseSnowflake(e.GuildID), model.MustParseSnowflake(e.ChannelID)
	case *discordgo.MessageCreate:
		gID, cID = model.MustParseSnowflake(e.GuildID), model.MustParseSnowflake(e.ChannelID)
	case *discordgo.MessageDelete:
		gID, cID = model.MustParseSnowflake(e.GuildID), model.MustParseSnowflake(e.ChannelID)
	case *discordgo.MessageDeleteBulk:
		gID, cID = model.MustParseSnowflake(e.GuildID), model.MustParseSnowflake(e.ChannelID)
	case *discordgo.MessageReactionAdd:
		gID, cID = model.MustParseSnowflake(e.GuildID), model.MustParseSnowflake(e.ChannelID)
	case *discordgo.MessageReactionRemove:
		gID, cID = model.MustParseSnowflake(e.GuildID), model.MustParseSnowflake(e.ChannelID)
	case *discordgo.MessageReactionRemoveAll:
		gID, cID = model.MustParseSnowflake(e.GuildID), model.MustParseSnowflake(e.ChannelID)
	default:
		panic(fmt.Errorf("unknown event type %s", reflect.TypeOf(e).Name()))
	}

	if !d.config.guilds.Contains(gID) {
		return true
	}

	if !isGuildEvent(e) {
		_, ok := e.(*discordgo.MessageUpdate)
		return ok
	}

	if !d.config.chans.Contains(cID) {
		return true
	}

	return false
}

// Event handlers

func (d *Discord) onReady(_ *discordgo.Session, _ *discordgo.Ready) {
	d.buildChannelGuildCache()
	d.createChannelsAndGuilds()
	d.syncChannels()
}

func (d *Discord) onMessageCreate(_ *discordgo.Session, e *discordgo.MessageCreate) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	d.createPost(e.Message)
}

func (d *Discord) onMessageUpdate(_ *discordgo.Session, e *discordgo.MessageUpdate) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	d.updatePost(e.Message)
}

func (d *Discord) onMessageDelete(_ *discordgo.Session, e *discordgo.MessageDelete) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	d.deletePost(e.Message)
}

func (d *Discord) onMessageDeleteBulk(_ *discordgo.Session, e *discordgo.MessageDeleteBulk) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	d.deletePostsBulk(e.Messages)
}

func (d *Discord) onMessageReactionAdd(_ *discordgo.Session, e *discordgo.MessageReactionAdd) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	d.addReaction(e.MessageReaction)
}

func (d *Discord) onMessageReactionRemove(_ *discordgo.Session, e *discordgo.MessageReactionRemove) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	d.removeReaction(e.MessageReaction)
}

func (d *Discord) onMessageReactionRemoveAll(_ *discordgo.Session, e *discordgo.MessageReactionRemoveAll) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	d.removeReactionsBulk(e.MessageReaction)
}
