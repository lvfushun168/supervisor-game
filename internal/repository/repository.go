package repository

import (
	"errors"

	"supervisor-game/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) FirstUserSetting() (model.UserSetting, error) {
	var setting model.UserSetting
	err := r.db.Order("id ASC").First(&setting).Error
	return setting, err
}

func (r *Repository) SaveUserSetting(setting *model.UserSetting) error {
	return r.db.Save(setting).Error
}

func (r *Repository) EnabledScenes() ([]model.Scene, error) {
	var scenes []model.Scene
	err := r.db.Where("enabled = ?", true).Order("id ASC").Find(&scenes).Error
	return scenes, err
}

func (r *Repository) DefaultCharacter() (model.Character, error) {
	var character model.Character
	err := r.db.Where("enabled = ?", true).Order("id ASC").First(&character).Error
	return character, err
}

func (r *Repository) LatestPatrolRule() (model.PatrolRule, error) {
	var rule model.PatrolRule
	err := r.db.Order("id ASC").First(&rule).Error
	return rule, err
}

func (r *Repository) CountByModel(value any) (int64, error) {
	var count int64
	err := r.db.Model(value).Count(&count).Error
	return count, err
}

func (r *Repository) UpsertAppSetting(setting model.AppSetting) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "setting_key"}},
		DoNothing: true,
	}).Create(&setting).Error
}

func (r *Repository) UpsertCharacter(character model.Character) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "character_key"}},
		DoNothing: true,
	}).Create(&character).Error
}

func (r *Repository) UpsertScene(scene model.Scene) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "scene_key"}},
		DoNothing: true,
	}).Create(&scene).Error
}

func (r *Repository) UpsertAction(action model.ActionConfig) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "scene_key"}, {Name: "action_key"}},
		DoNothing: true,
	}).Create(&action).Error
}

func (r *Repository) UpsertTask(task model.Task) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "task_key"}},
		DoNothing: true,
	}).Create(&task).Error
}

func (r *Repository) UpsertBadge(badge model.Badge) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "badge_key"}},
		DoNothing: true,
	}).Create(&badge).Error
}

func (r *Repository) EnsureUserSetting(setting model.UserSetting) error {
	var existing model.UserSetting
	err := r.db.Order("id ASC").First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.Create(&setting).Error
}

func (r *Repository) EnsurePatrolRule(rule model.PatrolRule) error {
	var existing model.PatrolRule
	err := r.db.Order("id ASC").First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.Create(&rule).Error
}

func (r *Repository) EnsureModelConfig(config model.ModelConfig) error {
	var existing model.ModelConfig
	err := r.db.Order("id ASC").First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.Create(&config).Error
}

func (r *Repository) EnsureUserProgress(progress model.UserProgress) error {
	var existing model.UserProgress
	err := r.db.Order("id ASC").First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.Create(&progress).Error
}
