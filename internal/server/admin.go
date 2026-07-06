package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"supervisor-game/internal/model"
	"supervisor-game/internal/service"
)

func (s *Server) adminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.cfg.AppKey == "" {
			s.writeAPIError(c, http.StatusForbidden, "APPKEY_NOT_CONFIGURED", "APP_KEY 未配置，管理端不可用。")
			c.Abort()
			return
		}
		appKey := c.GetHeader("X-App-Key")
		if appKey == "" {
			appKey = c.Query("appkey")
		}
		if appKey != s.cfg.AppKey {
			s.writeAPIError(c, http.StatusForbidden, "APPKEY_INVALID", "appkey 不正确。")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) requireService() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.svc == nil {
			s.writeError(c, service.ErrDatabaseUnavailable)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) adminRuntimeConfig(c *gin.Context) {
	result, err := s.svc.AdminRuntimeConfig(s.adminStatusInput())
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) adminCharacters(c *gin.Context) {
	items, err := s.svc.AdminCharacters()
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) createCharacter(c *gin.Context) {
	var input model.Character
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.CreateCharacter(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (s *Server) updateCharacter(c *gin.Context) {
	id, ok := s.paramID(c)
	if !ok {
		return
	}
	var input model.Character
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.UpdateCharacter(id, input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) deleteCharacter(c *gin.Context) {
	id, ok := s.paramID(c)
	if !ok {
		return
	}
	if err := s.svc.DeleteCharacter(id); err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *Server) adminScenes(c *gin.Context) {
	items, err := s.svc.AdminScenes()
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) createScene(c *gin.Context) {
	var input model.Scene
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.CreateScene(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (s *Server) updateScene(c *gin.Context) {
	id, ok := s.paramID(c)
	if !ok {
		return
	}
	var input model.Scene
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.UpdateScene(id, input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) deleteScene(c *gin.Context) {
	id, ok := s.paramID(c)
	if !ok {
		return
	}
	if err := s.svc.DeleteScene(id); err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *Server) adminActions(c *gin.Context) {
	items, err := s.svc.AdminActions(c.Query("sceneKey"))
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) createAction(c *gin.Context) {
	var input model.ActionConfig
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.CreateAction(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (s *Server) updateAction(c *gin.Context) {
	id, ok := s.paramID(c)
	if !ok {
		return
	}
	var input model.ActionConfig
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.UpdateAction(id, input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) deleteAction(c *gin.Context) {
	id, ok := s.paramID(c)
	if !ok {
		return
	}
	if err := s.svc.DeleteAction(id); err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *Server) adminModelConfig(c *gin.Context) {
	item, err := s.svc.AdminModelConfig()
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) updateModelConfig(c *gin.Context) {
	var input service.ModelConfigInput
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.UpdateModelConfig(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) testModelConfig(c *gin.Context) {
	c.JSON(http.StatusOK, s.svc.TestModelConfig())
}

func (s *Server) adminPatrolRule(c *gin.Context) {
	item, err := s.svc.AdminPatrolRule()
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) updatePatrolRule(c *gin.Context) {
	var input model.PatrolRule
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.UpdatePatrolRule(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) adminMySQLConfig(c *gin.Context) {
	item, err := s.svc.AdminMySQLConfig()
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) updateMySQLConfig(c *gin.Context) {
	var input service.MySQLConfigInput
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.UpdateMySQLConfig(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) testMySQLConfig(c *gin.Context) {
	var input service.MySQLConfigInput
	if !s.bindJSON(c, &input) {
		return
	}
	item, err := s.svc.TestMySQLConfig(input)
	status := http.StatusOK
	if err != nil {
		status = http.StatusBadRequest
	}
	c.JSON(status, item)
}

func (s *Server) migrateMySQLConfig(c *gin.Context) {
	if err := s.svc.MigrateCurrentDB(); err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"migrated": true})
}

func (s *Server) bindJSON(c *gin.Context, value any) bool {
	if err := c.ShouldBindJSON(value); err != nil {
		s.writeAPIError(c, http.StatusBadRequest, "INVALID_JSON", "请求 JSON 格式不正确。")
		return false
	}
	return true
}

func (s *Server) paramID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		s.writeAPIError(c, http.StatusBadRequest, "INVALID_ID", "id 必须是正整数。")
		return 0, false
	}
	return uint(value), true
}
