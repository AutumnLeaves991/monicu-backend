package discord

import (
	"fmt"
	"reflect"

	"github.com/bwmarrin/discordgo"
	"pkg.mon.icu/monicu/internal/storage/model"
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

func (d *Discord) onReady(_ *discordgo.Session, e *discordgo.Ready) {
	d.logger.Infof("Logged in Discord API as %s.", e.User)
	d.buildChannelGuildCache()
	d.createChannelsAndGuilds()
	d.syncChannels()
}

func (d *Discord) onMessageCreate(_ *discordgo.Session, e *discordgo.MessageCreate) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	if err := d.createPost(e.Message); err != nil {
		d.logger.Errorf("Failed to create post: %s.", err)
	}
}

func (d *Discord) onMessageUpdate(_ *discordgo.Session, e *discordgo.MessageUpdate) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	if err := d.updatePost(e.Message); err != nil {
		d.logger.Errorf("Failed to update post: %s.", err)
	}
}

func (d *Discord) onMessageDelete(_ *discordgo.Session, e *discordgo.MessageDelete) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	if err := d.deletePost(e.Message); err != nil {
		d.logger.Errorf("Failed to update post: %s.", err)
	}
}

func (d *Discord) onMessageDeleteBulk(_ *discordgo.Session, e *discordgo.MessageDeleteBulk) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	d.logger.Debugf("Bulk-deleting posts in channel %s of guild %s.", e.ChannelID, e.GuildID)
	for _, m := range e.Messages {
		if err := d.deletePost(&discordgo.Message{ID: m}); err != nil {
			d.logger.Errorf("Failed to delete post: %s.", err)
		}
	}
}

func (d *Discord) onMessageReactionAdd(_ *discordgo.Session, e *discordgo.MessageReactionAdd) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	if err := d.addReaction(e.MessageReaction); err != nil {
		d.logger.Errorf("Failed to add reaction: %s.", err)
	}
}

func (d *Discord) onMessageReactionRemove(_ *discordgo.Session, e *discordgo.MessageReactionRemove) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	if err := d.removeReaction(e.MessageReaction); err != nil {
		d.logger.Errorf("Failed to remove reaction: %s.", err)
	}
}

func (d *Discord) onMessageReactionRemoveAll(_ *discordgo.Session, e *discordgo.MessageReactionRemoveAll) {
	if d.shouldIgnoreEvent(e) {
		return
	}
	if err := d.removeReactionsBulk(e.MessageReaction); err != nil {
		d.logger.Errorf("Failed to bulk remove reactions: %s.", err)
	}
}
