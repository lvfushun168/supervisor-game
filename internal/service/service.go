package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"supervisor-game/internal/config"
	"supervisor-game/internal/model"
	"supervisor-game/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrDatabaseUnavailable = errors.New("database unavailable")
	ErrInvalidInput        = errors.New("invalid input")
)

type Service struct {
	cfg  config.Config
	repo *repository.Repository
	now  func() time.Time
}

type UserSettingInput struct {
	Mode                  string   `json:"mode"`
	CustomDurationSeconds *int     `json:"customDurationSeconds"`
	PatrolFrequency       string   `json:"patrolFrequency"`
	BackgroundAudioKey    string   `json:"backgroundAudioKey"`
	BackgroundVolume      *float64 `json:"backgroundVolume"`
	ActionVolume          *float64 `json:"actionVolume"`
	UIVolume              *float64 `json:"uiVolume"`
	QuietPatrolEnabled    *bool    `json:"quietPatrolEnabled"`
	ScreenFilter          string   `json:"screenFilter"`
	CameraEnabled         *bool    `json:"cameraEnabled"`
	CameraDeviceID        string   `json:"cameraDeviceId"`
	MetadataJSON          string   `json:"metadataJson"`
}

type RuntimeConfig struct {
	App         RuntimeApp         `json:"app"`
	PatrolRule  RuntimePatrolRule  `json:"patrolRule"`
	Character   RuntimeCharacter   `json:"character"`
	UserSetting RuntimeUserSetting `json:"userSetting"`
}

type RuntimeApp struct {
	Env           string `json:"env"`
	AssetsBaseURL string `json:"assetsBaseUrl"`
	ServerTime    string `json:"serverTime"`
}

type RuntimePatrolRule struct {
	Slow                  RuntimePatrolRange `json:"slow"`
	Normal                RuntimePatrolRange `json:"normal"`
	High                  RuntimePatrolRange `json:"high"`
	MaxWarnings           int                `json:"maxWarnings"`
	MaxViolations         int                `json:"maxViolations"`
	CameraOffStrategy     string             `json:"cameraOffStrategy"`
	CaptureFailedStrategy string             `json:"captureFailedStrategy"`
	UserErrorMessage      string             `json:"userErrorMessage"`
}

type RuntimePatrolRange struct {
	MinSeconds int `json:"minSeconds"`
	MaxSeconds int `json:"maxSeconds"`
}

type RuntimeCharacter struct {
	CharacterKey string `json:"characterKey"`
	Name         string `json:"name"`
}

type RuntimeUserSetting struct {
	Mode                  string  `json:"mode"`
	CustomDurationSeconds int     `json:"customDurationSeconds"`
	PatrolFrequency       string  `json:"patrolFrequency"`
	BackgroundAudioKey    string  `json:"backgroundAudioKey"`
	BackgroundVolume      float64 `json:"backgroundVolume"`
	ActionVolume          float64 `json:"actionVolume"`
	UIVolume              float64 `json:"uiVolume"`
	QuietPatrolEnabled    bool    `json:"quietPatrolEnabled"`
	ScreenFilter          string  `json:"screenFilter"`
	CameraEnabled         bool    `json:"cameraEnabled"`
	CameraDeviceID        string  `json:"cameraDeviceId"`
}

func New(cfg config.Config, repo *repository.Repository) *Service {
	return &Service{cfg: cfg, repo: repo, now: time.Now}
}

