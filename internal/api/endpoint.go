package api

//// registerGetPostsGuildChannel GET /posts/{guildID}/{channelID}
//func (a *API) registerGetPostsGuildChannel() {
//	var param struct {
//		Guild   entity.Ref `uri:"guild" binding:"required"`
//		Channel entity.Ref `uri:"chan" binding:"required"`
//	}
//
//	a.router.GET("/posts/:guild/:chan", func(c *gin.Context) {
//		if err := c.ShouldBindUri(&param); err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//			return
//		}
//
//		if err := a.storage.Begin(a.ctx, func(tx pgx.Tx) error {
//			return nil
//		}); err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//	})
//}
