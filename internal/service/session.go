package service

import (
	"errors"
	"fmt"
	"time"

	"supervisor-game/internal/model"

	"gorm.io/gorm"
)

type SessionStartInput struct {
	SceneKey               string         `json:"sceneKey"`
	Mode                   string         `json:"mode"`
	PlannedDurationSeconds int            `json:"plannedDurationSeconds"`
	UserConfig             map[string]any `json:"userConfig"`
}

type SessionIDInput struct {
	SessionID uint `json:"sessionId"`
}

type SessionFinishInput struct {
	SessionID          uint   `json:"sessionId"`
	FinishReason       string `json:"finishReason"`
	ActualFocusSeconds int    `json:"actualFocusSeconds"`
}

type SessionResponse struct {
	Session SessionView `json:"session"`
}

type SessionView struct {
	ID                     uint      `json:"id"`
	SceneKey               string    `json:"sceneKey"`
	Mode                   string    `json:"mode"`
	PlannedDurationSeconds int       `json:"plannedDurationSeconds"`
	StartedAt              time.Time `json:"startedAt"`
	Status                 string    `json:"status"`
	WarningCount           int       `json:"warningCount"`
	ViolationCount         int       `json:"violationCount"`
}

type SettlementResponse struct {
	Settlement SettlementView `json:"settlement"`
}

type SettlementView struct {
	SessionID          uint             `json:"sessionId"`
	Result             string           `json:"result"`
	ActualFocusSeconds int              `json:"actualFocusSeconds"`
	PatrolCount        int              `json:"patrolCount"`
	WarningCount       int              `json:"warningCount"`
	ViolationCount     int              `json:"violationCount"`
	EarnedCurrency     int              `json:"earnedCurrency"`
	LevelBefore        int              `json:"levelBefore"`
	LevelAfter         int              `json:"levelAfter"`
	CurrencyAfter      int              `json:"currencyAfter"`
	SettlementAction   SettlementAction `json:"settlementAction"`
}

type SettlementAction struct {
	ActionKey string `json:"actionKey"`
	VideoURL  string `json:"videoUrl"`
	PosterURL string `json:"posterUrl"`
}

func (s *Service) StartSession(input SessionStartInput) (SessionResponse, error) {
	if s.repo == nil || s.repo.DB() == nil {
		return SessionResponse{}, ErrDatabaseUnavailable
	}
	if !allowed(input.Mode, "pomodoro", "custom", "infinite") {
		return SessionResponse{}, fmt.Errorf("%w: mode must be one of pomodoro, custom, infinite", ErrInvalidInput)
	}
	if input.Mode != "infinite" && input.PlannedDurationSeconds < 300 {
		return SessionResponse{}, fmt.Errorf("%w: plannedDurationSeconds must be at least 300", ErrInvalidInput)
	}
	if input.Mode == "infinite" && input.PlannedDurationSeconds < 0 {
		return SessionResponse{}, fmt.Errorf("%w: plannedDurationSeconds cannot be negative", ErrInvalidInput)
	}
	var scene model.Scene
	if err := s.repo.DB().Where("scene_key = ? AND enabled = ?", input.SceneKey, true).First(&scene).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SessionResponse{}, fmt.Errorf("%w: sceneKey must reference an enabled scene", ErrInvalidInput)
		}
		return SessionResponse{}, err
	}

	now := s.now()
	var created model.WorkSession
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		if err := abandonOpenSessions(tx, now); err != nil {
			return err
		}
		session := model.WorkSession{
			SceneKey:               input.SceneKey,
			Mode:                   input.Mode,
			PlannedDurationSeconds: input.PlannedDurationSeconds,
			StartedAt:              now,
			ActualFocusSeconds:     0,
			PatrolCount:            0,
			WarningCount:           0,
			ViolationCount:         0,
			Status:                 "working",
			Result:                 "",
			FinishReason:           "",
			EarnedCurrency:         0,
		}
		if err := tx.Create(&session).Error; err != nil {
			return err
		}
		created = session
		return nil
	})
	if err != nil {
		return SessionResponse{}, err
	}
	return SessionResponse{Session: toSessionView(created)}, nil
}

func (s *Service) PauseSession(input SessionIDInput) (SessionResponse, error) {
	return s.updateSessionStatus(input.SessionID, "paused")
}

func (s *Service) ResumeSession(input SessionIDInput) (SessionResponse, error) {
	return s.updateSessionStatus(input.SessionID, "working")
}

func (s *Service) FinishSession(input SessionFinishInput) (SettlementResponse, error) {
	if s.repo == nil || s.repo.DB() == nil {
		return SettlementResponse{}, ErrDatabaseUnavailable
	}
	if input.SessionID == 0 {
		return SettlementResponse{}, fmt.Errorf("%w: sessionId is required", ErrInvalidInput)
	}
	if !allowed(input.FinishReason, "countdown_complete", "user_stop", "max_warning", "max_violation", "page_unload") {
		return SettlementResponse{}, fmt.Errorf("%w: finishReason is invalid", ErrInvalidInput)
	}
	if input.ActualFocusSeconds < 0 {
		return SettlementResponse{}, fmt.Errorf("%w: actualFocusSeconds cannot be negative", ErrInvalidInput)
	}

	var settlement SettlementView
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var session model.WorkSession
		if err := tx.First(&session, input.SessionID).Error; err != nil {
			return err
		}
		if session.EndedAt != nil {
			settlement = s.settlementFromSession(tx, session)
			return nil
		}

		now := s.now()
		result := resultForFinishReason(input.FinishReason)
		earned := earnedCurrency(result, input.ActualFocusSeconds)
		levelBefore, levelAfter, currencyAfter, err := applyProgress(tx, input.ActualFocusSeconds, earned, result)
		if err != nil {
			return err
		}

		session.EndedAt = &now
		session.ActualFocusSeconds = input.ActualFocusSeconds
		session.Result = result
		session.FinishReason = input.FinishReason
		session.EarnedCurrency = earned
		session.Status = statusForResult(result)
		if err := tx.Save(&session).Error; err != nil {
			return err
		}
		settlement = s.settlementFromSession(tx, session)
		settlement.LevelBefore = levelBefore
		settlement.LevelAfter = levelAfter
		settlement.CurrencyAfter = currencyAfter
		return nil
	})
	if err != nil {
		return SettlementResponse{}, err
	}
	return SettlementResponse{Settlement: settlement}, nil
}