func (s *Service) RuntimeConfig() (RuntimeConfig, error) {
	if s.repo == nil || s.repo.DB() == nil {
		return RuntimeConfig{}, ErrDatabaseUnavailable
	}

	rule, err := s.repo.LatestPatrolRule()
	if err != nil {
		return RuntimeConfig{}, err
	}
	character, err := s.repo.DefaultCharacter()
	if err != nil {
		return RuntimeConfig{}, err
	}
	setting, err := s.repo.FirstUserSetting()
	if err != nil {
		return RuntimeConfig{}, err
	}

	return RuntimeConfig{
		App: RuntimeApp{
			Env:           s.cfg.AppEnv,
			AssetsBaseURL: "/assets/",
			ServerTime:    s.now().Format(time.RFC3339),
		},
		PatrolRule: RuntimePatrolRule{
			Slow: RuntimePatrolRange{
				MinSeconds: rule.SlowMinSeconds,
				MaxSeconds: rule.SlowMaxSeconds,
			},
			Normal: RuntimePatrolRange{
				MinSeconds: rule.NormalMinSeconds,
				MaxSeconds: rule.NormalMaxSeconds,
			},
			High: RuntimePatrolRange{
				MinSeconds: rule.HighMinSeconds,
				MaxSeconds: rule.HighMaxSeconds,
			},
			MaxWarnings:           rule.MaxWarnings,
			MaxViolations:         rule.MaxViolations,
			CameraOffStrategy:     rule.CameraOffStrategy,
			CaptureFailedStrategy: rule.CaptureFailedStrategy,
			UserErrorMessage:      rule.UserErrorMessage,
		},
		Character: RuntimeCharacter{
			CharacterKey: character.CharacterKey,
			Name:         character.Name,
		},
		UserSetting: toRuntimeUserSetting(setting),
	}, nil
}

func (s *Service) EnabledScenes() ([]model.Scene, error) {
	if s.repo == nil || s.repo.DB() == nil {
		return nil, ErrDatabaseUnavailable
	}
	return s.repo.EnabledScenes()
}

func (s *Service) UserSetting() (model.UserSetting, error) {
	if s.repo == nil || s.repo.DB() == nil {
		return model.UserSetting{}, ErrDatabaseUnavailable
	}
	return s.repo.FirstUserSetting()
}

func (s *Service) UpdateUserSetting(input UserSettingInput) (model.UserSetting, error) {
	if s.repo == nil || s.repo.DB() == nil {
		return model.UserSetting{}, ErrDatabaseUnavailable
	}

	setting, err := s.repo.FirstUserSetting()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setting = DefaultUserSetting()
	} else if err != nil {
		return model.UserSetting{}, err
	}

	applyUserSettingInput(&setting, input)
	if err := ValidateUserSetting(setting); err != nil {
		return model.UserSetting{}, err
	}

	if err := s.repo.SaveUserSetting(&setting); err != nil {
		return model.UserSetting{}, err
	}
	return setting, nil
}

func (s *Service) SeedDefaults() error {
	if s.repo == nil || s.repo.DB() == nil {
		return nil
	}

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		repo := repository.New(tx)
		for _, setting := range DefaultAppSettings() {
			if err := repo.UpsertAppSetting(setting); err != nil {
				return err
			}
		}
		if err := repo.UpsertCharacter(DefaultCharacter()); err != nil {
			return err
		}
		if err := repo.UpsertScene(DefaultScene()); err != nil {
			return err
		}
		for _, action := range DefaultActions() {
			if err := repo.UpsertAction(action); err != nil {
				return err
			}
		}
		if err := repo.EnsurePatrolRule(DefaultPatrolRule()); err != nil {
			return err
		}
		if err := repo.EnsureModelConfig(DefaultModelConfig()); err != nil {
			return err
		}
		if err := repo.EnsureUserSetting(DefaultUserSetting()); err != nil {
			return err
		}
		for _, task := range DefaultTasks() {
			if err := repo.UpsertTask(task); err != nil {
				return err
			}
		}
		for _, badge := range DefaultBadges() {
			if err := repo.UpsertBadge(badge); err != nil {
				return err
			}
		}
		return repo.EnsureUserProgress(DefaultUserProgress())
	})
}

