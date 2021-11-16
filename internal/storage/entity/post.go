package entity

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
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

func NewPostFromDiscord(m *discordgo.Message, userID Ref, channelID Ref) (*Post, error) {
	discordID, err := strconv.ParseUint(m.ID, 10, 64)
	if err != nil {
		return nil, err
	}
	return NewPost(0, discordID, channelID, userID, m.Content), nil
}
