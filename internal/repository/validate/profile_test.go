package validate

import (
	"strings"
	"testing"
)

func TestProfileValidatorDefaultAndValidation(t *testing.T) {
	v := NewProfileValidator(nil)
	if v.config.MaxNameLength != 100 {
		t.Fatalf("MaxNameLength = %d, want 100", v.config.MaxNameLength)
	}

	assertValidationErrorContains(t, v.ValidateName(""), "profile name cannot be empty")
	assertValidationErrorContains(t, v.ValidateName("   "), "profile name cannot contain only whitespace")
	assertValidationErrorContains(t, v.ValidateName("bad name"), "profile name cannot contain spaces")
	assertValidationErrorContains(t, v.ValidateName(strings.Repeat("p", 101)), "is too long")

	if err := v.ValidateName("work"); err != nil {
		t.Fatalf("ValidateName(valid) error = %v", err)
	}
	if err := v.ValidateCreate("dev"); err != nil {
		t.Fatalf("ValidateCreate(valid) error = %v", err)
	}

	name := "prod"
	if err := v.ValidateUpdate(&name); err != nil {
		t.Fatalf("ValidateUpdate(valid) error = %v", err)
	}
	if err := v.ValidateUpdate(nil); err != nil {
		t.Fatalf("ValidateUpdate(nil) error = %v", err)
	}

	badName := ""
	if err := v.ValidateUpdate(&badName); err == nil {
		t.Fatalf("ValidateUpdate(bad name) expected error")
	}
}

func TestProfileValidatorCustomConfig(t *testing.T) {
	config := ProfileValidatorConfig{MaxNameLength: 5}
	v := NewProfileValidator(&config)

	assertValidationErrorContains(t, v.ValidateName("toolong"), "is too long")
	if err := v.ValidateName("abcde"); err != nil {
		t.Fatalf("ValidateName(abcde) at max length error = %v, want nil", err)
	}
}

func TestProfileValidatorNameLengthBoundary(t *testing.T) {
	v := NewProfileValidator(nil)

	if err := v.ValidateName(strings.Repeat("p", 100)); err != nil {
		t.Fatalf("ValidateName(100 chars) error = %v, want nil", err)
	}
	assertValidationErrorContains(t, v.ValidateName(strings.Repeat("p", 101)), "is too long")
}

func TestProfileValidatorNameTrimmedChecks(t *testing.T) {
	v := NewProfileValidator(nil)

	if err := v.ValidateName("  work  "); err != nil {
		t.Fatalf("ValidateName(padded valid name) error = %v, want nil", err)
	}
}
