package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
)

func setupCommandRepositoryTest(t *testing.T) (CommandRepository, ProfileRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:command-repo-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Fatalf("close sql db: %v", closeErr)
		}
	})

	// SQLite does not enforce foreign keys by default; enable it so FK constraints
	// declared on the models (e.g. CommandTag.TagID) are actually enforced in tests.
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Profile{},
		&models.ProfileState{},
		&models.Command{},
		&models.Variable{},
		&models.Tag{},
		&models.CommandTag{},
	); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	defaultDescription := "Automatically created default profile."
	defaultProfile := models.Profile{Name: DefaultProfileName, Description: &defaultDescription}
	if err := db.Create(&defaultProfile).Error; err != nil {
		t.Fatalf("seed default profile: %v", err)
	}

	state := models.ProfileState{
		ActiveCommandProfileID:  defaultProfile.ID,
		ActiveVariableProfileID: defaultProfile.ID,
		ActiveSettingsProfileID: defaultProfile.ID,
	}
	if err := db.Create(&state).Error; err != nil {
		t.Fatalf("seed profile state: %v", err)
	}

	profileRepo := NewProfileRepository(db, validate.NewProfileValidator(nil))
	commandRepo := NewCommandRepository(db, profileRepo, validate.NewCommandValidator(nil))
	return commandRepo, profileRepo, db
}

func createTestProfile(t *testing.T, db *gorm.DB, name string) models.Profile {
	t.Helper()
	profile := models.Profile{Name: name}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("create profile %q: %v", name, err)
	}
	return profile
}

func createTestTag(t *testing.T, db *gorm.DB, name string) models.Tag {
	t.Helper()
	tag := models.Tag{Name: name}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("create tag %q: %v", name, err)
	}
	return tag
}

func mustCreateCommand(t *testing.T, repo CommandRepository, input CommandCreateConfig) *models.Command {
	t.Helper()
	cmd, err := repo.Create(input)
	if err != nil {
		t.Fatalf("Create(%+v) error = %v", input, err)
	}
	return cmd
}

func intPtr(v int) *int {
	return &v
}

// --- Create ---

