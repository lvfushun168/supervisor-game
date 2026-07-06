package service

import (
	"errors"
	"testing"
)

func TestValidateUserSettingAcceptsDefault(t *testing.T) {
	if err := ValidateUserSetting(DefaultUserSetting()); err != nil {
		t.Fatalf("default user setting should be valid: %v", err)
	}
}

func TestValidateUserSettingRejectsInvalidEnum(t *testing.T) {
	setting := DefaultUserSetting()
	setting.Mode = "deep_work"

	err := ValidateUserSetting(setting)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestValidateUserSettingRejectsInvalidVolume(t *testing.T) {
	setting := DefaultUserSetting()
	setting.BackgroundVolume = 1.2

	err := ValidateUserSetting(setting)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestValidateUserSettingRejectsInvalidJSON(t *testing.T) {
	setting := DefaultUserSetting()
	setting.MetadataJSON = "{"

	err := ValidateUserSetting(setting)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
