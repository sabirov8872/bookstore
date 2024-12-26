package app

import (
	"fmt"
	"log"

	"github.com/sabirov8872/bookstore/config"
	"github.com/sabirov8872/bookstore/internal/handler"
	"github.com/sabirov8872/bookstore/internal/repository"
	"github.com/sabirov8872/bookstore/internal/routes"
	"github.com/sabirov8872/bookstore/internal/service"
	"github.com/sabirov8872/bookstore/pkg/minio"
	"github.com/sabirov8872/bookstore/pkg/postgres"
	"github.com/sabirov8872/bookstore/pkg/redis"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.Get(cfg.Postgres)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mc, err := minio.NewClient(cfg.Minio)
	if err != nil {
		log.Fatal(err)
	}

	rc, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("START")

	repo := repository.NewRepository(db)
	serv := service.NewService(repo, rc, mc)
	hand := handler.NewHandler(serv)
	routes.Run(hand, cfg.Server.Port, repo)
}
