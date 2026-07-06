package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"supervisor-game/internal/config"
	appcrypto "supervisor-game/internal/crypto"
	"supervisor-game/internal/model"

	driverMysql "github.com/go-sql-driver/mysql"
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

type RuntimeConnection struct {
	DB               *gorm.DB
	BootstrapDB      *gorm.DB
	Source           string
	Migrated         bool
	BootstrapError   error
	RuntimeError     error
	EnabledConfigID  uint
	EnabledConfigDSN string
}

func OpenRuntime(cfg config.Config) RuntimeConnection {
	result := RuntimeConnection{Source: "DB_DSN"}
	bootstrapDB, err := Open(cfg.DBDSN)
	if err != nil {
		result.BootstrapError = err
		result.RuntimeError = err
		return result
	}
	result.BootstrapDB = bootstrapDB
	result.DB = bootstrapDB

	if err := Migrate(bootstrapDB); err != nil {
		result.BootstrapError = err
		result.RuntimeError = err
		return result
	}
	result.Migrated = true

	var mysqlConfig model.MySQLConfig
	err = bootstrapDB.Where("enabled = ?", true).Order("updated_at DESC, id DESC").First(&mysqlConfig).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result
	}
	if err != nil {
		result.RuntimeError = fmt.Errorf("read mysql_configs: %w", err)
		return result
	}

	result.EnabledConfigID = mysqlConfig.ID
	dsn, err := DSNFromConfig(mysqlConfig, cfg.ConfigEncryptionKey)
	if err != nil {
		result.RuntimeError = err
		return result
	}
	result.EnabledConfigDSN = dsn

	runtimeDB, err := Open(dsn)
	if err != nil {
		result.RuntimeError = err
		return result
	}
	result.DB = runtimeDB
	result.Source = "mysql_configs"
	result.Migrated = false
	result.RuntimeError = nil
	return result
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

func DSNFromConfig(config model.MySQLConfig, encryptionKey string) (string, error) {
	password := ""
	if config.PasswordEncrypted != "" {
		decrypted, err := appcrypto.DecryptString(encryptionKey, config.PasswordEncrypted)
		if err != nil {
			return "", fmt.Errorf("decrypt mysql password: %w", err)
		}
		password = decrypted
	}

	loc := config.Timezone
	if loc == "" {
		loc = "Local"
	}
	params := map[string]string{
		"charset":   fallback(config.Charset, "utf8mb4"),
		"parseTime": "True",
		"loc":       loc,
	}

	dsnConfig := driverMysql.Config{
		User:                 config.Username,
		Passwd:               password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", config.Host, config.Port),
		DBName:               config.DatabaseName,
		Params:               params,
		AllowNativePasswords: true,
		ParseTime:            true,
	}
	return dsnConfig.FormatDSN(), nil
}

func TestMySQLConfig(config model.MySQLConfig, encryptionKey string) error {
	dsn, err := DSNFromConfig(config, encryptionKey)
	if err != nil {
		return err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetConnMaxLifetime(10 * time.Second)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db.Ping()
}

func fallback(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
