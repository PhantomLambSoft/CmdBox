package validate

import "strings"

type VariableValidatorConfig struct {
	ReservedNames map[string]struct{}
	MaxNameLength int
}

func DefaultVariableValidatorConfig() VariableValidatorConfig {
	reservedNames := []string{
		"help",
		"init",
		"list",
		"ls",
		"add",
		"rm",
		"delete",
		// Reserved because they are options to the run command and can conflict when dynamically called
		"preview",
		"cwd",
		"env",
		"capture",
		"shell",
		"emit",
		"verbose",
	}
	set := make(map[string]struct{}, len(reservedNames))
	for _, name := range reservedNames {
		set[name] = struct{}{}
	}
	return VariableValidatorConfig{
		ReservedNames: set,
		MaxNameLength: 100,
	}
}

type VariableValidator struct {
	config VariableValidatorConfig
}

func NewVariableValidator(config *VariableValidatorConfig) *VariableValidator {
	if config == nil {
		defaultConfig := DefaultVariableValidatorConfig()
		config = &defaultConfig
	}
	return &VariableValidator{config: *config}
}

func (v *VariableValidator) ValidateCreate(name, value string) error {
	if err := v.ValidateName(name); err != nil {
		return err
	}
	if err := v.ValidateValue(value); err != nil {
		return err
	}
	return v.ValidateNoSelfReference(name, value)
}

func (v *VariableValidator) ValidateUpdate(name, value *string) error {
	if name != nil {
		if err := v.ValidateName(*name); err != nil {
			return err
		}
	}
	if value != nil {
		if err := v.ValidateValue(*value); err != nil {
			return err
		}
	}
	if name != nil && value != nil {
		return v.ValidateNoSelfReference(*name, *value)
	}
	return nil
}

func (v *VariableValidator) ValidateName(name string) error {
	if name == "" {
		return validationErrorf("variable name cannot be empty")
	}
	stripped := strings.TrimSpace(name)
	if stripped == "" {
		return validationErrorf("variable name cannot contain only whitespace")
	}
	if strings.Contains(stripped, " ") {
		return validationErrorf("variable name cannot contain spaces")
	}
	if _, reserved := v.config.ReservedNames[stripped]; reserved {
		return validationErrorf("variable name %s is reserved", stripped)
	}
	if len(stripped) > v.config.MaxNameLength {
		return validationErrorf("variable name %s is too long, maximum length is %d", stripped, v.config.MaxNameLength)
	}
	return nil
}

func (v *VariableValidator) ValidateValue(value string) error {
	if value == "" {
		return validationErrorf("variable value cannot be empty")
	}
	if strings.TrimSpace(value) == "" {
		return validationErrorf("variable value cannot contain only whitespace")
	}
	return nil
}

func (v *VariableValidator) ValidateNoSelfReference(name, value string) error {
	selfRef := "<" + name + ">"
	if strings.Contains(value, selfRef) {
		return validationErrorf("variable value cannot contain self reference %s", selfRef)
	}
	return nil
}
