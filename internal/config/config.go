package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort        string
	DatabaseURL    string
	LogLevel       string
	PartnerShare   float64
	JWTSecret      string
	JWTExpiryHours int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	partnerShare, err := strconv.ParseFloat(getEnv("PARTNER_SHARE", "0.10"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid PARTNER_SHARE: %w", err)
	}
	if partnerShare <= 0 || partnerShare > 1 {
		return nil, fmt.Errorf("PARTNER_SHARE must be in (0, 1], got %f", partnerShare)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	jwtExpiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil || jwtExpiryHours <= 0 {
		return nil, fmt.Errorf("JWT_EXPIRY_HOURS must be a positive integer")
	}

	cfg := &Config{
		AppPort:        getEnv("APP_PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		PartnerShare:   partnerShare,
		JWTSecret:      jwtSecret,
		JWTExpiryHours: jwtExpiryHours,
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}

	return defaultValue
}
