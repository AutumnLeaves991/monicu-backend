package model

import (
	"pkg.mon.icu/monicu/internal/util"
)

type User struct {
	ID        uint    `gorm:"type:int;primaryKey;auto_increment" json:"id"`
	DiscordID uint64  `gorm:"notNull;uniqueIndex" json:"-"`
	Posts     []*Post `json:"-"`
}

func ForUserID(DiscordID string) *User {
	return &User{DiscordID: util.MustParseSnowflake(DiscordID)}
}
