package server

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"supervisor-game/internal/config"
	"supervisor-game/internal/database"
	"supervisor-game/internal/model"
)

type Server struct {
	cfg     config.Config
	db      *gorm.DB
	distFS  fs.FS
	dbError error
}

func New(cfg config.Config, db *gorm.DB, dbError error, distFS fs.FS) *Server {
	return &Server{
		cfg:     cfg,
		db:      db,
		distFS:  distFS,
		dbError: dbError,
	}
}

func (s *Server) Handler() http.Handler {
	if s.cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	api := router.Group("/api")
	api.GET("/health", s.health)
	api.GET("/admin/status", s.adminStatus)

	s.mountFrontend(router)

	return router
}

func (s *Server) Migrate() error {
	if s.db == nil {
		return nil
	}
	return s.db.AutoMigrate(&model.Setting{})
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"appEnv":   s.cfg.AppEnv,
		"database": s.databaseStatus(),
	})
}

func (s *Server) adminStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":  "supervisor-game",
		"database": s.databaseStatus(),
	})
}

func (s *Server) databaseStatus() gin.H {
	if s.db == nil {
		status := "not_configured"
		if s.dbError != nil && !errors.Is(s.dbError, database.ErrDSNMissing) {
			status = "error"
		}

		message := ""
		if s.dbError != nil {
			message = s.dbError.Error()
		}

		return gin.H{
			"status":  status,
			"message": message,
		}
	}

	return gin.H{"status": "connected"}
}

func (s *Server) mountFrontend(router *gin.Engine) {
	if s.distFS == nil {
		router.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "frontend dist is not embedded; run npm run build in frontend before building the Go binary"})
		})
		return
	}

	assets, err := fs.Sub(s.distFS, "frontend/dist")
	if err != nil {
		router.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "frontend dist is unavailable"})
		})
		return
	}

	fileServer := http.FileServer(http.FS(assets))
	router.NoRoute(func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(assets, path); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		if _, err := fs.Stat(assets, "index.html"); errors.Is(err, fs.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "frontend index.html is missing"})
			return
		}

		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
