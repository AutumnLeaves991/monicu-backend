package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"pkg.mon.icu/monicu/internal/storage/model"
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
		ID        model.Ref `json:"id"`
		ChannelID model.Ref `json:"channel"`
		UserID    model.Ref `json:"user"`
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
			var posts []*model.Post
			var err error
			if posts, err = model.FindPosts(a.ctx, tx, param.Page * 100, 100); err != nil {
				return err
			}

			pm := make([]*postModel, len(posts))
			for i, p := range posts {
				var images []*model.Image
				if images, err = model.FindImages(a.ctx, tx, p); err != nil {
					return err
				}

				imm := make([]*imageModel, len(images))
				for i, im := range images {
					imm[i] = &imageModel{im.URL, im.Width, im.Height, im.Size}
				}

				var rc uint32
				if rc, err = model.CountUserReactions(a.ctx, tx, p); err != nil {
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
