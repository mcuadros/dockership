package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var router = gin.Default()

func main() {
	router.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "static/index.html")
	})

	router.GET("/app.js", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "static/app.js")
	})

	router.Run(":8080")
}
