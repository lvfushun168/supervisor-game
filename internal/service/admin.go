package service

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	appcrypto "supervisor-game/internal/crypto"
	"supervisor-game/internal/database"
	"supervisor-game/internal/model"

	"gorm.io/gorm"
)

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type AdminStatusInput struct {
	StartedAt    time.Time
	Addr         string
	AssetsDir    string
	DBStatus     any
	DBSource     string
	DBError      string
	BootstrapDSN bool
}

type AdminStatus struct {
	Service           string           `json:"service"`
	StartedAt         string           `json:"startedAt"`
	Addr              string           `json:"addr"`
	AssetsDir         string           `json:"assetsDir"`
	AssetsAvailable   bool             `json:"assetsAvailable"`
	Database          any              `json:"database"`
	DBSource          string           `json:"dbSource"`
	DBError           string           `json:"dbError"`
	RestartRequired   bool             `json:"restartRequired"`
	EnabledSceneCount int64            `json:"enabledSceneCount"`
	ModelConfig       AdminModelView   `json:"modelConfig"`
	PatrolRule        model.PatrolRule `json:"patrolRule"`
}

type AdminRuntimeConfig struct {
	RuntimeConfig RuntimeConfig `json:"runtimeConfig"`
	Diagnostics   any           `json:"diagnostics"`
}

type AdminModelView struct {
	model.ModelConfig
	HasAPIKey    bool   `json:"hasApiKey"`
	APIKeyMasked string `json:"apiKeyMasked"`
}

type AdminMySQLView struct {
	model.MySQLConfig
	HasPassword    bool   `json:"hasPassword"`
	PasswordMasked string `json:"passwordMasked"`
}

type ModelConfigInput struct {
	Name               string   `json:"name"`
	Provider           string   `json:"provider"`
	Enabled            *bool    `json:"enabled"`
	BaseURL            string   `json:"baseUrl"`
	APIKey             string   `json:"apiKey"`
	Model              string   `json:"model"`
	TimeoutMS          int      `json:"timeoutMs"`
	MaxImageWidth      int      `json:"maxImageWidth"`
	Temperature        *float64 `json:"temperature"`
	Prompt             string   `json:"prompt"`
	ResponseSchemaJSON string   `json:"responseSchemaJson"`
	RetryCount         int      `json:"retryCount"`
}

type MySQLConfigInput struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"databaseName"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Charset      string `json:"charset"`
	Timezone     string `json:"timezone"`
	MaxOpenConns int    `json:"maxOpenConns"`
	MaxIdleConns int    `json:"maxIdleConns"`
	Enabled      *bool  `json:"enabled"`
}

func (s *Service) AdminStatus(input AdminStatusInput) (AdminStatus, error) {
	if s.repo == nil || s.repo.DB() == nil {
		return AdminStatus{
			Service:         "supervisor-game",
			StartedAt:       input.StartedAt.Format(time.RFC3339),
			Addr:            input.Addr,
			AssetsDir:       input.AssetsDir,
			AssetsAvailable: pathExists(input.AssetsDir),
			Database:        input.DBStatus,
			DBSource:        input.DBSource,
			DBError:         input.DBError,
		}, nil
	}
	var enabledScenes int64
	_ = s.repo.DB().Model(&model.Scene{}).Where("enabled = ?", true).Count(&enabledScenes).Error
	modelConfig, _ := s.AdminModelConfig()
	patrolRule, _ := s.AdminPatrolRule()
	return AdminStatus{
		Service:           "supervisor-game",
		StartedAt:         input.StartedAt.Format(time.RFC3339),
		Addr:              input.Addr,
		AssetsDir:         input.AssetsDir,
		AssetsAvailable:   pathExists(input.AssetsDir),
		Database:          input.DBStatus,
		DBSource:          input.DBSource,
		DBError:           input.DBError,
		RestartRequired:   s.mysqlRestartRequired(input.DBSource),
		EnabledSceneCount: enabledScenes,
		ModelConfig:       modelConfig,
		PatrolRule:        patrolRule,
	}, nil
}

