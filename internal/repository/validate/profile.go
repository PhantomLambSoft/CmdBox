package validate

import "strings"

// ProfileValidatorConfig defines configuration options for validating user profiles, such as maximum name length.
type ProfileValidatorConfig struct {
	MaxNameLength int
}

func DefaultProfileValidatorConfig() ProfileValidatorConfig {
	return ProfileValidatorConfig{
		MaxNameLength: 100,
	}
}

type ProfileValidator struct {
	config ProfileValidatorConfig
}

func NewProfileValidator(config *ProfileValidatorConfig) *ProfileValidator {
	if config == nil {
		defaultConfig := DefaultProfileValidatorConfig()
		config = &defaultConfig
	}
	return &ProfileValidator{config: *config}
}

// ValidateCreate validates the profile name and description before profile creation.
// Returns an error if the name is invalid.
func (v *ProfileValidator) ValidateCreate(name, description string) error {
	return v.ValidateName(name)
}

// ValidateUpdate validates the provided profile name and description during an update operation.
// Returns an error if the name is invalid.
func (v *ProfileValidator) ValidateUpdate(name, description *string) error {
	if name != nil {
		if err := v.ValidateName(*name); err != nil {
			return err
		}
	}
	return nil
}

// ValidateName validates a profile name against specific criteria such as non-emptiness, no whitespace, and max length.
func (v *ProfileValidator) ValidateName(name string) error {
	if name == "" {
		return validationErrorf("profile name cannot be empty")
	}
	stripped := strings.TrimSpace(name)
	if stripped == "" {
		return validationErrorf("profile name cannot contain only whitespace")
	}
	if strings.Contains(stripped, " ") {
		return validationErrorf("profile name cannot contain spaces")
	}
	if len(stripped) > v.config.MaxNameLength {
		return validationErrorf("profile name %s is too long, maximum length is %d", stripped, v.config.MaxNameLength)
	}
	return nil
}
