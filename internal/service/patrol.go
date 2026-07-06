package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	appcrypto "supervisor-game/internal/crypto"
	"supervisor-game/internal/model"

	"gorm.io/gorm"
)

const (
	PatrolErrModelConfigMissing   = "MODEL_CONFIG_MISSING"
	PatrolErrModelCallFailed      = "MODEL_CALL_FAILED"
	PatrolErrModelResponseInvalid = "MODEL_RESPONSE_INVALID"
	PatrolErrCameraFrameMissing   = "CAMERA_FRAME_MISSING"
	PatrolErrActionConfigMissing  = "ACTION_CONFIG_MISSING"
)

type PatrolError struct {
	Code    string
	Message string
}

func (e PatrolError) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Message
}

type PatrolCheckInput struct {
	SessionID        uint   `json:"sessionId"`
	SceneKey         string `json:"sceneKey"`
	ImageBase64      string `json:"imageBase64"`
	CameraEnabled    bool   `json:"cameraEnabled"`
	ManualViolation  bool   `json:"manualViolation"`
	CaptureErrorCode string `json:"captureErrorCode"`
}

type PatrolCheckResponse struct {
	Status         string               `json:"status"`
	Confidence     float64              `json:"confidence"`
	Reason         string               `json:"reason"`
	Action         PatrolActionView     `json:"action"`
	WarningDelta   int                  `json:"warningDelta"`
	ViolationDelta int                  `json:"violationDelta"`
	SessionSummary PatrolSessionSummary `json:"sessionSummary"`
}

type PatrolActionView struct {
	ActionKey  string `json:"actionKey"`
	Name       string `json:"name"`
	VideoURL   string `json:"videoUrl"`
	PosterURL  string `json:"posterUrl"`
	DurationMS int    `json:"durationMs"`
}

type PatrolSessionSummary struct {
	SessionID      uint   `json:"sessionId"`
	PatrolCount    int    `json:"patrolCount"`
	WarningCount   int    `json:"warningCount"`
	ViolationCount int    `json:"violationCount"`
	Failed         bool   `json:"failed"`
	FinishReason   string `json:"finishReason"`
}

type modelPatrolResult struct {
	Status     string          `json:"status"`
	Confidence float64         `json:"confidence"`
	Reason     string          `json:"reason"`
	ActionKey  string          `json:"actionKey"`
	Raw        json.RawMessage `json:"-"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content any `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *Service) CheckPatrol(input PatrolCheckInput) (PatrolCheckResponse, error) {
	if s.repo == nil || s.repo.DB() == nil {
		return PatrolCheckResponse{}, ErrDatabaseUnavailable
	}
	if input.SessionID == 0 {
		return PatrolCheckResponse{}, fmt.Errorf("%w: sessionId is required", ErrInvalidInput)
	}
	if input.SceneKey == "" {
		return PatrolCheckResponse{}, fmt.Errorf("%w: sceneKey is required", ErrInvalidInput)
	}

	var response PatrolCheckResponse
	var patrolReturnErr error
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var session model.WorkSession
		if err := tx.First(&session, input.SessionID).Error; err != nil {
			return err
		}
		if session.EndedAt != nil || session.Status == "finished" || session.Status == "failed" {
			return fmt.Errorf("%w: session already ended", ErrInvalidInput)
		}
		if !allowed(session.Status, "working", "paused") {
			return fmt.Errorf("%w: session status cannot be patrolled", ErrInvalidInput)
		}
		if session.SceneKey != input.SceneKey {
			return fmt.Errorf("%w: sceneKey does not match session", ErrInvalidInput)
		}

		var scene model.Scene
		if err := tx.Where("scene_key = ? AND enabled = ?", input.SceneKey, true).First(&scene).Error; err != nil {
			return err
		}
		rule, err := firstPatrolRule(tx)
		if err != nil {
			return err
		}

		result, modelErr := s.patrolDecision(tx, input, rule)
		if modelErr != nil {
			record := patrolRecordFromError(input, rule, modelErr)
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			patrolReturnErr = modelErr
			return nil
		}

		action, actionErr := actionForPatrolResult(tx, scene, result)
		if actionErr != nil {
			record := patrolRecordFromResult(input, result, 0, 0)
			record.ErrorCode = PatrolErrActionConfigMissing
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			patrolReturnErr = actionErr
			return nil
		}

		warningDelta, violationDelta := patrolDeltas(result.Status, rule)
		session.PatrolCount++
		session.WarningCount += warningDelta
		session.ViolationCount += violationDelta
		failed, finishReason := patrolFailure(session, result.Status, rule)
		if failed {
			now := s.now()
			session.Status = "failed"
			session.Result = "failed"
			session.FinishReason = finishReason
			session.EndedAt = &now
			session.ActualFocusSeconds = int(now.Sub(session.StartedAt).Seconds())
			if session.ActualFocusSeconds < 0 {
				session.ActualFocusSeconds = 0
			}
		}
		if err := tx.Save(&session).Error; err != nil {
			return err
		}

		record := patrolRecordFromResult(input, result, warningDelta, violationDelta)
		record.ActionKey = action.ActionKey
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		response = PatrolCheckResponse{
			Status:         result.Status,
			Confidence:     result.Confidence,
			Reason:         result.Reason,
			Action:         toPatrolActionView(action),
			WarningDelta:   warningDelta,
			ViolationDelta: violationDelta,
			SessionSummary: PatrolSessionSummary{
				SessionID:      session.ID,
				PatrolCount:    session.PatrolCount,
				WarningCount:   session.WarningCount,
				ViolationCount: session.ViolationCount,
				Failed:         failed,
				FinishReason:   finishReason,
			},
		}
		return nil
	})
	if err != nil {
		return PatrolCheckResponse{}, err
	}
	if patrolReturnErr != nil {
		return PatrolCheckResponse{}, patrolReturnErr
	}
	return response, nil
}