func (s *Service) AdminRuntimeConfig(input AdminStatusInput) (AdminRuntimeConfig, error) {
	runtimeConfig, err := s.RuntimeConfig()
	if err != nil {
		return AdminRuntimeConfig{}, err
	}
	status, err := s.AdminStatus(input)
	if err != nil {
		return AdminRuntimeConfig{}, err
	}
	return AdminRuntimeConfig{RuntimeConfig: runtimeConfig, Diagnostics: status}, nil
}

func (s *Service) AdminCharacters() ([]model.Character, error) {
	var items []model.Character
	return items, s.repo.DB().Order("id ASC").Find(&items).Error
}

func (s *Service) CreateCharacter(input model.Character) (model.Character, error) {
	input.ID = 0
	if input.ProfileJSON == "" {
		input.ProfileJSON = emptyJSON
	}
	if input.MetadataJSON == "" {
		input.MetadataJSON = emptyJSON
	}
	if err := s.validateCharacter(input); err != nil {
		return model.Character{}, err
	}
	return input, s.repo.DB().Create(&input).Error
}

func (s *Service) UpdateCharacter(id uint, input model.Character) (model.Character, error) {
	var existing model.Character
	if err := s.repo.DB().First(&existing, id).Error; err != nil {
		return model.Character{}, err
	}
	input.ID = existing.ID
	input.CreatedAt = existing.CreatedAt
	if input.ProfileJSON == "" {
		input.ProfileJSON = emptyJSON
	}
	if input.MetadataJSON == "" {
		input.MetadataJSON = emptyJSON
	}
	if err := s.validateCharacter(input); err != nil {
		return model.Character{}, err
	}
	return input, s.repo.DB().Save(&input).Error
}

func (s *Service) DeleteCharacter(id uint) error {
	return s.repo.DB().Delete(&model.Character{}, id).Error
}

func (s *Service) AdminScenes() ([]model.Scene, error) {
	var items []model.Scene
	return items, s.repo.DB().Order("id ASC").Find(&items).Error
}

func (s *Service) CreateScene(input model.Scene) (model.Scene, error) {
	input.ID = 0
	defaultSceneJSON(&input)
	if err := s.validateScene(input); err != nil {
		return model.Scene{}, err
	}
	return input, s.repo.DB().Create(&input).Error
}

func (s *Service) UpdateScene(id uint, input model.Scene) (model.Scene, error) {
	var existing model.Scene
	if err := s.repo.DB().First(&existing, id).Error; err != nil {
		return model.Scene{}, err
	}
	input.ID = existing.ID
	input.CreatedAt = existing.CreatedAt
	defaultSceneJSON(&input)
	if err := s.validateScene(input); err != nil {
		return model.Scene{}, err
	}
	return input, s.repo.DB().Save(&input).Error
}

func (s *Service) DeleteScene(id uint) error {
	var scene model.Scene
	if err := s.repo.DB().First(&scene, id).Error; err != nil {
		return err
	}
	var count int64
	if err := s.repo.DB().Model(&model.Character{}).Where("default_scene_key = ?", scene.SceneKey).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: scene is referenced by characters", ErrInvalidInput)
	}
	if err := s.repo.DB().Model(&model.ActionConfig{}).Where("scene_key = ?", scene.SceneKey).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: scene has actions; delete actions first", ErrInvalidInput)
	}
	return s.repo.DB().Delete(&scene).Error
}

func (s *Service) AdminActions(sceneKey string) ([]model.ActionConfig, error) {
	var items []model.ActionConfig
	query := s.repo.DB().Order("scene_key ASC, priority ASC, id ASC")
	if sceneKey != "" {
		query = query.Where("scene_key = ?", sceneKey)
	}
	return items, query.Find(&items).Error
}

func (s *Service) CreateAction(input model.ActionConfig) (model.ActionConfig, error) {
	input.ID = 0
	defaultActionJSON(&input)
	if err := s.validateAction(input); err != nil {
		return model.ActionConfig{}, err
	}
	return input, s.repo.DB().Create(&input).Error
}

