package model

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type User struct {
	IdentifiableDiscordEntity
}

func NewUser(ID ID, discordID Snowflake) *User {
	return &User{IdentifiableDiscordEntity{IdentifiableEntity{ID}, discordID}}
}

func WrapUserID(ID string) *User {
	return NewUser(0, MustParseSnowflake(ID))
}

func FindOrCreateUser(ctx context.Context, tx pgx.Tx, u *User) error {
	return query(ctx, tx, `with e as (insert into "user" (discord_id) values ($1) on conflict do nothing returning id) select id from e union select id from "user" where discord_id = $1`, []interface{}{u.DiscordID}, []interface{}{&u.ID})
}
