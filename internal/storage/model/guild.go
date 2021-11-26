package model

import (
	"pkg.mon.icu/monicu/internal/util"
)

type Guild struct {
	ID        uint       `gorm:"type:int;primaryKey;auto_increment" json:"id"`
	DiscordID uint64     `gorm:"notNull;uniqueIndex" json:"-"`
	Channels  []*Channel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func ForGuildID(DiscordID string) *Guild {
	return &Guild{DiscordID: util.MustParseSnowflake(DiscordID)}
}