func (s *Service) UpdateAction(id uint, input model.ActionConfig) (model.ActionConfig, error) {
	var existing model.ActionConfig
	if err := s.repo.DB().First(&existing, id).Error; err != nil {
		return model.ActionConfig{}, err
	}
	input.ID = existing.ID
	input.CreatedAt = existing.CreatedAt
	defaultActionJSON(&input)
	if err := s.validateAction(input); err != nil {
		return model.ActionConfig{}, err
	}
	return input, s.repo.DB().Save(&input).Error
}

func (s *Service) DeleteAction(id uint) error {
	var action model.ActionConfig
	if err := s.repo.DB().First(&action, id).Error; err != nil {
		return err
	}
	var scene model.Scene
	err := s.repo.DB().Where("scene_key = ?", action.SceneKey).First(&scene).Error
	if err == nil && scene.Enabled && scene.DefaultActionKey == action.ActionKey {
		return fmt.Errorf("%w: action is the default action of an enabled scene", ErrInvalidInput)
	}
	return s.repo.DB().Delete(&action).Error
}

func (s *Service) AdminModelConfig() (AdminModelView, error) {
	var item model.ModelConfig
	err := s.repo.DB().Order("id ASC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item = DefaultModelConfig()
		if err := s.repo.DB().Create(&item).Error; err != nil {
			return AdminModelView{}, err
		}
	} else if err != nil {
		return AdminModelView{}, err
	}
	return maskModel(item), nil
}

func (s *Service) UpdateModelConfig(input ModelConfigInput) (AdminModelView, error) {
	var existing model.ModelConfig
	err := s.repo.DB().Order("id ASC").First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		existing = DefaultModelConfig()
	} else if err != nil {
		return AdminModelView{}, err
	}
	if input.Name != "" {
		existing.Name = input.Name
	}
	if input.Provider != "" {
		existing.Provider = input.Provider
	}
	if input.Enabled != nil {
		existing.Enabled = *input.Enabled
	}
	existing.BaseURL = input.BaseURL
	if input.APIKey != "" {
		encrypted, err := appcrypto.EncryptString(s.cfg.ConfigEncryptionKey, input.APIKey)
		if err != nil {
			return AdminModelView{}, fmt.Errorf("%w: CONFIG_ENCRYPTION_KEY is required to save apiKey", ErrInvalidInput)
		}
		existing.APIKeyEncrypted = encrypted
	}
	existing.Model = input.Model
	if input.TimeoutMS > 0 {
		existing.TimeoutMS = input.TimeoutMS
	}
	if input.MaxImageWidth > 0 {
		existing.MaxImageWidth = input.MaxImageWidth
	}
	if input.Temperature != nil {
		existing.Temperature = *input.Temperature
	}
	existing.Prompt = input.Prompt
	if input.ResponseSchemaJSON != "" {
		existing.ResponseSchemaJSON = input.ResponseSchemaJSON
	}
	if input.RetryCount >= 0 {
		existing.RetryCount = input.RetryCount
	}
	if err := s.validateModelConfig(existing); err != nil {
		return AdminModelView{}, err
	}
	if existing.ID == 0 {
		err = s.repo.DB().Create(&existing).Error
	} else {
		err = s.repo.DB().Save(&existing).Error
	}
	if err != nil {
		return AdminModelView{}, err
	}
	return maskModel(existing), nil
}

func (s *Service) TestModelConfig() any {
	return map[string]string{
		"status":  "todo",
		"message": "TODO: M4 接入真实视觉模型测试，M2 仅保留测试入口。",
	}
}

func (s *Service) AdminPatrolRule() (model.PatrolRule, error) {
	var item model.PatrolRule
	err := s.repo.DB().Order("id ASC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item = DefaultPatrolRule()
		err = s.repo.DB().Create(&item).Error
	}
	return item, err
}

func (s *Service) UpdatePatrolRule(input model.PatrolRule) (model.PatrolRule, error) {
	existing, err := s.AdminPatrolRule()
	if err != nil {
		return model.PatrolRule{}, err
	}
	input.ID = existing.ID
	input.CreatedAt = existing.CreatedAt
	if err := ValidatePatrolRule(input); err != nil {
		return model.PatrolRule{}, err
	}
	return input, s.repo.DB().Save(&input).Error
}

