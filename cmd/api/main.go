package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/iperez/new-expenses-go/internal/cache"
	"github.com/iperez/new-expenses-go/internal/config"
	"github.com/iperez/new-expenses-go/internal/database"
	"github.com/iperez/new-expenses-go/internal/server"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.NewPostgres(cfg.DatabaseURL, cfg.Env == "DEV" || cfg.Env == "LOCAL")
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	redisClient, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	srv := server.New(cfg, db, redisClient)

	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
