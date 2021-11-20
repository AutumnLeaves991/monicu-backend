package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"pkg.mon.icu/monicu/internal/storage/entity"
)

// registerGetPosts GET /posts/:page
func (a *API) registerGetPosts() {
	var param struct {
		Page uint32 `uri:"page"`
	}

	type imageModel struct {
		URL    string `json:"url"`
		Width  uint32 `json:"width"`
		Height uint32 `json:"height"`
		Size   uint64 `json:"size"`
	}

	type postModel struct {
		ID        entity.Ref    `json:"id"`
		ChannelID entity.Ref    `json:"channel"`
		UserID    entity.Ref    `json:"user"`
		//Message   string        `json:"message"`
		Images    []*imageModel `json:"images"`
		Reactions uint32        `json:"reactions"`
	}

	a.router.GET("/posts/:page", func(c *gin.Context) {
		if err := c.ShouldBindUri(&param); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := a.storage.Begin(a.ctx, func(tx pgx.Tx) error {
			var posts []*entity.Post
			var err error
			if posts, err = entity.FindPosts(a.ctx, tx, param.Page * 100, 100); err != nil {
				return err
			}

			pm := make([]*postModel, len(posts))
			for i, p := range posts {
				var images []*entity.Image
				if images, err = entity.FindImages(a.ctx, tx, p); err != nil {
					return err
				}

				imm := make([]*imageModel, len(images))
				for i, im := range images {
					imm[i] = &imageModel{im.URL, im.Width, im.Height, im.Size}
				}

				var rc uint32
				if rc, err = entity.CountUserReactions(a.ctx, tx, p); err != nil {
					return err
				}

				pm[i] = &postModel{p.ID, p.ChannelID, p.UserID, /*p.Message,*/ imm, rc}
			}

			c.JSON(http.StatusOK, pm)
			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})
}