func (s *Service) AdminMySQLConfig() (AdminMySQLView, error) {
	var item model.MySQLConfig
	err := s.repo.DB().Where("enabled = ?", true).Order("updated_at DESC, id DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item = model.MySQLConfig{
			Host:         "127.0.0.1",
			Port:         3306,
			DatabaseName: "supervisor_game",
			Username:     "supervisor",
			Charset:      "utf8mb4",
			Timezone:     "Local",
			MaxOpenConns: 20,
			MaxIdleConns: 5,
			Enabled:      false,
		}
	} else if err != nil {
		return AdminMySQLView{}, err
	}
	return maskMySQL(item), nil
}

func (s *Service) UpdateMySQLConfig(input MySQLConfigInput) (AdminMySQLView, error) {
	var existing model.MySQLConfig
	err := s.repo.DB().Where("enabled = ?", true).Order("updated_at DESC, id DESC").First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		existing = model.MySQLConfig{}
	} else if err != nil {
		return AdminMySQLView{}, err
	}
	applyMySQLInput(&existing, input)
	if input.Password != "" {
		encrypted, err := appcrypto.EncryptString(s.cfg.ConfigEncryptionKey, input.Password)
		if err != nil {
			return AdminMySQLView{}, fmt.Errorf("%w: CONFIG_ENCRYPTION_KEY is required to save password", ErrInvalidInput)
		}
		existing.PasswordEncrypted = encrypted
	}
	if err := ValidateMySQLConfig(existing); err != nil {
		return AdminMySQLView{}, err
	}
	if existing.Enabled {
		if err := s.repo.DB().Model(&model.MySQLConfig{}).Where("id <> ?", existing.ID).Update("enabled", false).Error; err != nil {
			return AdminMySQLView{}, err
		}
	}
	if existing.ID == 0 {
		err = s.repo.DB().Create(&existing).Error
	} else {
		err = s.repo.DB().Save(&existing).Error
	}
	if err != nil {
		return AdminMySQLView{}, err
	}
	return maskMySQL(existing), nil
}

func (s *Service) TestMySQLConfig(input MySQLConfigInput) (AdminMySQLView, error) {
	config := model.MySQLConfig{}
	applyMySQLInput(&config, input)
	if input.Password != "" {
		encrypted, err := appcrypto.EncryptString(s.cfg.ConfigEncryptionKey, input.Password)
		if err != nil {
			return AdminMySQLView{}, fmt.Errorf("%w: CONFIG_ENCRYPTION_KEY is required to test password", ErrInvalidInput)
		}
		config.PasswordEncrypted = encrypted
	}
	if err := ValidateMySQLConfig(config); err != nil {
		return AdminMySQLView{}, err
	}
	now := time.Now()
	config.LastTestedAt = &now
	err := database.TestMySQLConfig(config, s.cfg.ConfigEncryptionKey)
	if err != nil {
		config.LastTestResult = "failed"
		config.LastTestError = err.Error()
	} else {
		config.LastTestResult = "success"
		config.LastTestError = ""
	}
	var existing model.MySQLConfig
	findErr := s.repo.DB().Where("enabled = ?", true).Order("updated_at DESC, id DESC").First(&existing).Error
	if errors.Is(findErr, gorm.ErrRecordNotFound) {
		findErr = s.repo.DB().Order("updated_at DESC, id DESC").First(&existing).Error
	}
	if findErr == nil {
		existing.LastTestedAt = config.LastTestedAt
		existing.LastTestResult = config.LastTestResult
		existing.LastTestError = config.LastTestError
		_ = s.repo.DB().Save(&existing).Error
		return maskMySQL(existing), err
	}
	return maskMySQL(config), err
}

func (s *Service) MigrateCurrentDB() error {
	return database.Migrate(s.repo.DB())
}

func (s *Service) validateCharacter(item model.Character) error {
	if !validKey(item.CharacterKey) {
		return fmt.Errorf("%w: characterKey must contain only letters, numbers, and underscores", ErrInvalidInput)
	}
	if item.Enabled && item.Name == "" {
		return fmt.Errorf("%w: enabled character requires name", ErrInvalidInput)
	}
	if err := ValidateJSONFields(map[string]string{"profileJson": item.ProfileJSON, "metadataJson": item.MetadataJSON}); err != nil {
		return err
	}
	if item.DefaultSceneKey != "" {
		var count int64
		if err := s.repo.DB().Model(&model.Scene{}).Where("scene_key = ?", item.DefaultSceneKey).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("%w: defaultSceneKey must reference an existing scene", ErrInvalidInput)
		}
	}
	return nil
}

