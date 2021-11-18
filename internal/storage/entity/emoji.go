package entity

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
)

type Emoji struct {
	NullIdentifiableDiscordEntity
	Name string
}

func NewEmoji(ID ID, discordID NullableSnowflake, name string) *Emoji {
	return &Emoji{NullIdentifiableDiscordEntity{IdentifiableEntity{ID}, discordID}, name}
}

func NewEmojiFromDiscord(em *discordgo.Emoji) *Emoji {
	if em.ID == "" {
		return NewEmoji(0, NullableSnowflake{}, em.Name)
	} else {
		return NewEmoji(0, NullableSnowflake{Int64: int64(MustParseSnowflake(em.ID)), Valid: true}, em.Name)
	}
}

func FindEmoji(ctx context.Context, tx pgx.Tx, em *Emoji) error {
	var sql string
	var args []interface{}

	if em.DiscordID.Valid {
		sql = `select id from emoji where discord_id = $1`
		args = []interface{}{em.DiscordID}
	} else {
		sql = `select id from emoji where name = $1`
		args = []interface{}{em.Name}
	}

	return query(ctx, tx, sql, args, []interface{}{&em.ID})
}

func FindOrCreateEmoji(ctx context.Context, tx pgx.Tx, em *Emoji) error {
	var sql string
	var args []interface{}

	if em.DiscordID.Valid {
		sql = `with e as (insert into emoji (discord_id, name) values ($1, $2) on conflict do nothing returning id) select id from e union select id from emoji where discord_id = $1`
		args = []interface{}{em.DiscordID, em.Name}
	} else {
		sql = `with e as (insert into emoji (name) values ($1) on conflict do nothing returning id) select id from e union select id from emoji where name = $1`
		args = []interface{}{em.Name}
	}

	return query(ctx, tx, sql, args, []interface{}{&em.ID})
}