func TestCommandRepositoryCreate(t *testing.T) {
	repo, _, db := setupCommandRepositoryTest(t)

	t.Run("creates command using active profile", func(t *testing.T) {
		cmd, err := repo.Create(CommandCreateConfig{Alias: "build", Template: "go build ./..."})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if cmd.ID == 0 {
			t.Fatalf("Create() returned zero ID")
		}
		if cmd.Alias != "build" {
			t.Fatalf("Alias = %q, want %q", cmd.Alias, "build")
		}
		if cmd.Template != "go build ./..." {
			t.Fatalf("Template = %q, want %q", cmd.Template, "go build ./...")
		}
		if cmd.ProfileID == 0 {
			t.Fatalf("ProfileID = 0, want active profile id")
		}

		fromDB, err := repo.GetByID(cmd.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() after Create error = %v", err)
		}
		if fromDB.Alias != "build" {
			t.Fatalf("persisted alias = %q, want %q", fromDB.Alias, "build")
		}
	})

	t.Run("creates command with explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "explicit-profile")
		cmd, err := repo.Create(CommandCreateConfig{Alias: "deploy", Template: "echo deploy", ProfileID: &other.ID})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if cmd.ProfileID != other.ID {
			t.Fatalf("ProfileID = %d, want %d", cmd.ProfileID, other.ID)
		}
	})

	t.Run("trims alias whitespace", func(t *testing.T) {
		cmd, err := repo.Create(CommandCreateConfig{Alias: "  spaced  ", Template: "echo hi"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if cmd.Alias != "spaced" {
			t.Fatalf("Alias = %q, want %q", cmd.Alias, "spaced")
		}
	})

	t.Run("encodes non-empty env as json", func(t *testing.T) {
		env := map[string]string{"FOO": "bar", "BAZ": "qux"}
		cmd, err := repo.Create(CommandCreateConfig{Alias: "with-env", Template: "echo hi", Env: env})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if cmd.Env == nil {
			t.Fatalf("Env = nil, want encoded json")
		}
		want, err := json.Marshal(env)
		if err != nil {
			t.Fatalf("json.Marshal(env) error = %v", err)
		}
		if *cmd.Env != string(want) {
			t.Fatalf("Env = %q, want %q", *cmd.Env, string(want))
		}
	})

	t.Run("empty env map results in nil Env", func(t *testing.T) {
		cmd, err := repo.Create(CommandCreateConfig{Alias: "no-env", Template: "echo hi", Env: map[string]string{}})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if cmd.Env != nil {
			t.Fatalf("Env = %v, want nil", *cmd.Env)
		}
	})

	t.Run("stores optional fields", func(t *testing.T) {
		cwd := "/tmp"
		shell := "/bin/bash"
		timeout := 30
		description := "builds the project"
		cmd, err := repo.Create(CommandCreateConfig{
			Alias:       "full",
			Template:    "echo hi",
			Description: &description,
			Cwd:         &cwd,
			Shell:       &shell,
			Timeout:     &timeout,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if cmd.Description == nil || *cmd.Description != description {
			t.Fatalf("Description mismatch: %v", cmd.Description)
		}
		if cmd.Cwd == nil || *cmd.Cwd != cwd {
			t.Fatalf("Cwd mismatch: %v", cmd.Cwd)
		}
		if cmd.Shell == nil || *cmd.Shell != shell {
			t.Fatalf("Shell mismatch: %v", cmd.Shell)
		}
		if cmd.Timeout == nil || *cmd.Timeout != timeout {
			t.Fatalf("Timeout mismatch: %v", cmd.Timeout)
		}
	})

	t.Run("rejects empty alias", func(t *testing.T) {
		_, err := repo.Create(CommandCreateConfig{Alias: "", Template: "echo hi"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects reserved alias", func(t *testing.T) {
		_, err := repo.Create(CommandCreateConfig{Alias: "help", Template: "echo hi"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects self-referencing template", func(t *testing.T) {
		_, err := repo.Create(CommandCreateConfig{Alias: "deploy2", Template: "run <deploy2> now"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects duplicate alias in same profile", func(t *testing.T) {
		if _, err := repo.Create(CommandCreateConfig{Alias: "dup", Template: "echo one"}); err != nil {
			t.Fatalf("Create() first error = %v", err)
		}
		_, err := repo.Create(CommandCreateConfig{Alias: "dup", Template: "echo two"})
		if !errors.Is(err, ErrAliasConflict) {
			t.Fatalf("Create() second error = %v, want ErrAliasConflict", err)
		}
	})

	t.Run("allows same alias across different profiles", func(t *testing.T) {
		profileA := createTestProfile(t, db, "profile-a")
		profileB := createTestProfile(t, db, "profile-b")

		if _, err := repo.Create(CommandCreateConfig{Alias: "shared", Template: "echo a", ProfileID: &profileA.ID}); err != nil {
			t.Fatalf("Create() profile A error = %v", err)
		}
		if _, err := repo.Create(CommandCreateConfig{Alias: "shared", Template: "echo b", ProfileID: &profileB.ID}); err != nil {
			t.Fatalf("Create() profile B error = %v", err)
		}
	})

	t.Run("propagates error resolving active profile", func(t *testing.T) {
		if err := db.Delete(&models.ProfileState{}, 1).Error; err != nil {
			t.Fatalf("delete profile state: %v", err)
		}
		_, err := repo.Create(CommandCreateConfig{Alias: "no-state", Template: "echo hi"})
		if err == nil {
			t.Fatalf("Create() error = nil, want error resolving profile state")
		}
	})
}

// --- GetByAlias ---

func TestCommandRepositoryGetByAlias(t *testing.T) {
	repo, _, db := setupCommandRepositoryTest(t)
	cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "build", Template: "go build ./..."})

	t.Run("returns command scoped to active profile", func(t *testing.T) {
		got, err := repo.GetByAlias("build", nil)
		if err != nil {
			t.Fatalf("GetByAlias() error = %v", err)
		}
		if got.ID != cmd.ID {
			t.Fatalf("GetByAlias() id = %d, want %d", got.ID, cmd.ID)
		}
	})

	t.Run("lookup is case-sensitive", func(t *testing.T) {
		_, err := repo.GetByAlias("BUILD", nil)
		if !errors.Is(err, ErrUnknownAlias) {
			t.Fatalf("GetByAlias(\"BUILD\") error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("returns ErrUnknownAlias when missing", func(t *testing.T) {
		_, err := repo.GetByAlias("missing", nil)
		if !errors.Is(err, ErrUnknownAlias) {
			t.Fatalf("GetByAlias() error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("scopes lookup to explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "other-profile")
		otherCmd, err := repo.Create(CommandCreateConfig{Alias: "build", Template: "echo other", ProfileID: &other.ID})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repo.GetByAlias("build", &other.ID)
		if err != nil {
			t.Fatalf("GetByAlias(scoped) error = %v", err)
		}
		if got.ID != otherCmd.ID {
			t.Fatalf("GetByAlias(scoped) id = %d, want %d", got.ID, otherCmd.ID)
		}

		// still resolves to the active-profile version when no profile id is given.
		got, err = repo.GetByAlias("build", nil)
		if err != nil {
			t.Fatalf("GetByAlias(active) error = %v", err)
		}
		if got.ID != cmd.ID {
			t.Fatalf("GetByAlias(active) id = %d, want %d", got.ID, cmd.ID)
		}
	})

	t.Run("preserves and retrieves mixed-case alias exactly", func(t *testing.T) {
		mixed := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "MixedCase", Template: "echo mixed"})

		got, err := repo.GetByAlias("MixedCase", nil)
		if err != nil {
			t.Fatalf("GetByAlias(\"MixedCase\") error = %v", err)
		}
		if got.ID != mixed.ID {
			t.Fatalf("GetByAlias(\"MixedCase\") id = %d, want %d", got.ID, mixed.ID)
		}

		_, err = repo.GetByAlias("mixedcase", nil)
		if !errors.Is(err, ErrUnknownAlias) {
			t.Fatalf("GetByAlias(\"mixedcase\") error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("aliases differing only by case coexist in the same profile", func(t *testing.T) {
		capitalized := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "Build", Template: "echo capitalized"})

		lower, err := repo.GetByAlias("build", nil)
		if err != nil {
			t.Fatalf("GetByAlias(\"build\") error = %v", err)
		}
		if lower.ID != cmd.ID {
			t.Fatalf("GetByAlias(\"build\") id = %d, want %d", lower.ID, cmd.ID)
		}

		upper, err := repo.GetByAlias("Build", nil)
		if err != nil {
			t.Fatalf("GetByAlias(\"Build\") error = %v", err)
		}
		if upper.ID != capitalized.ID {
			t.Fatalf("GetByAlias(\"Build\") id = %d, want %d", upper.ID, capitalized.ID)
		}
	})
}

// --- GetByID ---

func TestCommandRepositoryGetByID(t *testing.T) {
	repo, _, db := setupCommandRepositoryTest(t)
	cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "build", Template: "go build ./..."})

	t.Run("returns command scoped to active profile", func(t *testing.T) {
		got, err := repo.GetByID(cmd.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got.Alias != "build" {
			t.Fatalf("Alias = %q, want %q", got.Alias, "build")
		}
	})

	t.Run("returns ErrUnknownCommand when missing", func(t *testing.T) {
		_, err := repo.GetByID(cmd.ID+999999, nil)
		if !errors.Is(err, ErrUnknownCommand) {
			t.Fatalf("GetByID() error = %v, want ErrUnknownCommand", err)
		}
	})

	t.Run("returns ErrUnknownCommand when scoped to wrong profile", func(t *testing.T) {
		other := createTestProfile(t, db, "wrong-profile")
		_, err := repo.GetByID(cmd.ID, &other.ID)
		if !errors.Is(err, ErrUnknownCommand) {
			t.Fatalf("GetByID(wrong profile) error = %v, want ErrUnknownCommand", err)
		}
	})
}

// --- Update ---

func TestCommandRepositoryUpdateValidation(t *testing.T) {
	repo, _, _ := setupCommandRepositoryTest(t)
	cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "build", Template: "go build ./..."})

	t.Run("nil command returns ErrNoUpdateTarget", func(t *testing.T) {
		_, err := repo.Update(nil, CommandUpdateConfig{})
		if !errors.Is(err, ErrNoUpdateTarget) {
			t.Fatalf("Update(nil) error = %v, want ErrNoUpdateTarget", err)
		}
	})

	t.Run("conflicting clear and set", func(t *testing.T) {
		cwd := "/tmp"
		shell := "/bin/bash"
		timeout := 10
		cases := []struct {
			name  string
			input CommandUpdateConfig
		}{
			{name: "cwd", input: CommandUpdateConfig{ClearCwd: true, Cwd: &cwd}},
			{name: "shell", input: CommandUpdateConfig{ClearShell: true, Shell: &shell}},
			{name: "timeout", input: CommandUpdateConfig{ClearTimeout: true, Timeout: &timeout}},
			{name: "env", input: CommandUpdateConfig{ClearEnv: true, Env: map[string]string{"A": "1"}}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				fresh := *cmd
				_, err := repo.Update(&fresh, tc.input)
				if !errors.Is(err, ErrConflictingClearAndSet) {
					t.Fatalf("Update() error = %v, want ErrConflictingClearAndSet", err)
				}
			})
		}
	})

	t.Run("rejects reserved alias on rename", func(t *testing.T) {
		fresh := *cmd
		newAlias := "help"
		_, err := repo.Update(&fresh, CommandUpdateConfig{Alias: &newAlias})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Update() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects self-referencing template after merge", func(t *testing.T) {
		fresh := *cmd
		newTemplate := "run <build> now"
		_, err := repo.Update(&fresh, CommandUpdateConfig{Template: &newTemplate})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Update() error = %v, want ErrValidation", err)
		}
	})
}

func TestCommandRepositoryUpdateFields(t *testing.T) {
	repo, _, db := setupCommandRepositoryTest(t)

	t.Run("updates alias and template", func(t *testing.T) {
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "rename-me", Template: "echo old"})
		newAlias := "renamed"
		newTemplate := "echo new"
		updated, err := repo.Update(cmd, CommandUpdateConfig{Alias: &newAlias, Template: &newTemplate})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Alias != "renamed" || updated.Template != "echo new" {
			t.Fatalf("Update() = %+v, want alias=renamed template='echo new'", updated)
		}

		persisted, err := repo.GetByID(cmd.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if persisted.Alias != "renamed" {
			t.Fatalf("persisted alias = %q, want %q", persisted.Alias, "renamed")
		}
	})

	t.Run("trims renamed alias whitespace", func(t *testing.T) {
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "trim-me", Template: "echo hi"})
		newAlias := "  trimmed  "
		updated, err := repo.Update(cmd, CommandUpdateConfig{Alias: &newAlias})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Alias != "trimmed" {
			t.Fatalf("Alias = %q, want %q", updated.Alias, "trimmed")
		}
	})

	t.Run("updates description", func(t *testing.T) {
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "desc-me", Template: "echo hi"})
		description := "new description"
		updated, err := repo.Update(cmd, CommandUpdateConfig{Description: &description})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Description == nil || *updated.Description != description {
			t.Fatalf("Description = %v, want %q", updated.Description, description)
		}
	})

	t.Run("sets and clears cwd", func(t *testing.T) {
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "cwd-me", Template: "echo hi"})
		cwd := "/tmp"
		updated, err := repo.Update(cmd, CommandUpdateConfig{Cwd: &cwd})
		if err != nil {
			t.Fatalf("Update(set cwd) error = %v", err)
		}
		if updated.Cwd == nil || *updated.Cwd != cwd {
			t.Fatalf("Cwd = %v, want %q", updated.Cwd, cwd)
		}

		updated, err = repo.Update(updated, CommandUpdateConfig{ClearCwd: true})
		if err != nil {
			t.Fatalf("Update(clear cwd) error = %v", err)
		}
		if updated.Cwd != nil {
			t.Fatalf("Cwd = %v, want nil", *updated.Cwd)
		}
	})

	t.Run("sets and clears shell", func(t *testing.T) {
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "shell-me", Template: "echo hi"})
		shell := "/bin/zsh"
		updated, err := repo.Update(cmd, CommandUpdateConfig{Shell: &shell})
		if err != nil {
			t.Fatalf("Update(set shell) error = %v", err)
		}
		if updated.Shell == nil || *updated.Shell != shell {
			t.Fatalf("Shell = %v, want %q", updated.Shell, shell)
		}

		updated, err = repo.Update(updated, CommandUpdateConfig{ClearShell: true})
		if err != nil {
			t.Fatalf("Update(clear shell) error = %v", err)
		}
		if updated.Shell != nil {
			t.Fatalf("Shell = %v, want nil", *updated.Shell)
		}
	})

	t.Run("sets and clears timeout", func(t *testing.T) {
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "timeout-me", Template: "echo hi"})
		updated, err := repo.Update(cmd, CommandUpdateConfig{Timeout: intPtr(45)})
		if err != nil {
			t.Fatalf("Update(set timeout) error = %v", err)
		}
		if updated.Timeout == nil || *updated.Timeout != 45 {
			t.Fatalf("Timeout = %v, want 45", updated.Timeout)
		}

		updated, err = repo.Update(updated, CommandUpdateConfig{ClearTimeout: true})
		if err != nil {
			t.Fatalf("Update(clear timeout) error = %v", err)
		}
		if updated.Timeout != nil {
			t.Fatalf("Timeout = %v, want nil", *updated.Timeout)
		}
	})

	t.Run("sets and clears env", func(t *testing.T) {
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "env-me", Template: "echo hi"})
		env := map[string]string{"A": "1"}
		updated, err := repo.Update(cmd, CommandUpdateConfig{Env: env})
		if err != nil {
			t.Fatalf("Update(set env) error = %v", err)
		}
		if updated.Env == nil {
			t.Fatalf("Env = nil, want encoded json")
		}
		want, _ := json.Marshal(env)
		if *updated.Env != string(want) {
			t.Fatalf("Env = %q, want %q", *updated.Env, string(want))
		}

		updated, err = repo.Update(updated, CommandUpdateConfig{ClearEnv: true})
		if err != nil {
			t.Fatalf("Update(clear env) error = %v", err)
		}
		if updated.Env != nil {
			t.Fatalf("Env = %v, want nil", *updated.Env)
		}
	})

	t.Run("no-op update leaves fields unchanged", func(t *testing.T) {
		description := "kept"
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "noop-me", Template: "echo hi", Description: &description})
		updated, err := repo.Update(cmd, CommandUpdateConfig{})
		if err != nil {
			t.Fatalf("Update(no-op) error = %v", err)
		}
		if updated.Alias != "noop-me" || updated.Template != "echo hi" {
			t.Fatalf("Update(no-op) changed alias/template: %+v", updated)
		}
		if updated.Description == nil || *updated.Description != description {
			t.Fatalf("Update(no-op) changed description: %v", updated.Description)
		}
	})

	t.Run("rejects rename to alias already used in same profile", func(t *testing.T) {
		_ = mustCreateCommand(t, repo, CommandCreateConfig{Alias: "taken", Template: "echo taken"})
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "renaming", Template: "echo hi"})

		newAlias := "taken"
		_, err := repo.Update(cmd, CommandUpdateConfig{Alias: &newAlias})
		if !errors.Is(err, ErrAliasConflict) {
			t.Fatalf("Update() error = %v, want ErrAliasConflict", err)
		}
	})

	t.Run("allows rename to alias used in a different profile", func(t *testing.T) {
		other := createTestProfile(t, db, "rename-profile")
		if _, err := repo.Create(CommandCreateConfig{Alias: "cross-profile", Template: "echo other", ProfileID: &other.ID}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "renaming2", Template: "echo hi"})

		newAlias := "cross-profile"
		updated, err := repo.Update(cmd, CommandUpdateConfig{Alias: &newAlias})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Alias != "cross-profile" {
			t.Fatalf("Alias = %q, want %q", updated.Alias, "cross-profile")
		}
	})
}

