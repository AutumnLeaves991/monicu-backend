package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pkg.mon.icu/monicu/internal/storage/model"
)

// registerGetPostsAll GET /posts/all
func (a *API) registerGetPostsAll() {
	a.router.GET("/posts/all", func(c *gin.Context) {
		var param struct {
			Limit  int `form:"limit" binding:"min=0,max=100"`
			Offset int `form:"offset" binding:"min=0"`
		}

		if err := c.ShouldBindUri(&param); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var posts []*model.Post
		if err := a.storage.Transaction(func(db *gorm.DB) error {
			return db.
				WithContext(a.ctx).
				Limit(param.Limit).
				Offset(param.Offset).
				Order("discord_id desc").
				Preload("Channel.Guild").
				Preload("User").
				Preload("Images").
				Preload("Reactions.Emoji").
				Preload("Reactions.UserReactions").
				Find(&posts).Error
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, posts)
		}
	})
}
