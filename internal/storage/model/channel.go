package model

import (
	"pkg.mon.icu/monicu/internal/util"
)

type Channel struct {
	ID        uint   `gorm:"type:int;primaryKey;auto_increment"`
	DiscordID uint64 `gorm:"notNull;uniqueIndex"`
	GuildID   uint   `gorm:"index"`
	Guild     *Guild `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Posts     []*Post
}

func ForChannelID(DiscordID string) *Channel {
	return &Channel{DiscordID: util.MustParseSnowflake(DiscordID)}
}

func ForChannelIDUint(DiscordID uint64) *Channel {
	return &Channel{DiscordID: DiscordID}
}
