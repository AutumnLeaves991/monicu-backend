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
	return query(ctx, tx, `insert into post (discord_id, channel_id, user_id, message) values ($1, $2, $3, $4) returning id`, []interface{}{p.DiscordID, p.ChannelID, p.UserID, p.Message}, []interface{}{&p.ID}, func(row pgx.QueryFuncRow) error { return nil })
}

func FindPost(ctx context.Context, tx pgx.Tx, p *Post) error {
	return query(ctx, tx, `select id, channel_id, user_id, message from post where discord_id = $1`, []interface{}{p.DiscordID}, []interface{}{&p.ID, &p.ChannelID, &p.UserID, &p.Message}, func(row pgx.QueryFuncRow) error { return nil })
}

func UpdatePost(ctx context.Context, tx pgx.Tx, p *Post) (bool,error) {
	return queryUpdateDelete(
		ctx,
		tx,
		`update post set (channel_id, user_id, message) = ($2, $3, $4) where discord_id = $1`,
		[]interface{}{p.DiscordID, p.ChannelID, p.UserID, p.Message},
	)
}


func DeletePost(ctx context.Context, tx pgx.Tx, p *Post) (bool, error) {
	return queryUpdateDelete(
		ctx,
		tx,
		`delete from post where discord_id = $1`,
		[]interface{}{p.DiscordID},
	)
}

func DeletePostImages(ctx context.Context, tx pgx.Tx, p *Post) (bool, error) {
	return queryUpdateDelete(ctx, tx, `delete from image where post_id = $1`, []interface{}{p.ID})
}

func IsChannelEmpty(ctx context.Context, tx pgx.Tx, c *Channel) (bool, error) {
	var i int
	if err := query(ctx, tx, `select 1 from post where channel_id = $1 limit 1`, []interface{}{c.ID}, []interface{}{&i}, func(row pgx.QueryFuncRow) error { return nil }); err != nil {
		return false, err
	} else {
		return i == 0, nil
	}
}
