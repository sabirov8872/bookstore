package main

import (
	"github.com/sabirov8872/bookstore/cmd/app"
	_ "github.com/sabirov8872/bookstore/docs"
)

// @title   Bookstore
// @version  1.0

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	app.Run()
}
