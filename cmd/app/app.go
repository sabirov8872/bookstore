package app

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/sabirov8872/bookstore/internal/cache"
	"github.com/sabirov8872/bookstore/internal/config"
	"github.com/sabirov8872/bookstore/internal/handler"
	"github.com/sabirov8872/bookstore/internal/minio_client"
	"github.com/sabirov8872/bookstore/internal/repository"
	"github.com/sabirov8872/bookstore/internal/routes"
	"github.com/sabirov8872/bookstore/internal/service"
)

func Run() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := sql.Open("postgres", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Error pinging database")
	}

	minioClient, err := minio_client.NewMinioClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKeyID,
		cfg.MinioSecretAccessKey,
		cfg.MinioBucketName)
	if err != nil {
		log.Fatal("Error connecting to minio")
	}

	if err = minioClient.CreateBucket(cfg.MinioLocation); err != nil {
		log.Fatal("Error creating bucket")
	}
	fmt.Println("Connected to minio and created bucket")

	Cache := cache.New()

	repo := repository.NewRepository(db)
	serv := service.NewService(repo)
	hand := handler.NewHandler(serv, cfg.SecretKey, Cache, minioClient)
	routes.Run(hand, cfg.ServerPort, cfg.SecretKey)
}
