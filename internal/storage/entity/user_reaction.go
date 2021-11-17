package entity

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type UserReaction struct {
	IdentifiableEntity
	ReactionID Ref
	UserID     Ref
}

func NewUserReaction(ID ID, reactionID, userID Ref) *UserReaction {
	return &UserReaction{IdentifiableEntity{ID}, reactionID, userID}
}

func CreateUserReaction(ctx context.Context, tx pgx.Tx, ur *UserReaction) error {
	return query(
		ctx,
		tx,
		`insert into user_reaction (reaction_id, user_id) values ($1, $2) returning id`,
		[]interface{}{ur.ReactionID, ur.UserID},
		[]interface{}{&ur.ID},
	)
}

func DeleteUserReaction(ctx context.Context, tx pgx.Tx, ur *UserReaction) (bool, error) {
	return queryDeletion(
		ctx,
		tx,
		`delete from user_reaction where reaction_id = $1 and user_id = $2`,
		[]interface{}{ur.ReactionID, ur.UserID},
	)
}
