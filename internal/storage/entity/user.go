package entity

import (
	"context"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
)

type User struct {
	IdentifiableDiscordEntity
}

func NewUser(ID ID, discordID Snowflake) *User {
	return &User{IdentifiableDiscordEntity{IdentifiableEntity{ID}, discordID}}
}

func NewUserFromDiscord(u *discordgo.User) (*User, error) {
	discordID, err := strconv.ParseUint(u.ID, 10, 64)
	if err != nil {
		return nil, err
	}
	return NewUser(0, discordID), nil
}

func FindOrCreateUser(ctx context.Context, tx pgx.Tx, u *User) error {
	return Query(
		ctx,
		tx,
		`with e as (insert into "user" (discord_id) values ($1) on conflict do nothing returning id) select id from e union select id from "user" where discord_id = $1`,
		[]interface{}{u.DiscordID},
		[]interface{}{&u.ID},
	)
}