// --- Delete ---

func TestCommandRepositoryDelete(t *testing.T) {
	repo, _, _ := setupCommandRepositoryTest(t)

	t.Run("nil command is a no-op", func(t *testing.T) {
		if err := repo.Delete(nil); err != nil {
			t.Fatalf("Delete(nil) error = %v", err)
		}
	})

	t.Run("deletes existing command", func(t *testing.T) {
		cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "delete-me", Template: "echo hi"})
		if err := repo.Delete(cmd); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		_, err := repo.GetByID(cmd.ID, nil)
		if !errors.Is(err, ErrUnknownCommand) {
			t.Fatalf("GetByID() after delete error = %v, want ErrUnknownCommand", err)
		}
	})
}

// --- RecordUse ---

func TestCommandRepositoryRecordUse(t *testing.T) {
	repo, _, _ := setupCommandRepositoryTest(t)
	cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "use-me", Template: "echo hi"})

	if cmd.Used != 0 {
		t.Fatalf("initial Used = %d, want 0", cmd.Used)
	}

	if err := repo.RecordUse(cmd.ID); err != nil {
		t.Fatalf("RecordUse() error = %v", err)
	}
	updated, err := repo.GetByID(cmd.ID, nil)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if updated.Used != 1 {
		t.Fatalf("Used = %d, want 1", updated.Used)
	}
	if updated.LastUsed == nil {
		t.Fatalf("LastUsed = nil, want set")
	}

	if err := repo.RecordUse(cmd.ID); err != nil {
		t.Fatalf("RecordUse() second call error = %v", err)
	}
	updated, err = repo.GetByID(cmd.ID, nil)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if updated.Used != 2 {
		t.Fatalf("Used = %d, want 2", updated.Used)
	}

	t.Run("unknown id returns ErrUnknownCommand", func(t *testing.T) {
		err := repo.RecordUse(cmd.ID + 999999)
		if !errors.Is(err, ErrUnknownCommand) {
			t.Fatalf("RecordUse() error = %v, want ErrUnknownCommand", err)
		}
	})
}

