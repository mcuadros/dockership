package main

import (
	"github.com/mcuadros/dockership/http"
)

var version string
var build string

func main() {
	http.Start(version, build)
}
