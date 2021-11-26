package model

type Reaction struct {
	ID            uint            `gorm:"type:int;primaryKey;auto_increment" json:"-"`
	PostID        uint            `gorm:"uniqueIndex:idx_reaction_post_id_emoji_id" json:"post_id"`
	Post          *Post           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	EmojiID       uint            `gorm:"uniqueIndex:idx_reaction_post_id_emoji_id" json:"emoji_id"`
	Emoji         *Emoji          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	UserReactions []*UserReaction `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"users"`
}

func NewReaction(PostID uint, EmojiID uint) *Reaction {
	return &Reaction{
		PostID:  PostID,
		EmojiID: EmojiID,
	}
}