// --- AddTags / RemoveTags ---

func TestCommandRepositoryAddTags(t *testing.T) {
	repo, _, db := setupCommandRepositoryTest(t)
	cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "tag-me", Template: "echo hi"})
	tagA := createTestTag(t, db, "tag-a")
	tagB := createTestTag(t, db, "tag-b")

	t.Run("empty tags is a no-op", func(t *testing.T) {
		result, err := repo.AddTags(cmd, nil)
		if err != nil {
			t.Fatalf("AddTags(nil) error = %v", err)
		}
		if len(result.Added) != 0 || len(result.Existing) != 0 {
			t.Fatalf("AddTags(nil) = %+v, want zero value", result)
		}
	})

	t.Run("adds new tags", func(t *testing.T) {
		result, err := repo.AddTags(cmd, []models.Tag{tagA, tagB})
		if err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}
		if len(result.Added) != 2 || len(result.Existing) != 0 {
			t.Fatalf("AddTags() = %+v, want both added", result)
		}

		var count int64
		if err := db.Model(&models.CommandTag{}).Where("command_id = ?", cmd.ID).Count(&count).Error; err != nil {
			t.Fatalf("count command_tags: %v", err)
		}
		if count != 2 {
			t.Fatalf("command_tags count = %d, want 2", count)
		}
	})

	t.Run("re-adding tags reports them as existing", func(t *testing.T) {
		result, err := repo.AddTags(cmd, []models.Tag{tagA, tagB})
		if err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}
		if len(result.Added) != 0 || len(result.Existing) != 2 {
			t.Fatalf("AddTags() = %+v, want both existing", result)
		}
	})

	t.Run("mix of new and existing tags", func(t *testing.T) {
		tagC := createTestTag(t, db, "tag-c")
		result, err := repo.AddTags(cmd, []models.Tag{tagA, tagC})
		if err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}
		if len(result.Added) != 1 || result.Added[0] != "tag-c" {
			t.Fatalf("Added = %v, want [tag-c]", result.Added)
		}
		if len(result.Existing) != 1 || result.Existing[0] != "tag-a" {
			t.Fatalf("Existing = %v, want [tag-a]", result.Existing)
		}
	})
}

