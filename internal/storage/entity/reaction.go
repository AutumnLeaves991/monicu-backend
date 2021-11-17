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
	return query(
		ctx,
		tx,
		`insert into reaction (post_id, emoji_id) values ($1, $2) returning id`,
		[]interface{}{r.PostID, r.EmojiID},
		[]interface{}{&r.ID},
	)
}