func (s *Service) validateScene(item model.Scene) error {
	if !validKey(item.SceneKey) {
		return fmt.Errorf("%w: sceneKey must contain only letters, numbers, and underscores", ErrInvalidInput)
	}
	if item.Name == "" {
		return fmt.Errorf("%w: scene name is required", ErrInvalidInput)
	}
	if item.BackgroundType == "" {
		return fmt.Errorf("%w: backgroundType is required", ErrInvalidInput)
	}
	if item.Enabled && item.BackgroundURL == "" && item.BackgroundVideoURL == "" {
		return fmt.Errorf("%w: enabled scene requires backgroundUrl or backgroundVideoUrl", ErrInvalidInput)
	}
	if err := ValidateJSONFields(map[string]string{
		"availableActionKeysJson":  item.AvailableActionKeysJSON,
		"modelResultActionMapJson": item.ModelResultActionMapJSON,
		"metadataJson":             item.MetadataJSON,
	}); err != nil {
		return err
	}
	if item.Enabled {
		var count int64
		if err := s.repo.DB().Model(&model.ActionConfig{}).Where("scene_key = ? AND enabled = ?", item.SceneKey, true).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("%w: enabled scene requires at least one enabled action", ErrInvalidInput)
		}
	}
	return nil
}

func (s *Service) validateAction(item model.ActionConfig) error {
	if !validKey(item.SceneKey) || !validKey(item.ActionKey) {
		return fmt.Errorf("%w: sceneKey and actionKey must contain only letters, numbers, and underscores", ErrInvalidInput)
	}
	if item.Name == "" {
		return fmt.Errorf("%w: action name is required", ErrInvalidInput)
	}
	if item.Enabled && item.VideoURL == "" {
		return fmt.Errorf("%w: enabled action requires videoUrl", ErrInvalidInput)
	}
	if item.DurationMS <= 0 {
		return fmt.Errorf("%w: durationMs must be greater than 0", ErrInvalidInput)
	}
	var sceneCount int64
	if err := s.repo.DB().Model(&model.Scene{}).Where("scene_key = ?", item.SceneKey).Count(&sceneCount).Error; err != nil {
		return err
	}
	if sceneCount == 0 {
		return fmt.Errorf("%w: sceneKey must reference an existing scene", ErrInvalidInput)
	}
	if err := ValidateJSONFields(map[string]string{
		"modelResultMapJson": item.ModelResultMapJSON,
		"localRuleMapJson":   item.LocalRuleMapJSON,
		"metadataJson":       item.MetadataJSON,
	}); err != nil {
		return err
	}
	return nil
}

func (s *Service) validateModelConfig(item model.ModelConfig) error {
	if item.Name == "" || item.Provider == "" {
		return fmt.Errorf("%w: name and provider are required", ErrInvalidInput)
	}
	if item.TimeoutMS <= 0 || item.MaxImageWidth <= 0 {
		return fmt.Errorf("%w: timeoutMs and maxImageWidth must be greater than 0", ErrInvalidInput)
	}
	if item.Temperature < 0 || item.Temperature > 2 {
		return fmt.Errorf("%w: temperature must be between 0 and 2", ErrInvalidInput)
	}
	if item.RetryCount < 0 {
		return fmt.Errorf("%w: retryCount must be greater than or equal to 0", ErrInvalidInput)
	}
	return validateJSON("responseSchemaJson", item.ResponseSchemaJSON)
}