func TestCommandRepositoryRemoveTags(t *testing.T) {
	repo, _, db := setupCommandRepositoryTest(t)
	cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "untag-me", Template: "echo hi"})
	tagA := createTestTag(t, db, "remove-a")
	tagB := createTestTag(t, db, "remove-b")

	t.Run("empty tags is a no-op", func(t *testing.T) {
		result, err := repo.RemoveTags(cmd, nil)
		if err != nil {
			t.Fatalf("RemoveTags(nil) error = %v", err)
		}
		if len(result.Removed) != 0 || len(result.NotAttached) != 0 {
			t.Fatalf("RemoveTags(nil) = %+v, want zero value", result)
		}
	})

	t.Run("removing unattached tags", func(t *testing.T) {
		result, err := repo.RemoveTags(cmd, []models.Tag{tagA, tagB})
		if err != nil {
			t.Fatalf("RemoveTags() error = %v", err)
		}
		if len(result.Removed) != 0 || len(result.NotAttached) != 2 {
			t.Fatalf("RemoveTags() = %+v, want both not attached", result)
		}
	})

	t.Run("removes attached tags and reports mix", func(t *testing.T) {
		if _, err := repo.AddTags(cmd, []models.Tag{tagA}); err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}

		result, err := repo.RemoveTags(cmd, []models.Tag{tagA, tagB})
		if err != nil {
			t.Fatalf("RemoveTags() error = %v", err)
		}
		if len(result.Removed) != 1 || result.Removed[0] != "remove-a" {
			t.Fatalf("Removed = %v, want [remove-a]", result.Removed)
		}
		if len(result.NotAttached) != 1 || result.NotAttached[0] != "remove-b" {
			t.Fatalf("NotAttached = %v, want [remove-b]", result.NotAttached)
		}

		var count int64
		if err := db.Model(&models.CommandTag{}).Where("command_id = ? AND tag_id = ?", cmd.ID, tagA.ID).Count(&count).Error; err != nil {
			t.Fatalf("count command_tags: %v", err)
		}
		if count != 0 {
			t.Fatalf("command_tags count for removed tag = %d, want 0", count)
		}
	})
}

