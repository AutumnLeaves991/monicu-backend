package model

type UserReaction struct {
	ID         uint      `gorm:"type:int;primaryKey;auto_increment" json:"-"`
	ReactionID uint      `gorm:"uniqueIndex:idx_user_reaction_reaction_id_user_id" json:"-"`
	Reaction   *Reaction `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	UserID     uint      `gorm:"uniqueIndex:idx_user_reaction_reaction_id_user_id" json:"user_id"`
	User       *User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func NewUserReaction(ReactionID uint, UserID uint) *UserReaction {
	return &UserReaction{
		ReactionID: ReactionID,
		UserID:     UserID,
	}
}
