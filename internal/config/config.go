package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SecretKey            string
	ServerPort           string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	MinioEndpoint        string
	MinioAccessKeyID     string
	MinioSecretAccessKey string
	MinioBucketName      string
	MinioLocation        string
	RedisAddress         string
}

func GetConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Config{
		SecretKey:            os.Getenv("SECRET_KEY"),
		ServerPort:           os.Getenv("SERVER_PORT"),
		DBHost:               os.Getenv("DB_HOST"),
		DBPort:               os.Getenv("DB_PORT"),
		DBUser:               os.Getenv("DB_USER"),
		DBPassword:           os.Getenv("DB_PASSWORD"),
		DBName:               os.Getenv("DB_NAME"),
		DBSSLMode:            os.Getenv("DB_SSL_MODE"),
		MinioEndpoint:        os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKeyID:     os.Getenv("MINIO_ACCESS_KEY_ID"),
		MinioSecretAccessKey: os.Getenv("MINIO_SECRET_ACCESS_KEY"),
		MinioBucketName:      os.Getenv("MINIO_BUCKET"),
		MinioLocation:        os.Getenv("MINIO_LOCATION"),
		RedisAddress:         os.Getenv("REDIS_ADDRESS"),
	}, nil
}
