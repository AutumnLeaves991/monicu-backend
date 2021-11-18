package entity

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type Post struct {
	IdentifiableDiscordEntity
	ChannelID Ref
	UserID    Ref
	Message   string
}

func NewPost(ID ID, discordID Snowflake, channelID Ref, userID Ref, message string) *Post {
	return &Post{IdentifiableDiscordEntity{IdentifiableEntity{ID}, discordID}, channelID, userID, message}
}

func NewPostFromSnowflakeID(id string, channelID Ref, userID Ref, message string) *Post {
	return NewPost(0, MustParseSnowflake(id), channelID, userID, message)
}

func CreatePost(ctx context.Context, tx pgx.Tx, p *Post) error {
	return query(
		ctx,
		tx,
		`insert into post (discord_id, channel_id, user_id, message) values ($1, $2, $3, $4) returning id`,
		[]interface{}{p.DiscordID, p.ChannelID, p.UserID, p.Message},
		[]interface{}{&p.ID},
	)
}

func FindPost(ctx context.Context, tx pgx.Tx, p *Post) error {
	return query(
		ctx,
		tx,
		`select id, channel_id, user_id, message from post where discord_id = $1`,
		[]interface{}{p.DiscordID},
		[]interface{}{&p.ID, &p.ChannelID, &p.UserID, &p.Message},
	)
}

func DeletePost(ctx context.Context, tx pgx.Tx, p *Post) (bool, error) {
	return queryDeletion(
		ctx,
		tx,
		`delete from post where discord_id = $1`,
		[]interface{}{p.DiscordID},
	)
}

func IsChannelEmpty(ctx context.Context, tx pgx.Tx, c *Channel) (bool, error) {
	var i int
	if err := query(
		ctx,
		tx,
		`select 1 from post where channel_id = $1 limit 1`,
		[]interface{}{c.ID},
		[]interface{}{&i},
	); err != nil {
		return false, err
	} else {
		return i == 0, nil
	}
}
