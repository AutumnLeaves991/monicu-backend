package model

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"pkg.mon.icu/monicu/internal/util"
)

type Emoji struct {
	ID        uint          `gorm:"type:int;primaryKey;auto_increment" json:"id"`
	DiscordID sql.NullInt64 `gorm:"uniqueIndex" json:"-"`
	Name      string        `gorm:"size:32;uniqueIndex:,where:discord_id is null" json:"name"`
}

func (e *Emoji) IsGuild() bool {
	return e.DiscordID.Valid
}

func ForEmoji(em *discordgo.Emoji) *Emoji {
	if em.ID == "" {
		return &Emoji{Name: em.Name}
	} else {
		return &Emoji{DiscordID: sql.NullInt64{Valid: true, Int64: util.MustParseSnowflakeInt64(em.ID)}, Name: em.Name}
	}
}
