package models

import "time"

// Profile is a named container that commands, variables, and settings can be scoped to. Commands and variables
// are scoped in the database, and settings are scoped via a per-profile named settings file.
type Profile struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex;not null"`
	Description *string
	DateCreated time.Time `gorm:"not null"`
	LastUsed    *time.Time
}

func (Profile) TableName() string {
	return "profile"
}

// ProfileState is a singleton row that tracks which profile is currently active for commands, variables, and settings
// independently of one another.
//
// All FK's have OnDelete:RESTRICT to prevent them from being deleted if they are currently active.
type ProfileState struct {
	ID uint `gorm:"primaryKey"`

	ActiveCommandProfileID uint    `gorm:"not null"`
	ActiveCommandProfile   Profile `gorm:"foreignKey:ActiveCommandProfileID;constraint:OnDelete:RESTRICT"`

	ActiveVariableProfileID uint    `gorm:"not null"`
	ActiveVariableProfile   Profile `gorm:"foreignKey:ActiveVariableProfileID;constraint:OnDelete:RESTRICT"`

	ActiveSettingsProfileID uint    `gorm:"not null"`
	ActiveSettingsProfile   Profile `gorm:"foreignKey:ActiveSettingsProfileID;constraint:OnDelete:RESTRICT"`
}

func (ProfileState) TableName() string {
	return "profile_state"
}

// Linked reports whether all three active profiles currently point to the same profile.
func (s ProfileState) Linked() bool {
	return s.ActiveCommandProfileID == s.ActiveVariableProfileID && s.ActiveVariableProfileID == s.ActiveSettingsProfileID
}
