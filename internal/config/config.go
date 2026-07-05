package config

import "os"

import "github.com/joho/godotenv"

type Config struct {
	AppEnv string
	Addr   string
	DBDSN  string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppEnv: getEnv("APP_ENV", "development"),
		Addr:   getEnv("APP_ADDR", ":8080"),
		DBDSN:  os.Getenv("DB_DSN"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
