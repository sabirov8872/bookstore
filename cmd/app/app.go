package app

import (
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

	repo := repository.NewRepository(db)
	serv := service.NewService(repo, rc, mc, cfg.Secret.Key)
	hand := handler.NewHandler(serv, cfg.Secret.Key)
	routes.Run(hand, cfg.Server.Port, cfg.Secret.Key)
}
