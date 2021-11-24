package model

import (
	"github.com/bwmarrin/discordgo"
	"pkg.mon.icu/monicu/internal/util"
)

type Post struct {
	ID        uint     `gorm:"type:int;primaryKey;auto_increment"`
	DiscordID uint64   `gorm:"notNull;uniqueIndex"`
	ChannelID uint     `gorm:"index"`
	Channel   *Channel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID    uint     `gorm:"index"`
	User      *User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Images    []*Image
	Reactions []*Reaction
	Content   string
}

func ForMessage(message *discordgo.Message) *Post {
	return &Post{
		DiscordID: util.MustParseSnowflake(message.ID),
		Content:   message.Content,
	}
}

func ForMessageID(DiscordID string) *Post {
	return &Post{
		DiscordID: util.MustParseSnowflake(DiscordID),
	}
}
