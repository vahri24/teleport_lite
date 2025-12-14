package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DSN       string
	JWTSecret string
	AppPort   string
}

func Load() Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, using system environment variables")
	} else {
		log.Println("✅ .env file loaded successfully!")
	}

	cfg := Config{
		DSN:       os.Getenv("MYSQL_DSN"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		AppPort:   os.Getenv("APP_PORT"),
	}

	if cfg.DSN == "" {
		log.Fatal("❌ MYSQL_DSN not set in environment")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "dev-secret-only"
	}
	if cfg.AppPort == "" {
		cfg.AppPort = "8080"
	}

	return cfg
}