func (s *Service) patrolDecision(tx *gorm.DB, input PatrolCheckInput, rule model.PatrolRule) (modelPatrolResult, error) {
	if input.ManualViolation {
		return localPatrolResult("violation", "手动标记违规。", `{"source":"manual"}`), nil
	}
	if !input.CameraEnabled {
		return localPatrolResult(rule.CameraOffStrategy, "摄像头已关闭，按巡查规则处理。", `{"source":"camera_off"}`), nil
	}
	if input.CaptureErrorCode != "" {
		raw := fmt.Sprintf(`{"source":"capture_error","captureErrorCode":%q}`, input.CaptureErrorCode)
		return localPatrolResult(rule.CaptureFailedStrategy, "摄像头截图失败，按巡查规则处理。", raw), nil
	}
	if strings.TrimSpace(input.ImageBase64) == "" {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrCameraFrameMissing, Message: rule.UserErrorMessage}
	}

	var cfg model.ModelConfig
	err := tx.Where("enabled = ?", true).Order("updated_at DESC, id DESC").First(&cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrModelConfigMissing, Message: rule.UserErrorMessage}
	}
	if err != nil {
		return modelPatrolResult{}, err
	}
	return s.callVisionModel(cfg, rule, input.ImageBase64)
}

func (s *Service) callVisionModel(cfg model.ModelConfig, rule model.PatrolRule, imageBase64 string) (modelPatrolResult, error) {
	if cfg.BaseURL == "" || cfg.Model == "" || cfg.APIKeyEncrypted == "" {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrModelConfigMissing, Message: rule.UserErrorMessage}
	}
	apiKey, err := appcrypto.DecryptString(s.cfg.ConfigEncryptionKey, cfg.APIKeyEncrypted)
	if err != nil {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrModelConfigMissing, Message: rule.UserErrorMessage}
	}

	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	retries := cfg.RetryCount
	if rule.ModelTimeoutRetryCount > retries {
		retries = rule.ModelTimeoutRetryCount
	}
	if retries < 0 {
		retries = 0
	}

	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		result, err := callVisionModelOnce(cfg, apiKey, imageBase64, timeout)
		if err == nil {
			return result, nil
		}
		var patrolErr PatrolError
		if errors.As(err, &patrolErr) && patrolErr.Code == PatrolErrModelResponseInvalid {
			return modelPatrolResult{}, err
		}
		lastErr = err
	}
	return modelPatrolResult{}, PatrolError{Code: PatrolErrModelCallFailed, Message: fmt.Sprintf("%s: %v", rule.UserErrorMessage, lastErr)}
}

