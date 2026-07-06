package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"supervisor-game/internal/service"
)

func (s *Server) startSession(c *gin.Context) {
	if s.svc == nil {
		s.writeError(c, service.ErrDatabaseUnavailable)
		return
	}
	var input service.SessionStartInput
	if err := c.ShouldBindJSON(&input); err != nil {
		s.writeAPIError(c, http.StatusBadRequest, "INVALID_JSON", "请求 JSON 格式不正确。")
		return
	}
	result, err := s.svc.StartSession(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) pauseSession(c *gin.Context) {
	if s.svc == nil {
		s.writeError(c, service.ErrDatabaseUnavailable)
		return
	}
	var input service.SessionIDInput
	if err := c.ShouldBindJSON(&input); err != nil {
		s.writeAPIError(c, http.StatusBadRequest, "INVALID_JSON", "请求 JSON 格式不正确。")
		return
	}
	result, err := s.svc.PauseSession(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) resumeSession(c *gin.Context) {
	if s.svc == nil {
		s.writeError(c, service.ErrDatabaseUnavailable)
		return
	}
	var input service.SessionIDInput
	if err := c.ShouldBindJSON(&input); err != nil {
		s.writeAPIError(c, http.StatusBadRequest, "INVALID_JSON", "请求 JSON 格式不正确。")
		return
	}
	result, err := s.svc.ResumeSession(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) finishSession(c *gin.Context) {
	if s.svc == nil {
		s.writeError(c, service.ErrDatabaseUnavailable)
		return
	}
	var input service.SessionFinishInput
	if err := c.ShouldBindJSON(&input); err != nil {
		s.writeAPIError(c, http.StatusBadRequest, "INVALID_JSON", "请求 JSON 格式不正确。")
		return
	}
	result, err := s.svc.FinishSession(input)
	if err != nil {
		s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
