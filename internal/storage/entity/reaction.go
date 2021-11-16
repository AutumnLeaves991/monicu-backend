package entity

type Reaction struct {
	IdentifiableEntity
	PostID  Ref
	EmojiID Ref
}
