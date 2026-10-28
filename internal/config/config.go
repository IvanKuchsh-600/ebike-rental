package config

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort           string
	DatabaseURL       string
	LogLevel          string
	PartnerShare      float64
	OwnerShare        float64
	AmortizationShare float64
	JWTSecret         string
	JWTExpiryHours    int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	partnerShare, err := strconv.ParseFloat(getEnv("PARTNER_SHARE", "0.40"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid PARTNER_SHARE: %w", err)
	}
	if partnerShare <= 0 || partnerShare > 1 {
		return nil, fmt.Errorf("PARTNER_SHARE must be in (0, 1], got %f", partnerShare)
	}

	ownerShare, err := strconv.ParseFloat(getEnv("OWNER_SHARE", "0.40"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid OWNER_SHARE: %w", err)
	}
	if ownerShare < 0 || ownerShare > 1 {
		return nil, fmt.Errorf("OWNER_SHARE must be in [0, 1], got %f", ownerShare)
	}

	amortizationShare, err := strconv.ParseFloat(getEnv("AMORTIZATION_SHARE", "0.20"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid AMORTIZATION_SHARE: %w", err)
	}
	if amortizationShare < 0 || amortizationShare > 1 {
		return nil, fmt.Errorf("AMORTIZATION_SHARE must be in [0, 1], got %f", amortizationShare)
	}

	// Проверка, что сумма = 1.0 (с погрешностью float)
	if math.Abs(ownerShare+partnerShare+amortizationShare-1.0) > 0.0001 {
		return nil, fmt.Errorf(
			"shares must sum to 1.0, got owner=%f + partner=%f + amortization=%f",
			ownerShare, partnerShare, amortizationShare,
		)
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
		AppPort:           getEnv("APP_PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		PartnerShare:      partnerShare,
		OwnerShare:        ownerShare,
		AmortizationShare: amortizationShare,
		JWTSecret:         jwtSecret,
		JWTExpiryHours:    jwtExpiryHours,
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
