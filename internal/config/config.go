package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	Addr                string
	AppKey              string
	DBDSN               string
	AssetsDir           string
	ConfigEncryptionKey string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		Addr:                getEnv("APP_ADDR", ":8080"),
		AppKey:              os.Getenv("APP_KEY"),
		DBDSN:               os.Getenv("DB_DSN"),
		AssetsDir:           getEnv("ASSETS_DIR", "assets"),
		ConfigEncryptionKey: os.Getenv("CONFIG_ENCRYPTION_KEY"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
