package service

import "supervisor-game/internal/model"

const emptyJSON = "{}"

func DefaultAppSettings() []model.AppSetting {
	return []model.AppSetting{
		{
			SettingKey:       "level_rules",
			SettingValueJSON: `{"levels":[{"level":1,"requiredFocusSeconds":0},{"level":2,"requiredFocusSeconds":3600},{"level":3,"requiredFocusSeconds":10800},{"level":4,"requiredFocusSeconds":21600},{"level":5,"requiredFocusSeconds":43200}]}`,
			Description:      "Default level thresholds for M1.",
		},
	}
}

func DefaultCharacter() model.Character {
	return model.Character{
		CharacterKey:    "default_oc",
		Name:            "督学员",
		Enabled:         true,
		Description:     "默认督学角色。",
		AvatarURL:       "assets/scenes/study_room.jpg",
		ProfileJSON:     `{"title":"OC 督学员","summary":"负责监督本次劳动。","traits":["严格","可靠"]}`,
		VoiceStyle:      "strict",
		DefaultSceneKey: "study_room",
		MetadataJSON:    emptyJSON,
	}
}

func DefaultUserSetting() model.UserSetting {
	return model.UserSetting{
		Mode:                  "pomodoro",
		CustomDurationSeconds: 1500,
		PatrolFrequency:       "normal",
		BackgroundAudioKey:    "library",
		BackgroundVolume:      0.4,
		ActionVolume:          0.8,
		UIVolume:              0.6,
		QuietPatrolEnabled:    false,
		ScreenFilter:          "normal",
		CameraEnabled:         true,
		CameraDeviceID:        "",
		MetadataJSON:          emptyJSON,
	}
}

func DefaultScene() model.Scene {
	return model.Scene{
		SceneKey:                 "study_room",
		Name:                     "自习室",
		Enabled:                  true,
		Description:              "默认自习室场景。",
		BackgroundType:           "image",
		BackgroundURL:            "assets/scenes/study_room.jpg",
		BackgroundVideoURL:       "assets/scenes/study_room.mp4",
		BackgroundPosterURL:      "assets/scenes/study_room.jpg",
		AmbientAudioURL:          "assets/audio/library_noise.mp3",
		DefaultActionKey:         "patrol_normal",
		AvailableActionKeysJSON:  `["patrol_enter","patrol_normal","patrol_phone","patrol_sleeping","patrol_absent","finish_success","fail"]`,
		ModelResultActionMapJSON: `{"normal":"patrol_normal","phone":"patrol_phone","sleeping":"patrol_sleeping","absent":"patrol_absent","success":"finish_success","failed":"fail"}`,
		MetadataJSON:             emptyJSON,
	}
}

func DefaultActions() []model.ActionConfig {
	return []model.ActionConfig{
		defaultAction("patrol_enter", "巡查入场", 10, "assets/actions/study_room/patrol_enter.mp4", "patrol_normal"),
		defaultAction("patrol_normal", "正常巡查", 20, "assets/actions/study_room/patrol_normal.mp4", ""),
		defaultAction("patrol_phone", "手机违规", 30, "assets/actions/study_room/patrol_phone.mp4", ""),
		defaultAction("patrol_sleeping", "睡觉违规", 30, "assets/actions/study_room/patrol_sleeping.mp4", ""),
		defaultAction("patrol_absent", "离席违规", 30, "assets/actions/study_room/patrol_absent.mp4", ""),
		defaultAction("finish_success", "光荣下班", 40, "assets/actions/study_room/finish_success.mp4", ""),
		defaultAction("fail", "禁闭失败", 50, "assets/actions/study_room/fail.mp4", ""),
	}
}

func DefaultModelConfig() model.ModelConfig {
	return model.ModelConfig{
		Name:               "default_vision_model",
		Provider:           "openai-compatible",
		Enabled:            false,
		BaseURL:            "",
		APIKeyEncrypted:    "",
		Model:              "",
		TimeoutMS:          30000,
		MaxImageWidth:      1024,
		Temperature:        0,
		Prompt:             "请判断截图中的用户是否正常专注学习或工作。",
		ResponseSchemaJSON: `{"type":"object","properties":{"status":{"type":"string"},"confidence":{"type":"number"},"reason":{"type":"string"}}}`,
		RetryCount:         1,
	}
}

func DefaultPatrolRule() model.PatrolRule {
	return model.PatrolRule{
		SlowMinSeconds:         180,
		SlowMaxSeconds:         480,
		NormalMinSeconds:       60,
		NormalMaxSeconds:       240,
		HighMinSeconds:         30,
		HighMaxSeconds:         120,
		MaxWarnings:            3,
		MaxViolations:          3,
		SuspiciousAddsWarning:  true,
		ViolationDirectFail:    false,
		CameraOffStrategy:      "suspicious",
		CaptureFailedStrategy:  "uncertain",
		ModelTimeoutRetryCount: 1,
		UserErrorMessage:       "巡查系统暂不可用，请联系管理员处理。",
	}
}

func DefaultTasks() []model.Task {
	return []model.Task{
		defaultTask("complete_one_session", "完成一次劳动", "完成任意一次劳动。", "session_complete", 1, 10, 10),
		defaultTask("focus_10_minutes", "连续专注 10 分钟", "单次劳动累计专注 10 分钟。", "focus_seconds", 600, 5, 20),
		defaultTask("clean_finish", "今日无案底完成劳动", "今日无违规完成一次劳动。", "clean_session", 1, 20, 30),
		defaultTask("camera_session", "开启摄像头完成一次劳动", "开启摄像头并完成一次劳动。", "camera_session", 1, 10, 40),
		defaultTask("view_profile", "查看角色档案", "查看一次角色档案。", "view_profile", 1, 3, 50),
	}
}

func DefaultBadges() []model.Badge {
	return []model.Badge{
		{
			BadgeKey:     "first_focus",
			Name:         "第一次劳动",
			Description:  "完成第一次专注劳动。",
			Enabled:      true,
			Unlocked:     false,
			MetadataJSON: emptyJSON,
		},
	}
}

func DefaultUserProgress() model.UserProgress {
	return model.UserProgress{
		Level:             1,
		TotalFocusSeconds: 0,
		Currency:          0,
		CurrentStreakDays: 0,
		LongestStreakDays: 0,
	}
}

func defaultAction(actionKey string, name string, priority int, videoURL string, nextActionKey string) model.ActionConfig {
	return model.ActionConfig{
		SceneKey:           "study_room",
		ActionKey:          actionKey,
		Name:               name,
		Enabled:            true,
		Priority:           priority,
		VideoURL:           videoURL,
		PosterURL:          "assets/scenes/study_room.jpg",
		DurationMS:         8000,
		NextActionKey:      nextActionKey,
		ModelResultMapJSON: emptyJSON,
		LocalRuleMapJSON:   emptyJSON,
		MetadataJSON:       emptyJSON,
	}
}

func defaultTask(taskKey string, name string, description string, taskType string, targetValue int, rewardCurrency int, sortOrder int) model.Task {
	return model.Task{
		TaskKey:        taskKey,
		Name:           name,
		Description:    description,
		TaskType:       taskType,
		TargetValue:    targetValue,
		RewardCurrency: rewardCurrency,
		Enabled:        true,
		SortOrder:      sortOrder,
		MetadataJSON:   emptyJSON,
	}
}
