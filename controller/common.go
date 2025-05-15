package controller

import "github.com/gin-gonic/gin"

type Router interface {
	Handler(c *gin.Engine)
}
