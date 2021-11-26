package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pkg.mon.icu/monicu/internal/storage/model"
)

// registerGetPostsReactions GET /posts/reactions
func (a *API) registerGetPostsReactions() {
	a.router.GET("/posts/reactions", func(c *gin.Context) {
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
			// select p.*, count(distinct ur.user_id) as reactions
			// from post p
			//          join reaction r on p.id = r.post_id
			//          join user_reaction ur on r.id = ur.reaction_id
			// group by p.id
			// order by reactions desc
			return db.
				WithContext(a.ctx).
				Limit(param.Limit).
				Offset(param.Offset).
				Joins("join reactions on posts.id = reactions.post_id").
				Joins("join user_reactions on reactions.id = user_reactions.reaction_id").
				//Joins("Reaction").
				//Joins("UserReaction").
				Order("count(distinct user_reactions.id) desc").
				Group("posts.id").
				Preload("Channel.Guild").
				Preload("User").
				Preload("Images").
				Preload("Reactions.Emoji").
				Preload("Reactions.UserReactions").
				Find(&posts).Error
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, posts)
		}
	})
}
