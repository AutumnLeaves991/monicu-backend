package model

import (
	"pkg.mon.icu/monicu/internal/util"
)

type Guild struct {
	ID        uint   `gorm:"type:int;primaryKey;auto_increment"`
	DiscordID uint64 `gorm:"notNull;uniqueIndex"`
	Channels  []*Channel
}

func ForGuildID(DiscordID string) *Guild {
	return &Guild{DiscordID: util.MustParseSnowflake(DiscordID)}
}
