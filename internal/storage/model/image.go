package model

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
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

func WrapDiscordAttachment(at *discordgo.MessageAttachment) *Image {
	return NewImage(0, 0, at.ProxyURL, uint32(at.Width), uint32(at.Height), uint64(at.Size))
}

func WrapDiscordEmbed(ctx context.Context, em *discordgo.MessageEmbed) (*Image, error) {
	size, err := fetchImageSize(ctx, em.Image.ProxyURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch image size: %w", err)
	}

	return NewImage(0, 0, em.Image.ProxyURL, uint32(em.Image.Width), uint32(em.Image.Height), size), nil
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

func CreateImage(ctx context.Context, tx pgx.Tx, im *Image) error {
	return query(ctx, tx, `insert into image (post_id, url, width, height, size) values ($1, $2, $3, $4, $5) returning id`, []interface{}{im.PostID, im.URL, im.Width, im.Height, im.Size}, []interface{}{&im.ID})
}

func FindImages(ctx context.Context, tx pgx.Tx, p *Post) ([]*Image, error) {
	images := make([]*Image, 0, 4)
	q, err := tx.Query(ctx, `select url, width, height, size from image where post_id = $1`, p.ID)
	if err != nil {
		return nil, err
	}

	defer q.Close()
	for q.Next() {
		im := &Image{}
		if err := q.Scan(&im.URL, &im.Width, &im.Height, &im.Size); err != nil {
			return nil, err
		}

		images = append(images, im)
	}

	return images, nil
}