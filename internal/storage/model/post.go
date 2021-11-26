package model

import (
	"github.com/bwmarrin/discordgo"
	"pkg.mon.icu/monicu/internal/util"
)

type Post struct {
	ID        uint        `gorm:"type:int;primaryKey;auto_increment" json:"id"`
	DiscordID uint64      `gorm:"notNull;uniqueIndex" json:"-"`
	ChannelID uint        `gorm:"index" json:"channel_id"`
	Channel   *Channel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	UserID    uint        `gorm:"index" json:"user_id"`
	User      *User       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Images    []*Image    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"images"`
	Reactions []*Reaction `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"reactions"`
	Content   string      `json:"content"`
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
