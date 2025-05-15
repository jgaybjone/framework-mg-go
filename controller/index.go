package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type IndexController struct {
	// todo service
}

func NewIndexController() Router {
	return &IndexController{}
}

func (i *IndexController) Handler(c *gin.Engine) {
	c.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})
}