func callVisionModelOnce(cfg model.ModelConfig, apiKey string, imageBase64 string, timeout time.Duration) (modelPatrolResult, error) {
	endpoint := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	body := map[string]any{
		"model":       cfg.Model,
		"temperature": cfg.Temperature,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": cfg.Prompt},
					{"type": "image_url", "image_url": map[string]any{"url": imageBase64}},
				},
			},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return modelPatrolResult{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return modelPatrolResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return modelPatrolResult{}, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return modelPatrolResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return modelPatrolResult{}, fmt.Errorf("vision model returned %d", resp.StatusCode)
	}
	var parsed openAIChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrModelResponseInvalid, Message: "模型响应格式不正确。"}
	}
	if len(parsed.Choices) == 0 {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrModelResponseInvalid, Message: "模型未返回 choices。"}
	}
	content, err := normalizeModelContent(parsed.Choices[0].Message.Content)
	if err != nil {
		return modelPatrolResult{}, err
	}
	return parseModelPatrolResult([]byte(content))
}

func normalizeModelContent(content any) (string, error) {
	switch value := content.(type) {
	case string:
		return value, nil
	case []any:
		var b strings.Builder
		for _, part := range value {
			item, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := item["text"].(string); ok {
				b.WriteString(text)
			}
		}
		if b.Len() > 0 {
			return b.String(), nil
		}
	}
	return "", PatrolError{Code: PatrolErrModelResponseInvalid, Message: "模型响应内容为空。"}
}

func parseModelPatrolResult(content []byte) (modelPatrolResult, error) {
	var payload struct {
		Status     string  `json:"status"`
		Confidence float64 `json:"confidence"`
		Reason     string  `json:"reason"`
		ActionKey  string  `json:"actionKey"`
	}
	if err := json.Unmarshal(content, &payload); err != nil {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrModelResponseInvalid, Message: "模型返回 JSON 不合法。"}
	}
	if !allowedPatrolStatus(payload.Status) {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrModelResponseInvalid, Message: "模型返回 status 不合法。"}
	}
	if payload.Confidence < 0 || payload.Confidence > 1 {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrModelResponseInvalid, Message: "模型返回 confidence 不合法。"}
	}
	if payload.Reason == "" {
		return modelPatrolResult{}, PatrolError{Code: PatrolErrModelResponseInvalid, Message: "模型返回 reason 不能为空。"}
	}
	return modelPatrolResult{
		Status:     payload.Status,
		Confidence: payload.Confidence,
		Reason:     payload.Reason,
		ActionKey:  payload.ActionKey,
		Raw:        append(json.RawMessage(nil), content...),
	}, nil
}

func actionForPatrolResult(tx *gorm.DB, scene model.Scene, result modelPatrolResult) (model.ActionConfig, error) {
	actionKey := result.ActionKey
	if actionKey != "" {
		action, err := findEnabledAction(tx, scene.SceneKey, actionKey)
		if err == nil {
			return action, nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ActionConfig{}, PatrolError{Code: PatrolErrActionConfigMissing, Message: "巡查动作配置缺失，请联系管理员处理。"}
		}
		if err != nil {
			return model.ActionConfig{}, err
		}
	}
	actionKey = mappedActionKey(scene.ModelResultActionMapJSON, result.Status)
	if actionKey == "" {
		return model.ActionConfig{}, PatrolError{Code: PatrolErrActionConfigMissing, Message: "巡查动作配置缺失，请联系管理员处理。"}
	}
	action, err := findEnabledAction(tx, scene.SceneKey, actionKey)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ActionConfig{}, PatrolError{Code: PatrolErrActionConfigMissing, Message: "巡查动作配置缺失，请联系管理员处理。"}
	}
	return action, err
}

