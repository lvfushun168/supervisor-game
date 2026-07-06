package model

import "time"

type AppSetting struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	SettingKey       string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"settingKey"`
	SettingValueJSON string    `gorm:"type:text;not null" json:"settingValueJson"`
	Description      string    `gorm:"type:text" json:"description"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type Character struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CharacterKey    string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"characterKey"`
	Name            string    `gorm:"type:varchar(128);not null" json:"name"`
	Enabled         bool      `gorm:"not null;default:true" json:"enabled"`
	Description     string    `gorm:"type:text" json:"description"`
	AvatarURL       string    `gorm:"type:varchar(512)" json:"avatarUrl"`
	ProfileJSON     string    `gorm:"type:text;not null" json:"profileJson"`
	VoiceStyle      string    `gorm:"type:varchar(128)" json:"voiceStyle"`
	DefaultSceneKey string    `gorm:"type:varchar(128);index" json:"defaultSceneKey"`
	MetadataJSON    string    `gorm:"type:text;not null" json:"metadataJson"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type UserSetting struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	Mode                  string    `gorm:"type:varchar(32);not null" json:"mode"`
	CustomDurationSeconds int       `gorm:"not null" json:"customDurationSeconds"`
	PatrolFrequency       string    `gorm:"type:varchar(32);not null" json:"patrolFrequency"`
	BackgroundAudioKey    string    `gorm:"type:varchar(128)" json:"backgroundAudioKey"`
	BackgroundVolume      float64   `gorm:"not null" json:"backgroundVolume"`
	ActionVolume          float64   `gorm:"not null" json:"actionVolume"`
	UIVolume              float64   `gorm:"not null" json:"uiVolume"`
	QuietPatrolEnabled    bool      `gorm:"not null" json:"quietPatrolEnabled"`
	ScreenFilter          string    `gorm:"type:varchar(32);not null" json:"screenFilter"`
	CameraEnabled         bool      `gorm:"not null" json:"cameraEnabled"`
	CameraDeviceID        string    `gorm:"type:varchar(256)" json:"cameraDeviceId"`
	MetadataJSON          string    `gorm:"type:text;not null" json:"metadataJson"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type Scene struct {
	ID                       uint      `gorm:"primaryKey" json:"id"`
	SceneKey                 string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"sceneKey"`
	Name                     string    `gorm:"type:varchar(128);not null" json:"name"`
	Enabled                  bool      `gorm:"not null;default:true;index" json:"enabled"`
	Description              string    `gorm:"type:text" json:"description"`
	BackgroundType           string    `gorm:"type:varchar(32);not null" json:"backgroundType"`
	BackgroundURL            string    `gorm:"type:varchar(512)" json:"backgroundUrl"`
	BackgroundVideoURL       string    `gorm:"type:varchar(512)" json:"backgroundVideoUrl"`
	BackgroundPosterURL      string    `gorm:"type:varchar(512)" json:"backgroundPosterUrl"`
	AmbientAudioURL          string    `gorm:"type:varchar(512)" json:"ambientAudioUrl"`
	DefaultActionKey         string    `gorm:"type:varchar(128)" json:"defaultActionKey"`
	AvailableActionKeysJSON  string    `gorm:"type:text;not null" json:"availableActionKeysJson"`
	ModelResultActionMapJSON string    `gorm:"type:text;not null" json:"modelResultActionMapJson"`
	MetadataJSON             string    `gorm:"type:text;not null" json:"metadataJson"`
	CreatedAt                time.Time `json:"createdAt"`
	UpdatedAt                time.Time `json:"updatedAt"`
}

