package validate

import (
	"strings"
	"testing"
)

func TestTagValidatorDefaultAndValidation(t *testing.T) {
	v := NewTagValidator(nil)
	if v.config.MaxNameLength != 100 {
		t.Fatalf("MaxNameLength = %d, want 100", v.config.MaxNameLength)
	}
	if _, ok := v.config.ReservedNames["list"]; !ok {
		t.Fatalf("default reserved names missing %q", "list")
	}

	assertValidationErrorContains(t, v.ValidateName(""), "tag name cannot be empty")
	assertValidationErrorContains(t, v.ValidateName("  \n"), "tag name cannot contain only whitespace")
	assertValidationErrorContains(t, v.ValidateName("has space"), "tag name cannot contain spaces")
	assertValidationErrorContains(t, v.ValidateName("help"), "tag name help is reserved")
	assertValidationErrorContains(t, v.ValidateName(strings.Repeat("t", 101)), "is too long")

	if err := v.ValidateName("release"); err != nil {
		t.Fatalf("ValidateName(valid) error = %v", err)
	}
	if err := v.ValidateCreate("stable"); err != nil {
		t.Fatalf("ValidateCreate(valid) error = %v", err)
	}

	name := "new-tag"
	if err := v.ValidateUpdate(&name); err != nil {
		t.Fatalf("ValidateUpdate(valid) error = %v", err)
	}
	if err := v.ValidateUpdate(nil); err != nil {
		t.Fatalf("ValidateUpdate(nil) error = %v", err)
	}

	badName := "has space"
	if err := v.ValidateUpdate(&badName); err == nil {
		t.Fatalf("ValidateUpdate(bad name) expected error")
	}

	assertValidationErrorContains(t, v.ValidateCreate("help"), "tag name help is reserved")
}

func TestTagValidatorDefaultConfigReservedNames(t *testing.T) {
	v := NewTagValidator(nil)

	reserved := []string{"help", "init", "list", "ls", "add", "rm", "delete"}
	for _, name := range reserved {
		t.Run(name, func(t *testing.T) {
			assertValidationErrorContains(t, v.ValidateName(name), "is reserved")
		})
	}
}

func TestTagValidatorCustomConfig(t *testing.T) {
	config := TagValidatorConfig{
		ReservedNames: map[string]struct{}{"custom": {}},
		MaxNameLength: 5,
	}
	v := NewTagValidator(&config)

	if err := v.ValidateName("help"); err != nil {
		t.Fatalf("ValidateName(help) with custom config error = %v, want nil", err)
	}
	assertValidationErrorContains(t, v.ValidateName("custom"), "tag name custom is reserved")
	assertValidationErrorContains(t, v.ValidateName("toolong"), "is too long")
	if err := v.ValidateName("abcde"); err != nil {
		t.Fatalf("ValidateName(abcde) at max length error = %v, want nil", err)
	}
}

func TestTagValidatorNameLengthBoundary(t *testing.T) {
	v := NewTagValidator(nil)

	if err := v.ValidateName(strings.Repeat("t", 100)); err != nil {
		t.Fatalf("ValidateName(100 chars) error = %v, want nil", err)
	}
	assertValidationErrorContains(t, v.ValidateName(strings.Repeat("t", 101)), "is too long")
}

func TestTagValidatorNameTrimmedChecks(t *testing.T) {
	v := NewTagValidator(nil)

	assertValidationErrorContains(t, v.ValidateName("  help  "), "tag name help is reserved")
	if err := v.ValidateName("  release  "); err != nil {
		t.Fatalf("ValidateName(padded valid name) error = %v, want nil", err)
	}
}
