package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	Port          string
	GinMode       string
	JWTSecret     string
	JWTExpiration time.Duration
}

var AppConfig *Config

func LoadConfig() error {
	// Load .env file if it exists
	_ = godotenv.Load()

	jwtExp := os.Getenv("JWT_EXPIRATION")
	if jwtExp == "" {
		jwtExp = "24h"
	}

	duration, err := time.ParseDuration(jwtExp)
	if err != nil {
		duration = 24 * time.Hour
	}

	AppConfig = &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          os.Getenv("PORT"),
		GinMode:       os.Getenv("GIN_MODE"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiration: duration,
	}

	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
	}

	if AppConfig.GinMode == "" {
		AppConfig.GinMode = "debug"
	}

	if AppConfig.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable not set")
	}

	return nil
}
