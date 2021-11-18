package entity

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type Channel struct {
	IdentifiableDiscordEntity
	GuildID Ref
}

func NewChannel(ID ID, discordID Snowflake, guildID Ref) *Channel {
	return &Channel{IdentifiableDiscordEntity{IdentifiableEntity{ID}, discordID}, guildID}
}

func NewChannelFromSnowflakeID(id string, guildID Ref) *Channel {
	return NewChannel(0, mustParseSnowflake(id), guildID)
}

func FindOrCreateChannel(ctx context.Context, tx pgx.Tx, ch *Channel) error {
	return query(
		ctx,
		tx,
		`with e as (insert into channel (discord_id, guild_id) values ($1, $2) on conflict do nothing returning id) select id from e union select id from channel where discord_id = $1 and guild_id = $2`,
		[]interface{}{ch.DiscordID, ch.GuildID},
		[]interface{}{&ch.ID},
	)
}

func FindChannel(ctx context.Context, tx pgx.Tx, ch *Channel) error {
	return query(
		ctx,
		tx,
		`select id from channel where discord_id = $1`,
		[]interface{}{ch.DiscordID},
		[]interface{}{&ch.ID},
	)
}
