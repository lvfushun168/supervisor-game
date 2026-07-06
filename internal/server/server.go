package server

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"supervisor-game/internal/config"
	"supervisor-game/internal/database"
	"supervisor-game/internal/repository"
	"supervisor-game/internal/service"
)

type Server struct {
	cfg        config.Config
	db         *gorm.DB
	svc        *service.Service
	distFS     fs.FS
	dbError    error
	dbSource   string
	dbMigrated bool
	startedAt  time.Time
}

func New(cfg config.Config, db *gorm.DB, dbError error, dbSource string, dbMigrated bool, distFS fs.FS) *Server {
	var svc *service.Service
	if db != nil {
		svc = service.New(cfg, repository.New(db))
	}
	if dbSource == "" {
		dbSource = "DB_DSN"
	}
	return &Server{
		cfg:        cfg,
		db:         db,
		svc:        svc,
		distFS:     distFS,
		dbError:    dbError,
		dbSource:   dbSource,
		dbMigrated: dbMigrated,
		startedAt:  time.Now(),
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
	api.GET("/runtime/config", s.runtimeConfig)
	api.GET("/scenes", s.scenes)
	api.GET("/settings", s.getSettings)
	api.PUT("/settings", s.putSettings)

	admin := api.Group("/admin")
	admin.Use(s.adminAuth())
	admin.Use(s.requireService())
	admin.GET("/status", s.adminStatus)
	admin.GET("/runtime-config", s.adminRuntimeConfig)
	admin.GET("/characters", s.adminCharacters)
	admin.POST("/characters", s.createCharacter)
	admin.PUT("/characters/:id", s.updateCharacter)
	admin.DELETE("/characters/:id", s.deleteCharacter)
	admin.GET("/scenes", s.adminScenes)
	admin.POST("/scenes", s.createScene)
	admin.PUT("/scenes/:id", s.updateScene)
	admin.DELETE("/scenes/:id", s.deleteScene)
	admin.GET("/actions", s.adminActions)
	admin.POST("/actions", s.createAction)
	admin.PUT("/actions/:id", s.updateAction)
	admin.DELETE("/actions/:id", s.deleteAction)
	admin.GET("/model-config", s.adminModelConfig)
	admin.PUT("/model-config", s.updateModelConfig)
	admin.POST("/model-config/test", s.testModelConfig)
	admin.GET("/patrol-rule", s.adminPatrolRule)
	admin.PUT("/patrol-rule", s.updatePatrolRule)
	admin.GET("/mysql-config", s.adminMySQLConfig)
	admin.PUT("/mysql-config", s.updateMySQLConfig)
	admin.POST("/mysql-config/test", s.testMySQLConfig)
	admin.POST("/mysql-config/migrate", s.migrateMySQLConfig)

	router.Static("/assets", s.cfg.AssetsDir)
	s.mountFrontend(router)

	return router
}

func (s *Server) Migrate() error {
	if s.db == nil {
		return nil
	}
	if !s.dbMigrated {
		if err := database.Migrate(s.db); err != nil {
			return err
		}
		s.dbMigrated = true
	}
	return s.svc.SeedDefaults()
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"appEnv":   s.cfg.AppEnv,
		"database": s.databaseStatus(),
	})
}

func (s *Server) adminStatus(c *gin.Context) {
	if s.svc == nil {
		c.JSON(http.StatusOK, gin.H{"service": "supervisor-game", "database": s.databaseStatus()})
		return
	}
	status, err := s.svc.AdminStatus(s.adminStatusInput())
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func (s *Server) runtimeConfig(c *gin.Context) {
	if s.svc == nil {
		s.writeError(c, service.ErrDatabaseUnavailable)
		return
	}
	runtimeConfig, err := s.svc.RuntimeConfig()
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, runtimeConfig)
}

func (s *Server) scenes(c *gin.Context) {
	if s.svc == nil {
		s.writeError(c, service.ErrDatabaseUnavailable)
		return
	}
	scenes, err := s.svc.EnabledScenes()
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": scenes})
}

func (s *Server) getSettings(c *gin.Context) {
	if s.svc == nil {
		s.writeError(c, service.ErrDatabaseUnavailable)
		return
	}
	setting, err := s.svc.UserSetting()
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, setting)
}

func (s *Server) putSettings(c *gin.Context) {
	if s.svc == nil {
		s.writeError(c, service.ErrDatabaseUnavailable)
		return
	}
	var input service.UserSettingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		s.writeAPIError(c, http.StatusBadRequest, "INVALID_JSON", "请求 JSON 格式不正确。")
		return
	}

	setting, err := s.svc.UpdateUserSetting(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, setting)
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

func (s *Server) adminStatusInput() service.AdminStatusInput {
	message := ""
	if s.dbError != nil {
		message = s.dbError.Error()
	}
	return service.AdminStatusInput{
		StartedAt:    s.startedAt,
		Addr:         s.cfg.Addr,
		AssetsDir:    s.cfg.AssetsDir,
		DBStatus:     s.databaseStatus(),
		DBSource:     s.dbSource,
		DBError:      message,
		BootstrapDSN: s.cfg.DBDSN != "",
	}
}

func (s *Server) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrDatabaseUnavailable):
		s.writeAPIError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "数据库暂不可用，请检查 DB_DSN。")
	case errors.Is(err, service.ErrInvalidInput):
		s.writeAPIError(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
	case errors.Is(err, gorm.ErrRecordNotFound):
		s.writeAPIError(c, http.StatusNotFound, "NOT_FOUND", "请求的数据不存在。")
	default:
		s.writeAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}

func (s *Server) writeAPIError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
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
