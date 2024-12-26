package config

import (
	"os"

	"github.com/sabirov8872/bookstore/pkg/minio"
	"github.com/sabirov8872/bookstore/pkg/postgres"
	"github.com/sabirov8872/bookstore/pkg/redis"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Postgres postgres.Config `yaml:"postgres"`
	Minio    minio.Config    `yaml:"minio"`
	Redis    redis.Config    `yaml:"redis"`
}

func Load() (*Config, error) {
	var c Config
	file, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
