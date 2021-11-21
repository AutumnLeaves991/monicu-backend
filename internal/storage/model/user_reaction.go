package model

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type UserReaction struct {
	IdentifiableEntity
	ReactionID Ref
	UserID     Ref
}

func NewUserReaction() *UserReaction {
	return &UserReaction{}
}

func CreateUserReaction(ctx context.Context, tx pgx.Tx, ur *UserReaction) error {
	return query(ctx, tx, `insert into user_reaction (reaction_id, user_id) values ($1, $2) returning id`, []interface{}{ur.ReactionID, ur.UserID}, []interface{}{&ur.ID})
}

func DeleteUserReaction(ctx context.Context, tx pgx.Tx, ur *UserReaction) (bool, error) {
	return queryUpdateDelete(
		ctx,
		tx,
		`delete from user_reaction where reaction_id = $1 and user_id = $2`,
		[]interface{}{ur.ReactionID, ur.UserID},
	)
}

func CountUserReactions(ctx context.Context, tx pgx.Tx, p *Post) (uint32, error) {
	var count uint32
	if err := query(ctx, tx, `select count(distinct ur.user_id) from post p join reaction r on p.id = r.post_id join user_reaction ur on r.id = ur.reaction_id where p.id = $1`, []interface{}{p.ID}, []interface{}{&count}); err != nil {
		return 0, err
	}

	return count, nil
}