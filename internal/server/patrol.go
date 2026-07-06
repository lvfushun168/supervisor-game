package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"supervisor-game/internal/service"
)

func (s *Server) checkPatrol(c *gin.Context) {
	if s.svc == nil {
		s.writeError(c, service.ErrDatabaseUnavailable)
		return
	}
	var input service.PatrolCheckInput
	if err := c.ShouldBindJSON(&input); err != nil {
		s.writeAPIError(c, http.StatusBadRequest, "INVALID_JSON", "请求 JSON 格式不正确。")
		return
	}
	result, err := s.svc.CheckPatrol(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
