package router

import "github.com/gin-gonic/gin"

func ReverseProxy() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	return r
}
