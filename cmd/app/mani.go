package app

import (
	"database/sql"
	"fmt"
	"github.com/sabirov8872/bookstore/integral/config"
	"log"
)

func Run() {
	cfg, err := config.NewConfig()
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

	fmt.Println("Connected to database")

}
