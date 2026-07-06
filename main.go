package main

import (
	"embed"
	"errors"
	"log"
	"net/http"

	"supervisor-game/internal/config"
	"supervisor-game/internal/database"
	"supervisor-game/internal/server"
)

//go:embed frontend/dist
var frontendDist embed.FS

func main() {
	cfg := config.Load()

	db, dbErr := database.Open(cfg.DBDSN)
	if errors.Is(dbErr, database.ErrDSNMissing) {
		log.Printf("database disabled: %v", dbErr)
	} else if dbErr != nil {
		log.Printf("database connection failed: %v", dbErr)
	}

	app := server.New(cfg, db, dbErr, frontendDist)
	if err := app.Migrate(); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	log.Printf("environment: %s", cfg.AppEnv)
	log.Printf("assets directory: %s", cfg.AssetsDir)
	if db != nil {
		log.Printf("database connected and migrated")
	}
	log.Printf("supervisor-game listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, app.Handler()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
