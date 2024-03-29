package model

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type Guild struct {
	IdentifiableDiscordEntity
}

func NewGuild(ID ID, discordID Snowflake) *Guild {
	return &Guild{IdentifiableDiscordEntity{IdentifiableEntity{ID}, discordID}}
}

func WrapGuildID(ID string) *Guild {
	return NewGuild(0, MustParseSnowflake(ID))
}

func FindOrCreateGuild(ctx context.Context, tx pgx.Tx, g *Guild) error {
	return query(ctx, tx, `with e as (insert into guild (discord_id) values ($1) on conflict do nothing returning id) select id from e union select id from guild where discord_id = $1`, []interface{}{g.DiscordID}, []interface{}{&g.ID})
}

func FindGuild(ctx context.Context, tx pgx.Tx, g *Guild) error {
	return query(ctx, tx, `select id from guild where discord_id = $1`, []interface{}{g.DiscordID}, []interface{}{&g.ID})
}