func findEnabledAction(tx *gorm.DB, sceneKey string, actionKey string) (model.ActionConfig, error) {
	var action model.ActionConfig
	err := tx.Where("scene_key = ? AND action_key = ? AND enabled = ?", sceneKey, actionKey, true).First(&action).Error
	return action, err
}

func mappedActionKey(mappingJSON string, status string) string {
	var mapping map[string]string
	if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
		return ""
	}
	return mapping[status]
}

func firstPatrolRule(tx *gorm.DB) (model.PatrolRule, error) {
	var rule model.PatrolRule
	err := tx.Order("id ASC").First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		rule = DefaultPatrolRule()
		err = tx.Create(&rule).Error
	}
	return rule, err
}

func patrolDeltas(status string, rule model.PatrolRule) (int, int) {
	switch status {
	case "normal":
		return 0, 0
	case "suspicious":
		if rule.SuspiciousAddsWarning {
			return 1, 0
		}
		return 0, 0
	case "violation", "using_phone", "sleeping", "absent":
		return 1, 1
	case "uncertain":
		return 1, 0
	default:
		return 0, 0
	}
}

func patrolFailure(session model.WorkSession, status string, rule model.PatrolRule) (bool, string) {
	if rule.ViolationDirectFail && isViolationStatus(status) {
		return true, "max_violation"
	}
	if session.ViolationCount >= rule.MaxViolations {
		return true, "max_violation"
	}
	if session.WarningCount >= rule.MaxWarnings {
		return true, "max_warning"
	}
	return false, ""
}

func isViolationStatus(status string) bool {
	return allowed(status, "violation", "using_phone", "sleeping", "absent")
}

func localPatrolResult(status string, reason string, raw string) modelPatrolResult {
	if !allowedPatrolStatus(status) {
		status = "uncertain"
	}
	return modelPatrolResult{
		Status:     status,
		Confidence: 1,
		Reason:     reason,
		Raw:        json.RawMessage(raw),
	}
}

func allowedPatrolStatus(status string) bool {
	return allowed(status, "normal", "suspicious", "violation", "using_phone", "sleeping", "absent", "uncertain")
}

func patrolRecordFromResult(input PatrolCheckInput, result modelPatrolResult, warningDelta int, violationDelta int) model.PatrolRecord {
	raw := string(result.Raw)
	if raw == "" {
		raw = "{}"
	}
	return model.PatrolRecord{
		SessionID:      input.SessionID,
		SceneKey:       input.SceneKey,
		TriggeredAt:    time.Now(),
		Status:         result.Status,
		Confidence:     result.Confidence,
		Reason:         result.Reason,
		ActionKey:      result.ActionKey,
		WarningDelta:   warningDelta,
		ViolationDelta: violationDelta,
		ModelRawJSON:   raw,
	}
}

func patrolRecordFromError(input PatrolCheckInput, rule model.PatrolRule, err error) model.PatrolRecord {
	code := "INTERNAL_ERROR"
	if patrolErr := (PatrolError{}); errors.As(err, &patrolErr) {
		code = patrolErr.Code
	}
	raw, _ := json.Marshal(map[string]any{
		"errorCode":        code,
		"captureErrorCode": input.CaptureErrorCode,
		"cameraEnabled":    input.CameraEnabled,
	})
	return model.PatrolRecord{
		SessionID:    input.SessionID,
		SceneKey:     input.SceneKey,
		TriggeredAt:  time.Now(),
		Status:       "uncertain",
		Confidence:   0,
		Reason:       rule.UserErrorMessage,
		ModelRawJSON: string(raw),
		ErrorCode:    code,
	}
}

func toPatrolActionView(action model.ActionConfig) PatrolActionView {
	return PatrolActionView{
		ActionKey:  action.ActionKey,
		Name:       action.Name,
		VideoURL:   action.VideoURL,
		PosterURL:  action.PosterURL,
		DurationMS: action.DurationMS,
	}
}
