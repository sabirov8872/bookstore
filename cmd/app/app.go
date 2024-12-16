package app

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sabirov8872/bookstore/config"
	"github.com/sabirov8872/bookstore/internal/handler"
	"github.com/sabirov8872/bookstore/internal/minioClient"
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

	MinioClient, err := minioClient.NewMinioClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKeyID,
		cfg.MinioSecretAccessKey,
		cfg.MinioBucketName)
	if err != nil {
		log.Fatal("Error connecting to minio")
	}

	err = MinioClient.CreateBucket(cfg.MinioLocation)
	if err != nil {
		log.Fatal("Error creating bucket")
	}

	cache := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: "",
		DB:       0,
	})
	fmt.Println("Connected to redis")

	repo := repository.NewRepository(db)
	serv := service.NewService(repo)
	hand := handler.NewHandler(serv, cfg.SecretKey, cache, MinioClient)
	routes.Run(hand, cfg.ServerPort, cfg.SecretKey)
}
