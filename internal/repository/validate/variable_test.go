package validate

import (
	"strings"
	"testing"
)

func TestVariableValidatorDefaultAndValidation(t *testing.T) {
	v := NewVariableValidator(nil)
	if v.config.MaxNameLength != 100 {
		t.Fatalf("MaxNameLength = %d, want 100", v.config.MaxNameLength)
	}
	if _, ok := v.config.ReservedNames["env"]; !ok {
		t.Fatalf("default reserved names missing %q", "env")
	}

	assertValidationErrorContains(t, v.ValidateName(""), "variable name cannot be empty")
	assertValidationErrorContains(t, v.ValidateName("\t  \n"), "variable name cannot contain only whitespace")
	assertValidationErrorContains(t, v.ValidateName("bad name"), "variable name cannot contain spaces")
	assertValidationErrorContains(t, v.ValidateName("init"), "variable name init is reserved")
	assertValidationErrorContains(t, v.ValidateName(strings.Repeat("v", 101)), "is too long")

	assertValidationErrorContains(t, v.ValidateValue(""), "variable value cannot be empty")
	assertValidationErrorContains(t, v.ValidateValue("   "), "variable value cannot contain only whitespace")
	if err := v.ValidateValue("some-value"); err != nil {
		t.Fatalf("ValidateValue(valid) error = %v", err)
	}

	assertValidationErrorContains(t, v.ValidateNoSelfReference("path", "${HOME}:<path>"), "variable value cannot contain self reference")
	if err := v.ValidateNoSelfReference("path", "${HOME}:<other>"); err != nil {
		t.Fatalf("ValidateNoSelfReference(non-self) error = %v", err)
	}

	if err := v.ValidateCreate("path", "/usr/local/bin"); err != nil {
		t.Fatalf("ValidateCreate(valid) error = %v", err)
	}
	assertValidationErrorContains(t, v.ValidateCreate("path", "<path>/bin"), "self reference")

	name := "home"
	value := "/home/user"
	if err := v.ValidateUpdate(&name, &value); err != nil {
		t.Fatalf("ValidateUpdate(valid) error = %v", err)
	}
	if err := v.ValidateUpdate(nil, nil); err != nil {
		t.Fatalf("ValidateUpdate(nil,nil) error = %v", err)
	}

	badName := ""
	if err := v.ValidateUpdate(&badName, nil); err == nil {
		t.Fatalf("ValidateUpdate(bad name) expected error")
	}
	badValue := ""
	if err := v.ValidateUpdate(nil, &badValue); err == nil {
		t.Fatalf("ValidateUpdate(bad value) expected error")
	}

	selfRefValue := "<home>"
	if err := v.ValidateUpdate(&name, &selfRefValue); err == nil {
		t.Fatalf("ValidateUpdate(self-reference) expected error")
	}
}

func TestVariableValidatorDefaultConfigReservedNames(t *testing.T) {
	v := NewVariableValidator(nil)

	reserved := []string{
		"help", "init", "list", "ls", "add", "rm", "delete",
		"preview", "cwd", "env", "capture", "shell", "emit", "verbose",
	}
	for _, name := range reserved {
		t.Run(name, func(t *testing.T) {
			assertValidationErrorContains(t, v.ValidateName(name), "is reserved")
		})
	}
}

func TestVariableValidatorCustomConfig(t *testing.T) {
	config := VariableValidatorConfig{
		ReservedNames: map[string]struct{}{"custom": {}},
		MaxNameLength: 5,
	}
	v := NewVariableValidator(&config)

	if err := v.ValidateName("env"); err != nil {
		t.Fatalf("ValidateName(env) with custom config error = %v, want nil", err)
	}
	assertValidationErrorContains(t, v.ValidateName("custom"), "variable name custom is reserved")
	assertValidationErrorContains(t, v.ValidateName("toolong"), "is too long")
	if err := v.ValidateName("abcde"); err != nil {
		t.Fatalf("ValidateName(abcde) at max length error = %v, want nil", err)
	}
}

func TestVariableValidatorNameLengthBoundary(t *testing.T) {
	v := NewVariableValidator(nil)

	if err := v.ValidateName(strings.Repeat("v", 100)); err != nil {
		t.Fatalf("ValidateName(100 chars) error = %v, want nil", err)
	}
	assertValidationErrorContains(t, v.ValidateName(strings.Repeat("v", 101)), "is too long")
}

func TestVariableValidatorNameTrimmedChecks(t *testing.T) {
	v := NewVariableValidator(nil)

	assertValidationErrorContains(t, v.ValidateName("  env  "), "variable name env is reserved")
	if err := v.ValidateName("  path  "); err != nil {
		t.Fatalf("ValidateName(padded valid name) error = %v, want nil", err)
	}
}
