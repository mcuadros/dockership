package main

import (
	"github.com/gin-gonic/gin"
)

var router = gin.Default()

func main() {
	router.Run(":8080")
}
