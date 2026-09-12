package models

import "time"

// Variable represents a variable model that contains a key-value that can be stored and recalled inside a command
// or other variables.
type Variable struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"uniqueIndex:idx_variable_name_profile;not null"`
	Value string `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	ProfileID uint    `gorm:"uniqueIndex:idx_variable_name_profile;not null"`
	Profile   Profile `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`

	// Tags is only populated at the service layer by VariableService.GetVariableWithTags. It is nil on commands
	// returned form the repository and does not get stored in a database table at all. It just picks up tags
	// queried at the service layer.
	Tags []Tag `gorm:"-"`
}
