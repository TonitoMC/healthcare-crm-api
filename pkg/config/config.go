package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all environment-based configuration for the app.
type Config struct {
	// DB Config
	DatabaseURL string

	// JWT Config
	JWTSecret string        // secret key for signing tokens
	JWTTTL    time.Duration // token time-to-live (default 24h)
	JWTIssuer string        // issuer name in JWT claims

	// Superuser Config
	SuperuserName     string
	SuperuserEmail    string
	SuperuserPassword string

	SecretaryName     string
	SecretaryEmail    string
	SecretaryPassword string

	// --- New S3 / MinIO ---
	S3Bucket         string
	S3Region         string
	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3ForcePathStyle bool
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

	// Superuser
	cfg.SuperuserName = os.Getenv("SUPERUSER_NAME")
	cfg.SuperuserEmail = os.Getenv("SUPERUSER_EMAIL")
	cfg.SuperuserPassword = os.Getenv("SUPERUSER_PASSWORD")

	// Not required — if empty, bootstrap will just skip
	if cfg.SuperuserName == "" || cfg.SuperuserEmail == "" || cfg.SuperuserPassword == "" {
		log.Println("⚠️ SUPERUSER_* variables not set — skipping auto-creation")
	}

	// Secretary
	cfg.SecretaryName = os.Getenv("SECRETARY_NAME")
	cfg.SecretaryEmail = os.Getenv("SECRETARY_EMAIL")
	cfg.SecretaryPassword = os.Getenv("SECRETARY_PASSWORD")

	// Not required — if empty, bootstrap will just skip
	if cfg.SecretaryName == "" || cfg.SecretaryEmail == "" || cfg.SecretaryPassword == "" {
		log.Println("⚠️ SECRETARY_* variables not set — skipping auto-creation")
	}

	// --- S3 / MinIO ---
	cfg.S3Bucket = os.Getenv("S3_BUCKET")
	cfg.S3Region = os.Getenv("S3_REGION")
	cfg.S3Endpoint = os.Getenv("S3_ENDPOINT")
	cfg.S3AccessKey = os.Getenv("S3_ACCESS_KEY")
	cfg.S3SecretKey = os.Getenv("S3_SECRET_KEY")

	if v := os.Getenv("S3_FORCE_PATH_STYLE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			log.Printf("Invalid S3_FORCE_PATH_STYLE, defaulting to false")
			cfg.S3ForcePathStyle = false
		} else {
			cfg.S3ForcePathStyle = b
		}
	}

	if cfg.S3Bucket == "" {
		log.Println("⚠️  S3_BUCKET not set — file uploads will be disabled")
	}

	return cfg
}