func ValidatePatrolRule(rule model.PatrolRule) error {
	if err := validateRange("slow", rule.SlowMinSeconds, rule.SlowMaxSeconds); err != nil {
		return err
	}
	if err := validateRange("normal", rule.NormalMinSeconds, rule.NormalMaxSeconds); err != nil {
		return err
	}
	if err := validateRange("high", rule.HighMinSeconds, rule.HighMaxSeconds); err != nil {
		return err
	}
	if rule.MaxWarnings <= 0 || rule.MaxViolations <= 0 {
		return fmt.Errorf("%w: maxWarnings and maxViolations must be greater than 0", ErrInvalidInput)
	}
	if !allowed(rule.CameraOffStrategy, "normal", "suspicious", "violation", "uncertain") {
		return fmt.Errorf("%w: cameraOffStrategy is invalid", ErrInvalidInput)
	}
	if !allowed(rule.CaptureFailedStrategy, "normal", "suspicious", "violation", "uncertain") {
		return fmt.Errorf("%w: captureFailedStrategy is invalid", ErrInvalidInput)
	}
	if rule.ModelTimeoutRetryCount < 0 {
		return fmt.Errorf("%w: modelTimeoutRetryCount must be greater than or equal to 0", ErrInvalidInput)
	}
	return nil
}

func ValidateMySQLConfig(config model.MySQLConfig) error {
	if config.Host == "" || config.Port <= 0 || config.DatabaseName == "" || config.Username == "" {
		return fmt.Errorf("%w: host, port, databaseName, and username are required", ErrInvalidInput)
	}
	if config.Charset == "" {
		return fmt.Errorf("%w: charset is required", ErrInvalidInput)
	}
	if config.Timezone == "" {
		return fmt.Errorf("%w: timezone is required", ErrInvalidInput)
	}
	if config.MaxOpenConns <= 0 || config.MaxIdleConns <= 0 {
		return fmt.Errorf("%w: maxOpenConns and maxIdleConns must be greater than 0", ErrInvalidInput)
	}
	return nil
}

func validateRange(name string, min int, max int) error {
	if min <= 0 {
		return fmt.Errorf("%w: %s minSeconds must be greater than 0", ErrInvalidInput, name)
	}
	if max < min {
		return fmt.Errorf("%w: %s maxSeconds must be greater than or equal to minSeconds", ErrInvalidInput, name)
	}
	return nil
}

func validKey(value string) bool {
	return value != "" && keyPattern.MatchString(value)
}

func defaultSceneJSON(item *model.Scene) {
	if item.AvailableActionKeysJSON == "" {
		item.AvailableActionKeysJSON = "[]"
	}
	if item.ModelResultActionMapJSON == "" {
		item.ModelResultActionMapJSON = emptyJSON
	}
	if item.MetadataJSON == "" {
		item.MetadataJSON = emptyJSON
	}
}

func defaultActionJSON(item *model.ActionConfig) {
	if item.ModelResultMapJSON == "" {
		item.ModelResultMapJSON = emptyJSON
	}
	if item.LocalRuleMapJSON == "" {
		item.LocalRuleMapJSON = emptyJSON
	}
	if item.MetadataJSON == "" {
		item.MetadataJSON = emptyJSON
	}
}

func maskModel(item model.ModelConfig) AdminModelView {
	view := AdminModelView{ModelConfig: item}
	view.APIKeyEncrypted = ""
	view.HasAPIKey = item.APIKeyEncrypted != ""
	if view.HasAPIKey {
		view.APIKeyMasked = "********"
	}
	return view
}

func maskMySQL(item model.MySQLConfig) AdminMySQLView {
	view := AdminMySQLView{MySQLConfig: item}
	view.PasswordEncrypted = ""
	view.HasPassword = item.PasswordEncrypted != ""
	if view.HasPassword {
		view.PasswordMasked = "********"
	}
	return view
}

func applyMySQLInput(config *model.MySQLConfig, input MySQLConfigInput) {
	config.Host = input.Host
	config.Port = input.Port
	config.DatabaseName = input.DatabaseName
	config.Username = input.Username
	config.Charset = input.Charset
	config.Timezone = input.Timezone
	config.MaxOpenConns = input.MaxOpenConns
	config.MaxIdleConns = input.MaxIdleConns
	if input.Enabled != nil {
		config.Enabled = *input.Enabled
	}
}

func (s *Service) mysqlRestartRequired(dbSource string) bool {
	var item model.MySQLConfig
	err := s.repo.DB().Where("enabled = ?", true).Order("updated_at DESC, id DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	if err != nil {
		return false
	}
	return dbSource != "mysql_configs"
}

func pathExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
