package model

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Image struct {
	ID     uint  `gorm:"type:int;primaryKey;auto_increment"`
	PostID uint  `gorm:"index"`
	Post   *Post `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	URL    string
	Width  uint
	Height uint
	Size   uint64
}

func ForDiscordAttachment(at *discordgo.MessageAttachment) *Image {
	return &Image{
		URL:    at.ProxyURL,
		Width:  uint(at.Width),
		Height: uint(at.Height),
		Size:   uint64(at.Size),
	}
}

func ForDiscordEmbed(ctx context.Context, em *discordgo.MessageEmbed) (*Image, error) {
	size, err := fetchImageSize(ctx, em.Image.ProxyURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch image size: %w", err)
	}

	return &Image{
		URL:    em.Image.ProxyURL,
		Width:  uint(em.Image.Width),
		Height: uint(em.Image.Height),
		Size:   size,
	}, nil
}

func fetchImageSize(ctx context.Context, url string) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	clen, err := strconv.ParseUint(res.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return 0, err
	}

	return clen, nil
}
