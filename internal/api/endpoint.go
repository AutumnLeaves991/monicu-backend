package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerGetPosts GET /posts/:page
func (a *API) registerGetPosts() {
	a.router.GET("/posts/:page", func(c *gin.Context) {
		var param struct {
			Page uint32 `uri:"page"`
		}

		if err := c.ShouldBindUri(&param); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if posts, err := a.getAllPosts(param.Page); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, posts)
		}
	})
}