type ActionConfig struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	SceneKey           string    `gorm:"type:varchar(128);uniqueIndex:idx_scene_action;not null;index" json:"sceneKey"`
	ActionKey          string    `gorm:"type:varchar(128);uniqueIndex:idx_scene_action;not null" json:"actionKey"`
	Name               string    `gorm:"type:varchar(128);not null" json:"name"`
	Enabled            bool      `gorm:"not null;default:true;index" json:"enabled"`
	Priority           int       `gorm:"not null;default:0" json:"priority"`
	VideoURL           string    `gorm:"type:varchar(512);not null" json:"videoUrl"`
	PosterURL          string    `gorm:"type:varchar(512)" json:"posterUrl"`
	DurationMS         int       `gorm:"not null;default:0" json:"durationMs"`
	NextActionKey      string    `gorm:"type:varchar(128)" json:"nextActionKey"`
	ModelResultMapJSON string    `gorm:"type:text;not null" json:"modelResultMapJson"`
	LocalRuleMapJSON   string    `gorm:"type:text;not null" json:"localRuleMapJson"`
	MetadataJSON       string    `gorm:"type:text;not null" json:"metadataJson"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type ModelConfig struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Name               string    `gorm:"type:varchar(128);not null" json:"name"`
	Provider           string    `gorm:"type:varchar(64);not null" json:"provider"`
	Enabled            bool      `gorm:"not null;default:false;index" json:"enabled"`
	BaseURL            string    `gorm:"type:varchar(512)" json:"baseUrl"`
	APIKeyEncrypted    string    `gorm:"type:text" json:"-"`
	Model              string    `gorm:"type:varchar(128)" json:"model"`
	TimeoutMS          int       `gorm:"not null;default:30000" json:"timeoutMs"`
	MaxImageWidth      int       `gorm:"not null;default:1024" json:"maxImageWidth"`
	Temperature        float64   `gorm:"not null;default:0" json:"temperature"`
	Prompt             string    `gorm:"type:text" json:"prompt"`
	ResponseSchemaJSON string    `gorm:"type:text;not null" json:"responseSchemaJson"`
	RetryCount         int       `gorm:"not null;default:1" json:"retryCount"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type PatrolRule struct {
	ID                     uint      `gorm:"primaryKey" json:"id"`
	SlowMinSeconds         int       `gorm:"not null" json:"slowMinSeconds"`
	SlowMaxSeconds         int       `gorm:"not null" json:"slowMaxSeconds"`
	NormalMinSeconds       int       `gorm:"not null" json:"normalMinSeconds"`
	NormalMaxSeconds       int       `gorm:"not null" json:"normalMaxSeconds"`
	HighMinSeconds         int       `gorm:"not null" json:"highMinSeconds"`
	HighMaxSeconds         int       `gorm:"not null" json:"highMaxSeconds"`
	MaxWarnings            int       `gorm:"not null" json:"maxWarnings"`
	MaxViolations          int       `gorm:"not null" json:"maxViolations"`
	SuspiciousAddsWarning  bool      `gorm:"not null" json:"suspiciousAddsWarning"`
	ViolationDirectFail    bool      `gorm:"not null" json:"violationDirectFail"`
	CameraOffStrategy      string    `gorm:"type:varchar(32);not null" json:"cameraOffStrategy"`
	CaptureFailedStrategy  string    `gorm:"type:varchar(32);not null" json:"captureFailedStrategy"`
	ModelTimeoutRetryCount int       `gorm:"not null" json:"modelTimeoutRetryCount"`
	UserErrorMessage       string    `gorm:"type:text" json:"userErrorMessage"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type MySQLConfig struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Host              string     `gorm:"type:varchar(255);not null" json:"host"`
	Port              int        `gorm:"not null" json:"port"`
	DatabaseName      string     `gorm:"type:varchar(128);not null" json:"databaseName"`
	Username          string     `gorm:"type:varchar(128);not null" json:"username"`
	PasswordEncrypted string     `gorm:"type:text" json:"-"`
	Charset           string     `gorm:"type:varchar(32);not null" json:"charset"`
	Timezone          string     `gorm:"type:varchar(64);not null" json:"timezone"`
	MaxOpenConns      int        `gorm:"not null" json:"maxOpenConns"`
	MaxIdleConns      int        `gorm:"not null" json:"maxIdleConns"`
	Enabled           bool       `gorm:"not null;default:false;index" json:"enabled"`
	LastTestedAt      *time.Time `json:"lastTestedAt"`
	LastTestResult    string     `gorm:"type:varchar(32)" json:"lastTestResult"`
	LastTestError     string     `gorm:"type:text" json:"lastTestError"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func (MySQLConfig) TableName() string {
	return "mysql_configs"
}

type WorkSession struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	SceneKey               string     `gorm:"type:varchar(128);not null;index" json:"sceneKey"`
	Mode                   string     `gorm:"type:varchar(32);not null" json:"mode"`
	PlannedDurationSeconds int        `gorm:"not null" json:"plannedDurationSeconds"`
	StartedAt              time.Time  `gorm:"not null;index" json:"startedAt"`
	EndedAt                *time.Time `json:"endedAt"`
	ActualFocusSeconds     int        `gorm:"not null;default:0" json:"actualFocusSeconds"`
	PatrolCount            int        `gorm:"not null;default:0" json:"patrolCount"`
	WarningCount           int        `gorm:"not null;default:0" json:"warningCount"`
	ViolationCount         int        `gorm:"not null;default:0" json:"violationCount"`
	Result                 string     `gorm:"type:varchar(32);index" json:"result"`
	FinishReason           string     `gorm:"type:varchar(32);index" json:"finishReason"`
	EarnedCurrency         int        `gorm:"not null;default:0" json:"earnedCurrency"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
}

type PatrolRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SessionID      uint      `gorm:"not null;index" json:"sessionId"`
	SceneKey       string    `gorm:"type:varchar(128);not null;index" json:"sceneKey"`
	TriggeredAt    time.Time `gorm:"not null;index" json:"triggeredAt"`
	Status         string    `gorm:"type:varchar(32);not null;index" json:"status"`
	Confidence     float64   `gorm:"not null;default:0" json:"confidence"`
	Reason         string    `gorm:"type:text" json:"reason"`
	ActionKey      string    `gorm:"type:varchar(128)" json:"actionKey"`
	WarningDelta   int       `gorm:"not null;default:0" json:"warningDelta"`
	ViolationDelta int       `gorm:"not null;default:0" json:"violationDelta"`
	ModelRawJSON   string    `gorm:"type:text;not null" json:"modelRawJson"`
	ErrorCode      string    `gorm:"type:varchar(64)" json:"errorCode"`
	CreatedAt      time.Time `json:"createdAt"`
}

type DailyStat struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	StatDate       time.Time `gorm:"type:date;uniqueIndex;not null" json:"statDate"`
	FocusSeconds   int       `gorm:"not null;default:0" json:"focusSeconds"`
	SessionCount   int       `gorm:"not null;default:0" json:"sessionCount"`
	PatrolCount    int       `gorm:"not null;default:0" json:"patrolCount"`
	WarningCount   int       `gorm:"not null;default:0" json:"warningCount"`
	ViolationCount int       `gorm:"not null;default:0" json:"violationCount"`
	EarnedCurrency int       `gorm:"not null;default:0" json:"earnedCurrency"`
	LastResult     string    `gorm:"type:varchar(32)" json:"lastResult"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Task struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TaskKey        string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"taskKey"`
	Name           string    `gorm:"type:varchar(128);not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	TaskType       string    `gorm:"type:varchar(64);not null" json:"taskType"`
	TargetValue    int       `gorm:"not null" json:"targetValue"`
	RewardCurrency int       `gorm:"not null" json:"rewardCurrency"`
	Enabled        bool      `gorm:"not null;default:true;index" json:"enabled"`
	SortOrder      int       `gorm:"not null;default:0" json:"sortOrder"`
	MetadataJSON   string    `gorm:"type:text;not null" json:"metadataJson"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type TaskRecord struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	TaskKey       string     `gorm:"type:varchar(128);uniqueIndex:idx_task_record_date;not null" json:"taskKey"`
	RecordDate    time.Time  `gorm:"type:date;uniqueIndex:idx_task_record_date;not null" json:"recordDate"`
	ProgressValue int        `gorm:"not null;default:0" json:"progressValue"`
	Status        string     `gorm:"type:varchar(32);not null" json:"status"`
	ClaimedAt     *time.Time `json:"claimedAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type Badge struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	BadgeKey     string     `gorm:"type:varchar(128);uniqueIndex;not null" json:"badgeKey"`
	Name         string     `gorm:"type:varchar(128);not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	Enabled      bool       `gorm:"not null;default:true;index" json:"enabled"`
	Unlocked     bool       `gorm:"not null;default:false;index" json:"unlocked"`
	UnlockedAt   *time.Time `json:"unlockedAt"`
	MetadataJSON string     `gorm:"type:text;not null" json:"metadataJson"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type UserProgress struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Level             int        `gorm:"not null;default:1" json:"level"`
	TotalFocusSeconds int        `gorm:"not null;default:0" json:"totalFocusSeconds"`
	Currency          int        `gorm:"not null;default:0" json:"currency"`
	CurrentStreakDays int        `gorm:"not null;default:0" json:"currentStreakDays"`
	LongestStreakDays int        `gorm:"not null;default:0" json:"longestStreakDays"`
	LastFocusDate     *time.Time `gorm:"type:date" json:"lastFocusDate"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func (UserProgress) TableName() string {
	return "user_progress"
}