func (s *Service) updateSessionStatus(sessionID uint, status string) (SessionResponse, error) {
	if s.repo == nil || s.repo.DB() == nil {
		return SessionResponse{}, ErrDatabaseUnavailable
	}
	if sessionID == 0 {
		return SessionResponse{}, fmt.Errorf("%w: sessionId is required", ErrInvalidInput)
	}
	var session model.WorkSession
	if err := s.repo.DB().First(&session, sessionID).Error; err != nil {
		return SessionResponse{}, err
	}
	if session.EndedAt != nil {
		return SessionResponse{}, fmt.Errorf("%w: session already ended", ErrInvalidInput)
	}
	session.Status = status
	if err := s.repo.DB().Save(&session).Error; err != nil {
		return SessionResponse{}, err
	}
	return SessionResponse{Session: toSessionView(session)}, nil
}

func (s *Service) settlementFromSession(tx *gorm.DB, session model.WorkSession) SettlementView {
	progress := model.UserProgress{}
	_ = tx.Order("id ASC").First(&progress).Error
	action := settlementActionForSession(tx, session)
	return SettlementView{
		SessionID:          session.ID,
		Result:             session.Result,
		ActualFocusSeconds: session.ActualFocusSeconds,
		PatrolCount:        session.PatrolCount,
		WarningCount:       session.WarningCount,
		ViolationCount:     session.ViolationCount,
		EarnedCurrency:     session.EarnedCurrency,
		LevelBefore:        progress.Level,
		LevelAfter:         progress.Level,
		CurrencyAfter:      progress.Currency,
		SettlementAction:   action,
	}
}

func toSessionView(session model.WorkSession) SessionView {
	return SessionView{
		ID:                     session.ID,
		SceneKey:               session.SceneKey,
		Mode:                   session.Mode,
		PlannedDurationSeconds: session.PlannedDurationSeconds,
		StartedAt:              session.StartedAt,
		Status:                 session.Status,
		WarningCount:           session.WarningCount,
		ViolationCount:         session.ViolationCount,
	}
}

func abandonOpenSessions(tx *gorm.DB, now time.Time) error {
	return tx.Model(&model.WorkSession{}).
		Where("ended_at IS NULL").
		Updates(map[string]any{
			"ended_at":      now,
			"status":        "failed",
			"result":        "abandoned",
			"finish_reason": "page_unload",
		}).Error
}

func resultForFinishReason(reason string) string {
	switch reason {
	case "countdown_complete":
		return "success"
	case "max_warning", "max_violation":
		return "failed"
	case "page_unload":
		return "abandoned"
	default:
		return "left"
	}
}

func statusForResult(result string) string {
	if result == "failed" || result == "abandoned" {
		return "failed"
	}
	return "finished"
}

func earnedCurrency(result string, focusSeconds int) int {
	if result != "success" {
		return 0
	}
	if focusSeconds >= 600 {
		return 10
	}
	return 0
}

func applyProgress(tx *gorm.DB, focusSeconds int, earned int, result string) (int, int, int, error) {
	var progress model.UserProgress
	err := tx.Order("id ASC").First(&progress).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		progress = DefaultUserProgress()
		if err := tx.Create(&progress).Error; err != nil {
			return 1, 1, 0, err
		}
	} else if err != nil {
		return 1, 1, 0, err
	}

	levelBefore := progress.Level
	if result != "abandoned" {
		progress.TotalFocusSeconds += focusSeconds
		progress.Currency += earned
		progress.Level = levelForFocus(progress.TotalFocusSeconds)
		today := time.Now()
		progress.LastFocusDate = &today
		if err := tx.Save(&progress).Error; err != nil {
			return levelBefore, progress.Level, progress.Currency, err
		}
	}
	return levelBefore, progress.Level, progress.Currency, nil
}

func levelForFocus(totalSeconds int) int {
	switch {
	case totalSeconds >= 43200:
		return 5
	case totalSeconds >= 21600:
		return 4
	case totalSeconds >= 10800:
		return 3
	case totalSeconds >= 3600:
		return 2
	default:
		return 1
	}
}

func settlementActionForSession(tx *gorm.DB, session model.WorkSession) SettlementAction {
	actionKey := "finish_success"
	if session.Result == "failed" || session.Result == "abandoned" {
		actionKey = "fail"
	}
	var action model.ActionConfig
	err := tx.Where("scene_key = ? AND action_key = ?", session.SceneKey, actionKey).First(&action).Error
	if err != nil {
		return SettlementAction{ActionKey: actionKey}
	}
	return SettlementAction{
		ActionKey: action.ActionKey,
		VideoURL:  action.VideoURL,
		PosterURL: action.PosterURL,
	}
}
