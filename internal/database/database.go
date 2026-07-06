package database

import (
	"errors"
	"fmt"
	"time"

	"supervisor-game/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var ErrDSNMissing = errors.New("DB_DSN is empty")

func Open(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, ErrDSNMissing
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("unwrap database: %w", err)
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	return db.AutoMigrate(
		&model.AppSetting{},
		&model.Character{},
		&model.UserSetting{},
		&model.Scene{},
		&model.ActionConfig{},
		&model.ModelConfig{},
		&model.PatrolRule{},
		&model.MySQLConfig{},
		&model.WorkSession{},
		&model.PatrolRecord{},
		&model.DailyStat{},
		&model.Task{},
		&model.TaskRecord{},
		&model.Badge{},
		&model.UserProgress{},
	)
}
