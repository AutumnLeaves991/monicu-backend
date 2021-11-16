package entity

type UserReaction struct {
	IdentifiableEntity
	ReactionID Ref
	UserID     Ref
}
