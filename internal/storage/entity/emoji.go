package entity

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

type Emoji struct {
	NullIdentifiableDiscordEntity
	Name string
}

func NewEmoji(ID ID, discordID NullableSnowflake, name string) *Emoji {
	return &Emoji{NullIdentifiableDiscordEntity{IdentifiableEntity{ID}, discordID}, name}
}

func NewEmojiFromDiscord(em *discordgo.Emoji) (*Emoji, error) {
	if em.ID == "" {
		return NewEmoji(0, NullableSnowflake{}, em.Name), nil
	} else {
		discordID, err := strconv.ParseUint(em.ID, 10, 64)
		if err != nil {
			return nil, err
		}
		return NewEmoji(0, NullableSnowflake{Int64: int64(discordID), Valid: true}, em.Name), nil
	}
}
