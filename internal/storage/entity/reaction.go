package entity

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type Reaction struct {
	IdentifiableEntity
	PostID  Ref
	EmojiID Ref
}

func NewReaction(ID ID, postID, emojiID Ref) *Reaction {
	return &Reaction{IdentifiableEntity{ID}, postID, emojiID}
}

func CreateReaction(ctx context.Context, tx pgx.Tx, r *Reaction) error {
	return query(ctx, tx, `insert into reaction (post_id, emoji_id) values ($1, $2) returning id`, []interface{}{r.PostID, r.EmojiID}, []interface{}{&r.ID}, func(row pgx.QueryFuncRow) error { return nil })
}

func FindReaction(ctx context.Context, tx pgx.Tx, r *Reaction) error {
	return query(ctx, tx, `select id from reaction where post_id = $1 and emoji_id = $2`, []interface{}{r.PostID, r.EmojiID}, []interface{}{&r.ID}, func(row pgx.QueryFuncRow) error { return nil })
}

func DeleteAllReactions(ctx context.Context, tx pgx.Tx, p *Post) (bool, error) {
	return queryUpdateDelete(
		ctx,
		tx,
		`delete from reaction where post_id = $1`,
		[]interface{}{p.ID},
	)
}

func FindOrCreateReaction(ctx context.Context, tx pgx.Tx, r *Reaction) error {
	return query(ctx, tx, `with e as (insert into reaction (post_id, emoji_id) values ($1, $2) on conflict do nothing returning id) select id from e union select id from reaction where post_id = $1 and emoji_id = $2`, []interface{}{r.PostID, r.EmojiID}, []interface{}{&r.ID}, func(row pgx.QueryFuncRow) error { return nil })
}