// --- ListAll ---

func TestCommandRepositoryListAll(t *testing.T) {
	repo, _, db := setupCommandRepositoryTest(t)
	_ = mustCreateCommand(t, repo, CommandCreateConfig{Alias: "zeta", Template: "echo z"})
	_ = mustCreateCommand(t, repo, CommandCreateConfig{Alias: "alpha", Template: "echo a"})
	_ = mustCreateCommand(t, repo, CommandCreateConfig{Alias: "mid", Template: "echo m"})

	t.Run("defaults to alias order and limit 25", func(t *testing.T) {
		commands, err := repo.ListAll("", 0, nil)
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(commands) != 3 {
			t.Fatalf("len(commands) = %d, want 3", len(commands))
		}
		wantOrder := []string{"alpha", "mid", "zeta"}
		for i, want := range wantOrder {
			if commands[i].Alias != want {
				t.Fatalf("commands[%d].Alias = %q, want %q", i, commands[i].Alias, want)
			}
		}
	})

	t.Run("respects explicit order and limit", func(t *testing.T) {
		commands, err := repo.ListAll("-alias", 2, nil)
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(commands) != 2 {
			t.Fatalf("len(commands) = %d, want 2", len(commands))
		}
		if commands[0].Alias != "zeta" || commands[1].Alias != "mid" {
			t.Fatalf("commands = %+v, want [zeta, mid]", commands)
		}
	})

	t.Run("scopes to explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "list-all-other")
		if _, err := repo.Create(CommandCreateConfig{Alias: "in-other", Template: "echo hi", ProfileID: &other.ID}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		commands, err := repo.ListAll("", 0, &other.ID)
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(commands) != 1 || commands[0].Alias != "in-other" {
			t.Fatalf("commands = %+v, want only [in-other]", commands)
		}
	})

	t.Run("invalid order field returns error", func(t *testing.T) {
		_, err := repo.ListAll("bogus", 0, nil)
		if err == nil {
			t.Fatalf("ListAll() error = nil, want error")
		}
	})

	t.Run("propagates error resolving active profile", func(t *testing.T) {
		if err := db.Delete(&models.ProfileState{}, 1).Error; err != nil {
			t.Fatalf("delete profile state: %v", err)
		}
		_, err := repo.ListAll("", 0, nil)
		if err == nil {
			t.Fatalf("ListAll() error = nil, want error")
		}
	})
}

