package model

import (
	"context"

	"github.com/bwmarrin/discordgo"
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

func WrapDiscordMessage(m *discordgo.Message) *Post {
	return NewPost(0, MustParseSnowflake(m.ID), 0, 0, m.Content)
}

func WrapMessageID(ID string) *Post {
	return NewPost(0, MustParseSnowflake(ID), 0, 0, "")
}

func CreatePost(ctx context.Context, tx pgx.Tx, p *Post) error {
	return query(ctx, tx, `insert into post (discord_id, channel_id, user_id, message) values ($1, $2, $3, $4) returning id`, []interface{}{p.DiscordID, p.ChannelID, p.UserID, p.Message}, []interface{}{&p.ID})
}

func FindPost(ctx context.Context, tx pgx.Tx, p *Post) error {
	return query(ctx, tx, `select id, channel_id, user_id, message from post where discord_id = $1`, []interface{}{p.DiscordID}, []interface{}{&p.ID, &p.ChannelID, &p.UserID, &p.Message})
}

func FindPosts(ctx context.Context, tx pgx.Tx, offset uint32, limit uint64) ([]*Post, error) {
	p := make([]*Post, 0, limit)
	q, err := tx.Query(ctx, `select id, channel_id, user_id, message from post order by discord_id desc limit $1 offset $2`, limit, offset)
	if err != nil {
		return nil, err
	}

	defer q.Close()
	for q.Next() {
		ep := &Post{}
		if err := q.Scan(&ep.ID, &ep.ChannelID, &ep.UserID, &ep.Message); err != nil {
			return nil, err
		}

		p = append(p, ep)
	}

	return p, nil
}

func UpdatePost(ctx context.Context, tx pgx.Tx, p *Post) (bool, error) {
	return queryUpdateDelete(
		ctx,
		tx,
		`update post set message = $2 where discord_id = $1`,
		[]interface{}{p.DiscordID, p.Message},
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
	if err := query(ctx, tx, `select 1 from post where channel_id = $1 limit 1`, []interface{}{c.ID}, []interface{}{&i}); err != nil {
		return false, err
	} else {
		return i == 0, nil
	}
}
