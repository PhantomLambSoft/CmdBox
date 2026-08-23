package validate

import (
	"strings"
)

type CommandValidatorConfig struct {
	ReservedAliases map[string]struct{}
	MaxAliasLength  int
}

var ReservedNames = []string{
	"help",
	"init",
	"add",
	"get",
	"update",
	"edit",
	"list",
	"ls",
	"search",
	"find",
	"delete",
	"del",
	"rm",
	"remove",
	"tag",
	"untag",
	// Reserved because they are options to the run command and can conflict when dynamically called
	"preview",
	"cwd",
	"env",
	"capture",
	"shell",
	"emit",
	"verbose",
}

func DefaultCommandValidatorConfig() CommandValidatorConfig {
	set := make(map[string]struct{}, len(ReservedNames))
	for _, r := range ReservedNames {
		set[r] = struct{}{}
	}
	return CommandValidatorConfig{
		ReservedAliases: set,
		MaxAliasLength:  100,
	}
}

type CommandValidator struct {
	config CommandValidatorConfig
}

func NewCommandValidator(config *CommandValidatorConfig) *CommandValidator {
	if config == nil {
		defaultConfig := DefaultCommandValidatorConfig()
		config = &defaultConfig
	}
	return &CommandValidator{config: *config}
}

func (v *CommandValidator) ValidateCreate(alias, template string) error {
	if err := v.ValidateAlias(alias); err != nil {
		return err
	}
	if err := v.ValidateTemplate(template); err != nil {
		return err
	}
	return v.ValidateNoSelfReference(alias, template)
}

func (v *CommandValidator) ValidateUpdate(alias, template *string) error {
	if alias != nil {
		if err := v.ValidateAlias(*alias); err != nil {
			return err
		}
	}
	if template != nil {
		if err := v.ValidateTemplate(*template); err != nil {
			return err
		}
	}
	if alias != nil && template != nil {
		return v.ValidateNoSelfReference(*alias, *template)
	}
	return nil
}

func (v *CommandValidator) ValidateAlias(alias string) error {
	if alias == "" {
		return validationErrorf("alias cannot be empty")
	}
	stripped := strings.TrimSpace(alias)
	if stripped == "" {
		return validationErrorf("alias cannot contain only whitespace")
	}
	if strings.Contains(stripped, " ") {
		return validationErrorf("alias cannot contain spaces")
	}
	if _, reserved := v.config.ReservedAliases[stripped]; reserved {
		return validationErrorf("alias %s is reserved", stripped)
	}
	if len(stripped) > v.config.MaxAliasLength {
		return validationErrorf("alias %q is too long, maximum length is %d", stripped, v.config.MaxAliasLength)
	}
	return nil
}

func (v *CommandValidator) ValidateTemplate(template string) error {
	if template == "" {
		return validationErrorf("template cannot be empty")
	}
	if strings.TrimSpace(template) == "" {
		return validationErrorf("template cannot contain only whitespace")
	}
	return nil
}

func (v *CommandValidator) ValidateNoSelfReference(alias, template string) error {
	selfRef := "<" + alias + ">"
	if strings.Contains(template, selfRef) {
		return validationErrorf("template cannot contain self-reference %q", selfRef)
	}
	return nil
}