// --- ListByTags ---

func TestCommandRepositoryListByTags(t *testing.T) {
	repo, _, db := setupCommandRepositoryTest(t)
	tagA := createTestTag(t, db, "list-tag-a")
	tagB := createTestTag(t, db, "list-tag-b")

	cmdA := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "has-a", Template: "echo a"})
	cmdB := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "has-b", Template: "echo b"})
	cmdBoth := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "has-both", Template: "echo both"})
	_ = mustCreateCommand(t, repo, CommandCreateConfig{Alias: "has-neither", Template: "echo none"})

	if _, err := repo.AddTags(cmdA, []models.Tag{tagA}); err != nil {
		t.Fatalf("AddTags() error = %v", err)
	}
	if _, err := repo.AddTags(cmdB, []models.Tag{tagB}); err != nil {
		t.Fatalf("AddTags() error = %v", err)
	}
	if _, err := repo.AddTags(cmdBoth, []models.Tag{tagA, tagB}); err != nil {
		t.Fatalf("AddTags() error = %v", err)
	}

	t.Run("empty tags returns nil without error", func(t *testing.T) {
		commands, err := repo.ListByTags(nil, "", 0, nil)
		if err != nil {
			t.Fatalf("ListByTags(nil) error = %v", err)
		}
		if commands != nil {
			t.Fatalf("ListByTags(nil) = %v, want nil", commands)
		}
	})

	t.Run("returns commands matching any tag without duplicates", func(t *testing.T) {
		commands, err := repo.ListByTags([]models.Tag{tagA, tagB}, "alias", 0, nil)
		if err != nil {
			t.Fatalf("ListByTags() error = %v", err)
		}
		if len(commands) != 3 {
			t.Fatalf("len(commands) = %d, want 3 (no duplicates); commands = %+v", len(commands), commands)
		}
		wantAliases := map[string]bool{"has-a": true, "has-b": true, "has-both": true}
		for _, c := range commands {
			if !wantAliases[c.Alias] {
				t.Fatalf("unexpected command in results: %q", c.Alias)
			}
		}
	})

	t.Run("scopes to explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "list-tags-other")
		otherCmd, err := repo.Create(CommandCreateConfig{Alias: "other-tagged", Template: "echo hi", ProfileID: &other.ID})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if _, err := repo.AddTags(otherCmd, []models.Tag{tagA}); err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}

		commands, err := repo.ListByTags([]models.Tag{tagA}, "", 0, &other.ID)
		if err != nil {
			t.Fatalf("ListByTags() error = %v", err)
		}
		if len(commands) != 1 || commands[0].Alias != "other-tagged" {
			t.Fatalf("commands = %+v, want only [other-tagged]", commands)
		}
	})

	t.Run("invalid order field returns error", func(t *testing.T) {
		_, err := repo.ListByTags([]models.Tag{tagA}, "bogus", 0, nil)
		if err == nil {
			t.Fatalf("ListByTags() error = nil, want error")
		}
	})
}

// --- Search ---

