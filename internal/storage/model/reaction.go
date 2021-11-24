package model

type Reaction struct {
	ID            uint   `gorm:"type:int;primaryKey;auto_increment"`
	PostID        uint   `gorm:"uniqueIndex:idx_reaction_post_id_emoji_id"`
	Post          *Post  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	EmojiID       uint   `gorm:"uniqueIndex:idx_reaction_post_id_emoji_id"`
	Emoji         *Emoji `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserReactions []*UserReaction
}

func NewReaction(PostID uint, EmojiID uint) *Reaction {
	return &Reaction{
		PostID:  PostID,
		EmojiID: EmojiID,
	}
}