func ValidateUserSetting(setting model.UserSetting) error {
	if !allowed(setting.Mode, "pomodoro", "custom", "infinite") {
		return fmt.Errorf("%w: mode must be one of pomodoro, custom, infinite", ErrInvalidInput)
	}
	if setting.CustomDurationSeconds <= 0 {
		return fmt.Errorf("%w: customDurationSeconds must be greater than 0", ErrInvalidInput)
	}
	if !allowed(setting.PatrolFrequency, "slow", "normal", "high") {
		return fmt.Errorf("%w: patrolFrequency must be one of slow, normal, high", ErrInvalidInput)
	}
	if !allowed(setting.ScreenFilter, "normal", "grayscale", "dark") {
		return fmt.Errorf("%w: screenFilter must be one of normal, grayscale, dark", ErrInvalidInput)
	}
	if err := validateVolume("backgroundVolume", setting.BackgroundVolume); err != nil {
		return err
	}
	if err := validateVolume("actionVolume", setting.ActionVolume); err != nil {
		return err
	}
	if err := validateVolume("uiVolume", setting.UIVolume); err != nil {
		return err
	}
	if err := validateJSON("metadataJson", setting.MetadataJSON); err != nil {
		return err
	}
	return nil
}

func ValidateJSONFields(values map[string]string) error {
	for name, value := range values {
		if err := validateJSON(name, value); err != nil {
			return err
		}
	}
	return nil
}

func applyUserSettingInput(setting *model.UserSetting, input UserSettingInput) {
	if input.Mode != "" {
		setting.Mode = input.Mode
	}
	if input.CustomDurationSeconds != nil {
		setting.CustomDurationSeconds = *input.CustomDurationSeconds
	}
	if input.PatrolFrequency != "" {
		setting.PatrolFrequency = input.PatrolFrequency
	}
	if input.BackgroundAudioKey != "" {
		setting.BackgroundAudioKey = input.BackgroundAudioKey
	}
	if input.BackgroundVolume != nil {
		setting.BackgroundVolume = *input.BackgroundVolume
	}
	if input.ActionVolume != nil {
		setting.ActionVolume = *input.ActionVolume
	}
	if input.UIVolume != nil {
		setting.UIVolume = *input.UIVolume
	}
	if input.QuietPatrolEnabled != nil {
		setting.QuietPatrolEnabled = *input.QuietPatrolEnabled
	}
	if input.ScreenFilter != "" {
		setting.ScreenFilter = input.ScreenFilter
	}
	if input.CameraEnabled != nil {
		setting.CameraEnabled = *input.CameraEnabled
	}
	if input.CameraDeviceID != "" {
		setting.CameraDeviceID = input.CameraDeviceID
	}
	if input.MetadataJSON != "" {
		setting.MetadataJSON = input.MetadataJSON
	}
}

func toRuntimeUserSetting(setting model.UserSetting) RuntimeUserSetting {
	return RuntimeUserSetting{
		Mode:                  setting.Mode,
		CustomDurationSeconds: setting.CustomDurationSeconds,
		PatrolFrequency:       setting.PatrolFrequency,
		BackgroundAudioKey:    setting.BackgroundAudioKey,
		BackgroundVolume:      setting.BackgroundVolume,
		ActionVolume:          setting.ActionVolume,
		UIVolume:              setting.UIVolume,
		QuietPatrolEnabled:    setting.QuietPatrolEnabled,
		ScreenFilter:          setting.ScreenFilter,
		CameraEnabled:         setting.CameraEnabled,
		CameraDeviceID:        setting.CameraDeviceID,
	}
}

func validateVolume(name string, value float64) error {
	if value < 0 || value > 1 {
		return fmt.Errorf("%w: %s must be between 0 and 1", ErrInvalidInput, name)
	}
	return nil
}

func validateJSON(name string, value string) error {
	if value == "" {
		return fmt.Errorf("%w: %s must be valid JSON", ErrInvalidInput, name)
	}
	var raw any
	if err := json.Unmarshal([]byte(value), &raw); err != nil {
		return fmt.Errorf("%w: %s must be valid JSON", ErrInvalidInput, name)
	}
	return nil
}

func allowed(value string, values ...string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