func TestCommandRepositorySearch(t *testing.T) {
	repo, _, db := setupCommandRepositoryTest(t)
	_ = mustCreateCommand(t, repo, CommandCreateConfig{Alias: "apple", Template: "echo apple"})
	_ = mustCreateCommand(t, repo, CommandCreateConfig{Alias: "green-apple", Template: "echo green apple"})
	_ = mustCreateCommand(t, repo, CommandCreateConfig{Alias: "banana", Template: "echo banana"})

	t.Run("uses default fields when none provided", func(t *testing.T) {
		commands, err := repo.Search("apple", nil, 0, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(commands) != 2 {
			t.Fatalf("len(commands) = %d, want 2; commands = %+v", len(commands), commands)
		}
		if commands[0].Alias != "apple" {
			t.Fatalf("commands[0].Alias = %q, want %q (closest match first)", commands[0].Alias, "apple")
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		commands, err := repo.Search("apple", nil, 1, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(commands) != 1 {
			t.Fatalf("len(commands) = %d, want 1", len(commands))
		}
	})

	t.Run("scopes to explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "search-other")
		if _, err := repo.Create(CommandCreateConfig{Alias: "apple-elsewhere", Template: "echo apple", ProfileID: &other.ID}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		commands, err := repo.Search("apple", nil, 0, &other.ID)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(commands) != 1 || commands[0].Alias != "apple-elsewhere" {
			t.Fatalf("commands = %+v, want only [apple-elsewhere]", commands)
		}
	})

	t.Run("invalid explicit field returns error", func(t *testing.T) {
		_, err := repo.Search("apple", []string{"bogus"}, 0, nil)
		if err == nil {
			t.Fatalf("Search() error = nil, want error")
		}
	})

	t.Run("no matches returns empty slice", func(t *testing.T) {
		commands, err := repo.Search("nonexistent", nil, 0, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(commands) != 0 {
			t.Fatalf("commands = %+v, want empty", commands)
		}
	})
}

// --- unexported helpers ---

func TestEncodeEnv(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		got, err := encodeEnv(nil)
		if err != nil {
			t.Fatalf("encodeEnv(nil) error = %v", err)
		}
		if got != nil {
			t.Fatalf("encodeEnv(nil) = %v, want nil", *got)
		}
	})

	t.Run("empty map returns nil", func(t *testing.T) {
		got, err := encodeEnv(map[string]string{})
		if err != nil {
			t.Fatalf("encodeEnv({}) error = %v", err)
		}
		if got != nil {
			t.Fatalf("encodeEnv({}) = %v, want nil", *got)
		}
	})

	t.Run("non-empty map encodes to json", func(t *testing.T) {
		env := map[string]string{"A": "1", "B": "2"}
		got, err := encodeEnv(env)
		if err != nil {
			t.Fatalf("encodeEnv() error = %v", err)
		}
		want, _ := json.Marshal(env)
		if got == nil || *got != string(want) {
			t.Fatalf("encodeEnv() = %v, want %q", got, string(want))
		}
	})
}

func TestNewCommandRepositoryDefaultValidator(t *testing.T) {
	_, profileRepo, db := setupCommandRepositoryTest(t)
	repo := NewCommandRepository(db, profileRepo, nil)

	_, err := repo.Create(CommandCreateConfig{Alias: "help", Template: "echo hi"})
	if !errors.Is(err, validate.ErrValidation) {
		t.Fatalf("Create() with default validator error = %v, want ErrValidation (reserved alias)", err)
	}
}

func TestCommandRepositoryAddTagsPropagatesTransactionError(t *testing.T) {
	repo, _, _ := setupCommandRepositoryTest(t)
	cmd := mustCreateCommand(t, repo, CommandCreateConfig{Alias: "tag-fail", Template: "echo hi"})

	// Reference a tag id that was never persisted; the foreign key on
	// CommandTag.TagID makes the insert fail, and the error should be wrapped.
	bogusTag := models.Tag{ID: 999999, Name: "bogus"}
	_, err := repo.AddTags(cmd, []models.Tag{bogusTag})
	if !errors.Is(err, ErrTagAttachFailed) {
		t.Fatalf("AddTags() error = %v, want ErrTagAttachFailed", err)
	}
}

func TestValidateNoClearAndSet(t *testing.T) {
	cwd := "/tmp"
	shell := "/bin/bash"
	timeout := 5
	env := map[string]string{"A": "1"}

	cases := []struct {
		name    string
		input   CommandUpdateConfig
		wantErr bool
	}{
		{name: "no conflicts", input: CommandUpdateConfig{}, wantErr: false},
		{name: "cwd set without clear", input: CommandUpdateConfig{Cwd: &cwd}, wantErr: false},
		{name: "cwd conflict", input: CommandUpdateConfig{ClearCwd: true, Cwd: &cwd}, wantErr: true},
		{name: "shell conflict", input: CommandUpdateConfig{ClearShell: true, Shell: &shell}, wantErr: true},
		{name: "timeout conflict", input: CommandUpdateConfig{ClearTimeout: true, Timeout: &timeout}, wantErr: true},
		{name: "env conflict", input: CommandUpdateConfig{ClearEnv: true, Env: env}, wantErr: true},
		{name: "clear without set is fine", input: CommandUpdateConfig{ClearCwd: true, ClearShell: true, ClearTimeout: true, ClearEnv: true}, wantErr: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateNoClearAndSet(tc.input)
			if tc.wantErr && !errors.Is(err, ErrConflictingClearAndSet) {
				t.Fatalf("validateNoClearAndSet() error = %v, want ErrConflictingClearAndSet", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateNoClearAndSet() error = %v, want nil", err)
			}
		})
	}
}
