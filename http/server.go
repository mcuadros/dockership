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

	router.Run(":8080")
}
