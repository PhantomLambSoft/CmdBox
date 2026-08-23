package validate

import "strings"

type TagValidatorConfig struct {
	ReservedNames map[string]struct{}
	MaxNameLength int
}

func DefaultTagValidatorConfig() TagValidatorConfig {
	reservedNames := []string{
		"help",
		"init",
		"list",
		"ls",
		"add",
		"rm",
		"delete",
	}
	set := make(map[string]struct{}, len(reservedNames))
	for _, r := range reservedNames {
		set[r] = struct{}{}
	}
	return TagValidatorConfig{
		ReservedNames: set,
		MaxNameLength: 100,
	}
}

type TagValidator struct {
	config TagValidatorConfig
}

func NewTagValidator(config *TagValidatorConfig) *TagValidator {
	if config == nil {
		defaultConfig := DefaultTagValidatorConfig()
		config = &defaultConfig
	}
	return &TagValidator{config: *config}
}

func (v *TagValidator) ValidateCreate(name string) error {
	return v.ValidateName(name)
}

func (v *TagValidator) ValidateUpdate(name *string) error {
	if name != nil {
		return v.ValidateName(*name)
	}
	return nil
}

func (v *TagValidator) ValidateName(name string) error {
	if name == "" {
		return validationErrorf("tag name cannot be empty")
	}
	stripped := strings.TrimSpace(name)
	if stripped == "" {
		return validationErrorf("tag name cannot contain only whitespace")
	}
	if strings.Contains(stripped, " ") {
		return validationErrorf("tag name cannot contain spaces")
	}
	if _, reserved := v.config.ReservedNames[stripped]; reserved {
		return validationErrorf("tag name %s is reserved", stripped)
	}
	if len(stripped) > v.config.MaxNameLength {
		return validationErrorf("tag name %s is too long, maximum length is %d", stripped, v.config.MaxNameLength)
	}
	return nil
}
