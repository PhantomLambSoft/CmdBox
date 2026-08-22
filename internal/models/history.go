package models

import "time"

// CommandHistory records a single command execution which contains: the template as resolved at runtime, which
// variables were used, and the exit code of the command.
//
// Alias is stored as plain text, not a foreign key, so a history entry survives the source command being deleted or
// renamed.Profile is similarly a record of which profile the run happened under, not a live reference back to the
// command's current profile.
type CommandHistory struct {
	ID            string  `gorm:"primaryKey"`
	Alias         string  `gorm:"index;not null"`
	Template      string  `gorm:"not null"`
	Resolved      string  `gorm:"not null"`
	VariablesUsed *string // JSON-encoded map[string]string, or nil
	ExitCode      *int
	RanAt         time.Time `gorm:"index;not null"`

	ProfileID uint    `gorm:"index;not null"`
	Profile   Profile `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
}
