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
		log.Fatal("Error connecting to database")
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database")
	}

	minioClient, err := minio_client.NewMinioClient(cfg.MinioEndpoint, cfg.MinioAccessKeyID, cfg.MinioSecretAccessKey, cfg.MinioBucketName, cfg.MinioLocation)
	if err != nil {
		log.Fatal("Error connecting to minio")
	}

	minioClient.CreateBucket()

	//listObject, err := minioClient.ListObjectsInBucket(cfg.MinioBucketName)
	//if err != nil {
	//	log.Fatal("Error listing objects in bucket")
	//}
	//
	//for _, object := range listObject {
	//	fmt.Println(object)
	//}

	Cache := cache.New()

	repo := repository.NewRepository(db)
	serv := service.NewService(repo)
	hand := handler.NewHandler(serv, cfg.SecretKey, Cache, minioClient)
	routes.Run(hand, cfg.ServerPort, cfg.SecretKey)
}
