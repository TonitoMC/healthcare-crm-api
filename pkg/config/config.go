package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all environment-based configuration for the app.
type Config struct {
	DatabaseURL string
	JWTSecret   string        // secret key for signing tokens
	JWTTTL      time.Duration // token time-to-live (default 24h)
	JWTIssuer   string        // issuer name in JWT claims
}

// Load reads environment variables into a Config struct.
// It terminates the app early if any required value is missing.
func Load() *Config {
	cfg := &Config{}

	// Database connection URL
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	// JWT Secret
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET not set")
	}

	// Token expiration (in hours)
	if ttlStr := os.Getenv("JWT_TTL_HOURS"); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil && ttl > 0 {
			cfg.JWTTTL = time.Duration(ttl) * time.Hour
		} else {
			log.Printf("Invalid JWT_TTL_HOURS value, defaulting to 24h")
			cfg.JWTTTL = 24 * time.Hour
		}
	} else {
		cfg.JWTTTL = 24 * time.Hour
	}

	// JWT Issuer
	cfg.JWTIssuer = os.Getenv("JWT_ISSUER")
	if cfg.JWTIssuer == "" {
		log.Fatal("JWT_ISSUER not set")
	}

	return cfg
}
