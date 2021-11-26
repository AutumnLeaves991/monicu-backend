package model

import (
	"pkg.mon.icu/monicu/internal/util"
)

type Channel struct {
	ID        uint    `gorm:"type:int;primaryKey;auto_increment" json:"id"`
	DiscordID uint64  `gorm:"notNull;uniqueIndex" json:"-"`
	GuildID   uint    `gorm:"index" json:"guild_id"`
	Guild     *Guild  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Posts     []*Post `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func ForChannelID(DiscordID string) *Channel {
	return &Channel{DiscordID: util.MustParseSnowflake(DiscordID)}
}

func ForChannelIDUint(DiscordID uint64) *Channel {
	return &Channel{DiscordID: DiscordID}
}
