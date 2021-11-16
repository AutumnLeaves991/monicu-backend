package entity

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Image struct {
	IdentifiableEntity
	PostID Ref
	URL    string
	Width  uint32
	Height uint32
	Size   uint64
}

func NewImage(ID ID, postID Ref, url string, width uint32, height uint32, size uint64) *Image {
	return &Image{IdentifiableEntity{ID}, postID, url, width, height, size}
}

func NewImageFromAttachment(at *discordgo.MessageAttachment, postID Ref) *Image {
	return NewImage(0, postID, at.ProxyURL, uint32(at.Width), uint32(at.Height), uint64(at.Size))
}

func newImageFromEmbed(ctx context.Context, em *discordgo.MessageEmbed, postID Ref) (*Image, error) {
	if em.Image == nil {
		return nil, errors.New("embed has no images")
	}

	size, err := fetchImageSize(ctx, em.Image.ProxyURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch image size: %w", err)
	}

	return NewImage(0, postID, em.Image.ProxyURL, uint32(em.Image.Width), uint32(em.Image.Height), size), nil
}

func fetchImageSize(ctx context.Context, url string) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
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