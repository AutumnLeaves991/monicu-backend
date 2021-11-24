package model

import (
	"pkg.mon.icu/monicu/internal/util"
)

type User struct {
	ID        uint   `gorm:"type:int;primaryKey;auto_increment"`
	DiscordID uint64 `gorm:"notNull;uniqueIndex"`
	Posts     []*Post
}

func ForUserID(DiscordID string) *User {
	return &User{DiscordID: util.MustParseSnowflake(DiscordID)}
}
