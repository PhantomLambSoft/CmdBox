package models

import "time"

// Command represents a saved command - called by its alias, the template is what gets executed. Other optional
// execution context is stored and used when the command is executed.
type Command struct {
	ID          uint   `gorm:"primaryKey"`
	Alias       string `gorm:"uniqueIndex:idx_command_alias_profile;not null"`
	Template    string `gorm:"not null"`
	Description string

	Cwd     *string
	Shell   *string
	Env     *string
	Timeout *int

	CreatedAt time.Time
	UpdatedAt time.Time
	Used      int
	LastUsed  *time.Time

	ProfileID uint    `gorm:"uniqueIndex:idx_command_alias_profile;not null"`
	Profile   Profile `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
}
