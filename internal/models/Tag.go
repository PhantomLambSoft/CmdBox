package models

import "time"

// Tag represents a label or identifier that can be associated with commands, variables, or other entities for categorization.
type Tag struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex;not null"`
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CommandTag struct {
	ID        uint    `gorm:"primaryKey"`
	CommandID uint    `gorm:"uniqueIndex:idx_command_tag;not null"`
	Command   Command `gorm:"foreignKey:CommandID;constraint:OnDelete:CASCADE"`
	TagID     uint    `gorm:"uniqueIndex:idx_command_tag;not null"`
	Tag       Tag     `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
}

type VariableTag struct {
	ID         uint     `gorm:"primaryKey"`
	VariableID uint     `gorm:"uniqueIndex:idx_variable_tag;not null"`
	Variable   Variable `gorm:"foreignKey:VariableID;constraint:OnDelete:CASCADE"`
	TagID      uint     `gorm:"uniqueIndex:idx_variable_tag;not null"`
	Tag        Tag      `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE"`
	CreatedAt  time.Time
}
