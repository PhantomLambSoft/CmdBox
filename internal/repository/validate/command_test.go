package validate

import (
	"errors"
	"strings"
	"testing"
)

func assertValidationErrorContains(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected validation error containing %q, got nil", contains)
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("error %q does not contain %q", err.Error(), contains)
	}
}

func TestCommandValidatorDefaultConfig(t *testing.T) {
	v := NewCommandValidator(nil)

	if v.config.MaxAliasLength != 100 {
		t.Fatalf("MaxAliasLength = %d, want 100", v.config.MaxAliasLength)
	}
	if _, ok := v.config.ReservedAliases["help"]; !ok {
		t.Fatalf("default reserved aliases missing %q", "help")
	}
}

func TestCommandValidatorDefaultConfigReservedAliases(t *testing.T) {
	v := NewCommandValidator(nil)

	reserved := []string{
		"help", "init", "add", "get", "update", "edit", "list", "ls",
		"search", "find", "delete", "del", "rm", "remove", "tag", "untag",
		"preview", "cwd", "env", "capture", "shell", "emit", "verbose",
	}
	for _, name := range reserved {
		t.Run(name, func(t *testing.T) {
			assertValidationErrorContains(t, v.ValidateAlias(name), "is reserved")
		})
	}
}

func TestCommandValidatorCustomConfig(t *testing.T) {
	config := CommandValidatorConfig{
		ReservedAliases: map[string]struct{}{"custom": {}},
		MaxAliasLength:  5,
	}
	v := NewCommandValidator(&config)

	if err := v.ValidateAlias("help"); err != nil {
		t.Fatalf("ValidateAlias(help) with custom config error = %v, want nil", err)
	}
	assertValidationErrorContains(t, v.ValidateAlias("custom"), "alias custom is reserved")
	assertValidationErrorContains(t, v.ValidateAlias("toolong"), "is too long")
	if err := v.ValidateAlias("abcde"); err != nil {
		t.Fatalf("ValidateAlias(abcde) at max length error = %v, want nil", err)
	}
}

func TestCommandValidatorAliasLengthBoundary(t *testing.T) {
	v := NewCommandValidator(nil)

	if err := v.ValidateAlias(strings.Repeat("a", 100)); err != nil {
		t.Fatalf("ValidateAlias(100 chars) error = %v, want nil", err)
	}
	assertValidationErrorContains(t, v.ValidateAlias(strings.Repeat("a", 101)), "is too long")
}

func TestCommandValidatorAliasTrimmedChecks(t *testing.T) {
	v := NewCommandValidator(nil)

	// Reserved-word and length checks operate on the trimmed alias.
	assertValidationErrorContains(t, v.ValidateAlias("  help  "), "alias help is reserved")
	if err := v.ValidateAlias("  build  "); err != nil {
		t.Fatalf("ValidateAlias(padded valid alias) error = %v, want nil", err)
	}
}

func TestCommandValidatorValidateAlias(t *testing.T) {
	v := NewCommandValidator(nil)

	cases := []struct {
		name        string
		alias       string
		errContains string
	}{
		{name: "empty", alias: "", errContains: "alias cannot be empty"},
		{name: "whitespace-only", alias: "   ", errContains: "alias cannot contain only whitespace"},
		{name: "contains spaces", alias: "hello world", errContains: "alias cannot contain spaces"},
		{name: "reserved", alias: "help", errContains: "alias help is reserved"},
		{name: "too long", alias: strings.Repeat("a", 101), errContains: "is too long"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateAlias(tc.alias)
			assertValidationErrorContains(t, err, tc.errContains)
		})
	}

	if err := v.ValidateAlias("valid-alias"); err != nil {
		t.Fatalf("ValidateAlias(valid) error = %v", err)
	}
}

func TestCommandValidatorTemplateAndSelfReference(t *testing.T) {
	v := NewCommandValidator(nil)

	assertValidationErrorContains(t, v.ValidateTemplate(""), "template cannot be empty")
	assertValidationErrorContains(t, v.ValidateTemplate(" \t\n "), "template cannot contain only whitespace")
	if err := v.ValidateTemplate("echo hello"); err != nil {
		t.Fatalf("ValidateTemplate(valid) error = %v", err)
	}

	assertValidationErrorContains(t, v.ValidateNoSelfReference("deploy", "run <deploy> now"), "template cannot contain self-reference")
	if err := v.ValidateNoSelfReference("deploy", "run <other> now"); err != nil {
		t.Fatalf("ValidateNoSelfReference(non-self) error = %v", err)
	}
}

func TestCommandValidatorCreateAndUpdate(t *testing.T) {
	v := NewCommandValidator(nil)

	if err := v.ValidateCreate("build", "go build ./..."); err != nil {
		t.Fatalf("ValidateCreate(valid) error = %v", err)
	}
	assertValidationErrorContains(t, v.ValidateCreate("help", "go test ./..."), "reserved")

	alias := "release"
	template := "go test ./..."
	if err := v.ValidateUpdate(&alias, &template); err != nil {
		t.Fatalf("ValidateUpdate(valid) error = %v", err)
	}
	if err := v.ValidateUpdate(nil, nil); err != nil {
		t.Fatalf("ValidateUpdate(nil,nil) error = %v", err)
	}

	badAlias := ""
	if err := v.ValidateUpdate(&badAlias, nil); err == nil {
		t.Fatalf("ValidateUpdate(bad alias) expected error")
	}

	badTemplate := "run <release> now"
	if err := v.ValidateUpdate(&alias, &badTemplate); err == nil {
		t.Fatalf("ValidateUpdate(self-reference) expected error")
	}

	assertValidationErrorContains(t, v.ValidateCreate("deploy", "run <deploy> now"), "self-reference")
}

func TestCommandValidatorUpdatePartialFields(t *testing.T) {
	v := NewCommandValidator(nil)

	template := "go build ./..."
	if err := v.ValidateUpdate(nil, &template); err != nil {
		t.Fatalf("ValidateUpdate(nil alias, valid template) error = %v", err)
	}

	alias := "build"
	if err := v.ValidateUpdate(&alias, nil); err != nil {
		t.Fatalf("ValidateUpdate(valid alias, nil template) error = %v", err)
	}

	badTemplate := ""
	if err := v.ValidateUpdate(nil, &badTemplate); err == nil {
		t.Fatalf("ValidateUpdate(nil alias, empty template) expected error")
	}
}
