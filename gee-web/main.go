package main

import (
	"geeWeb"
	"net/http"
)

func main() {
	r := geeWeb.Default()
	r.GET("/", func(c *geeWeb.Context) {
		c.String(http.StatusOK, "Hello Geektutu\n")
	})
	// index out of range for testing Recovery()
	r.GET("/panic", func(c *geeWeb.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})

	r.Run(":9999")
}
