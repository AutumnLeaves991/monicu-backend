package model

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type Channel struct {
	IdentifiableDiscordEntity
	GuildID Ref
}

func WrapChannelID(ID string) *Channel {
	return &Channel{IdentifiableDiscordEntity{IdentifiableEntity{}, MustParseSnowflake(ID)}, 0}
}

func FindOrCreateChannel(ctx context.Context, tx pgx.Tx, ch *Channel) error {
	return query(ctx, tx, `with e as (insert into channel (discord_id, guild_id) values ($1, $2) on conflict do nothing returning id) select id from e union select id from channel where discord_id = $1 and guild_id = $2`, []interface{}{ch.DiscordID, ch.GuildID}, []interface{}{&ch.ID})
}

func FindChannel(ctx context.Context, tx pgx.Tx, ch *Channel) error {
	return query(ctx, tx, `select id from channel where discord_id = $1`, []interface{}{ch.DiscordID}, []interface{}{&ch.ID})
}
