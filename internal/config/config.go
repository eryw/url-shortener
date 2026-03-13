package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL       string
	Port          string
	DBPath        string
	SessionSecret string
	ResetAdmin    bool
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	resetAdmin := false
	if val := os.Getenv("RESET_ADMIN"); val != "" {
		resetAdmin, _ = strconv.ParseBool(val)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/urls.db"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("BASE_URL environment variable is required")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET environment variable is required")
	}

	return &Config{
		BaseURL:       baseURL,
		Port:          port,
		DBPath:        dbPath,
		SessionSecret: sessionSecret,
		ResetAdmin:    resetAdmin,
	}
}
