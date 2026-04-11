package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string
	Env  string

	// Database (app - pooler)
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// Database (migration - direct)
	MigrateDBHost string
	MigrateDBPort string

	// Supabase Auth
	SupabaseJWTSecret string // Dashboard → Settings → API → JWT Secret
	SupabaseJWKSUrl string // GET https://project-id.supabase.co/auth/v1/.well-known/jwks.json

	// Google Auth
	GoogleJWKSUrl string // GET https://www.googleapis.com/oauth2/v3/certs

	AllowedOrigin	string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:              	getEnv("PORT", "8080"),
		Env:               	getEnv("APP_ENV", "development"),
		DBHost:            	getEnv("DB_HOST", "localhost"),
		DBPort:            	getEnv("DB_PORT", "6543"),
		DBName:            	getEnv("DB_NAME", "postgres"),
		DBUser:            	getEnv("DB_USER", "postgres"),
		DBPassword:        	getEnv("DB_PASSWORD", ""),
		DBSSLMode:         	getEnv("DB_SSLMODE", "require"),
		MigrateDBHost:     	getEnv("MIGRATE_DB_HOST", "localhost"),
		MigrateDBPort:     	getEnv("MIGRATE_DB_PORT", "5432"),
		SupabaseJWTSecret: 	getEnv("SUPABASE_JWT_SECRET", ""),
		SupabaseJWKSUrl:   	getEnv("SUPABASE_JWKS_URL", ""),
		GoogleJWKSUrl:     	getEnv("GOOGLE_JWKS_URL", "https://www.googleapis.com/oauth2/v3/certs"),
		AllowedOrigin:    	getEnv("ALLOWED_ORIGIN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
