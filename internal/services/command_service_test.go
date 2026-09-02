package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
)

func setupCommandServiceTest(t *testing.T) (CommandService, repository.CommandRepository, repository.TagRepository, repository.ProfileRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:command-service-%d?mode=memory&cache=shared", time.Now().UnixNano())
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
	defaultProfile := models.Profile{Name: repository.DefaultProfileName, Description: &defaultDescription}
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

	profileRepo := repository.NewProfileRepository(db, validate.NewProfileValidator(nil))
	commandRepo := repository.NewCommandRepository(db, profileRepo, validate.NewCommandValidator(nil))
	tagRepo := repository.NewTagRepository(db, validate.NewTagValidator(nil))

	svc := NewCommandService(db, commandRepo, tagRepo, profileRepo)
	return svc, commandRepo, tagRepo, profileRepo, db
}

func mustCreateServiceProfile(t *testing.T, db *gorm.DB, name string) models.Profile {
	t.Helper()
	profile := models.Profile{Name: name}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("create profile %q: %v", name, err)
	}
	return profile
}

func mustCreateServiceTag(t *testing.T, tagRepo repository.TagRepository, name string) *models.Tag {
	t.Helper()
	tag, err := tagRepo.Create(repository.TagCreateConfig{Name: name})
	if err != nil {
		t.Fatalf("create tag %q: %v", name, err)
	}
	return tag
}

func mustCreateServiceCommand(t *testing.T, svc CommandService, input CreateCommandConfig) *models.Command {
	t.Helper()
	cmd, err := svc.CreateCommand(input)
	if err != nil {
		t.Fatalf("CreateCommand(%+v) error = %v", input, err)
	}
	return cmd
}

func strPtr(s string) *string {
	return &s
}

func timeoutPtr(v int) *int {
	return &v
}

// --- CreateCommand ---

func TestCommandServiceCreateCommand(t *testing.T) {
	t.Run("creates command using active profile", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		cmd, err := svc.CreateCommand(CreateCommandConfig{Alias: "build", Template: "go build ./..."})
		if err != nil {
			t.Fatalf("CreateCommand() error = %v", err)
		}
		if cmd.ID == 0 {
			t.Fatalf("CreateCommand() returned zero ID")
		}
	})

	t.Run("creates command using named profile", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		other := mustCreateServiceProfile(t, db, "other-profile")

		cmd, err := svc.CreateCommand(CreateCommandConfig{Alias: "deploy", Template: "echo deploy", ProfileName: strPtr("other-profile")})
		if err != nil {
			t.Fatalf("CreateCommand() error = %v", err)
		}
		if cmd.ProfileID != other.ID {
			t.Fatalf("ProfileID = %d, want %d", cmd.ProfileID, other.ID)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.CreateCommand(CreateCommandConfig{Alias: "build", Template: "echo hi", ProfileName: strPtr("nope")})
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("CreateCommand() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown tag returns error and does not create command", func(t *testing.T) {
		svc, commandRepo, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.CreateCommand(CreateCommandConfig{Alias: "tagged", Template: "echo hi", Tags: []string{"missing-tag"}})
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("CreateCommand() error = %v, want ErrUnknownTagName", err)
		}
		if _, getErr := commandRepo.GetByAlias("tagged", nil); !errors.Is(getErr, repository.ErrUnknownAlias) {
			t.Fatalf("command was persisted despite tag lookup failure: err = %v", getErr)
		}
	})

	t.Run("attaches tags within the same transaction", func(t *testing.T) {
		svc, _, tagRepo, _, db := setupCommandServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "tag-a")
		mustCreateServiceTag(t, tagRepo, "tag-b")

		cmd := mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "tagged", Template: "echo hi", Tags: []string{"tag-a", "tag-b"}})

		var count int64
		if err := db.Model(&models.CommandTag{}).Where("command_id = ?", cmd.ID).Count(&count).Error; err != nil {
			t.Fatalf("count command_tags: %v", err)
		}
		if count != 2 {
			t.Fatalf("command_tags count = %d, want 2", count)
		}
	})

	t.Run("reserved alias validation error propagates and rolls back", func(t *testing.T) {
		svc, commandRepo, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.CreateCommand(CreateCommandConfig{Alias: "help", Template: "echo hi"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("CreateCommand() error = %v, want ErrValidation", err)
		}
		if _, getErr := commandRepo.GetByAlias("help", nil); !errors.Is(getErr, repository.ErrUnknownAlias) {
			t.Fatalf("command was persisted despite validation failure: err = %v", getErr)
		}
	})

	t.Run("duplicate alias error rolls back tag attachment", func(t *testing.T) {
		svc, _, tagRepo, _, db := setupCommandServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "dup-tag")
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "dup", Template: "echo one"})

		_, err := svc.CreateCommand(CreateCommandConfig{Alias: "dup", Template: "echo two", Tags: []string{"dup-tag"}})
		if !errors.Is(err, repository.ErrAliasConflict) {
			t.Fatalf("CreateCommand() error = %v, want ErrAliasConflict", err)
		}

		var count int64
		if err := db.Model(&models.CommandTag{}).Count(&count).Error; err != nil {
			t.Fatalf("count command_tags: %v", err)
		}
		if count != 0 {
			t.Fatalf("command_tags count = %d, want 0 after rollback", count)
		}
	})

	t.Run("stores optional fields and env", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		cwd := "/tmp"
		env := map[string]string{"A": "1"}
		cmd, err := svc.CreateCommand(CreateCommandConfig{
			Alias:    "full",
			Template: "echo hi",
			Cwd:      &cwd,
			Env:      env,
			Timeout:  timeoutPtr(30),
		})
		if err != nil {
			t.Fatalf("CreateCommand() error = %v", err)
		}
		if cmd.Cwd == nil || *cmd.Cwd != cwd {
			t.Fatalf("Cwd = %v, want %q", cmd.Cwd, cwd)
		}
		if cmd.Timeout == nil || *cmd.Timeout != 30 {
			t.Fatalf("Timeout = %v, want 30", cmd.Timeout)
		}
		if cmd.Env == nil {
			t.Fatalf("Env = nil, want encoded json")
		}
	})
}

// --- UpdateCommand ---

func TestCommandServiceUpdateCommand(t *testing.T) {
	t.Run("updates existing command", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "rename-me", Template: "echo old"})

		newAlias := "renamed"
		updated, err := svc.UpdateCommand("rename-me", nil, UpdateCommandConfig{NewAlias: &newAlias})
		if err != nil {
			t.Fatalf("UpdateCommand() error = %v", err)
		}
		if updated.Alias != "renamed" {
			t.Fatalf("Alias = %q, want %q", updated.Alias, "renamed")
		}
	})

	t.Run("unknown alias returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.UpdateCommand("missing", nil, UpdateCommandConfig{})
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("UpdateCommand() error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("propagates validation error from repository", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "renaming", Template: "echo hi"})

		newAlias := "help"
		_, err := svc.UpdateCommand("renaming", nil, UpdateCommandConfig{NewAlias: &newAlias})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("UpdateCommand() error = %v, want ErrValidation", err)
		}
	})

	t.Run("propagates conflicting clear and set error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "conflict-me", Template: "echo hi"})

		cwd := "/tmp"
		_, err := svc.UpdateCommand("conflict-me", nil, UpdateCommandConfig{Cwd: &cwd, ClearCwd: true})
		if !errors.Is(err, repository.ErrConflictingClearAndSet) {
			t.Fatalf("UpdateCommand() error = %v, want ErrConflictingClearAndSet", err)
		}
	})

	t.Run("clears fields", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		cwd := "/tmp"
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "clear-me", Template: "echo hi", Cwd: &cwd})

		updated, err := svc.UpdateCommand("clear-me", nil, UpdateCommandConfig{ClearCwd: true})
		if err != nil {
			t.Fatalf("UpdateCommand() error = %v", err)
		}
		if updated.Cwd != nil {
			t.Fatalf("Cwd = %v, want nil", *updated.Cwd)
		}
	})
}

// --- DeleteCommand ---

func TestCommandServiceDeleteCommand(t *testing.T) {
	t.Run("deletes existing command", func(t *testing.T) {
		svc, commandRepo, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "delete-me", Template: "echo hi"})

		if err := svc.DeleteCommand("delete-me", nil); err != nil {
			t.Fatalf("DeleteCommand() error = %v", err)
		}
		if _, err := commandRepo.GetByAlias("delete-me", nil); !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("command still exists after delete: err = %v", err)
		}
	})

	t.Run("unknown alias returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		err := svc.DeleteCommand("missing", nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("DeleteCommand() error = %v, want ErrUnknownAlias", err)
		}
	})
}

// --- AddTags / RemoveTags ---

func TestCommandServiceAddTags(t *testing.T) {
	t.Run("adds new tags", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "tag-me", Template: "echo hi"})
		mustCreateServiceTag(t, tagRepo, "a")
		mustCreateServiceTag(t, tagRepo, "b")

		result, err := svc.AddTags("tag-me", []string{"a", "b"}, nil)
		if err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}
		if len(result.Added) != 2 {
			t.Fatalf("Added = %v, want 2 tags", result.Added)
		}
	})

	t.Run("unknown command returns error", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupCommandServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "a")

		_, err := svc.AddTags("missing", []string{"a"}, nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("AddTags() error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("unknown tag returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "tag-me", Template: "echo hi"})

		_, err := svc.AddTags("tag-me", []string{"missing"}, nil)
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("AddTags() error = %v, want ErrUnknownTagName", err)
		}
	})
}

func TestCommandServiceRemoveTags(t *testing.T) {
	t.Run("removes attached tags", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "untag-me", Template: "echo hi"})
		mustCreateServiceTag(t, tagRepo, "a")
		if _, err := svc.AddTags("untag-me", []string{"a"}, nil); err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}

		result, err := svc.RemoveTags("untag-me", []string{"a"}, nil)
		if err != nil {
			t.Fatalf("RemoveTags() error = %v", err)
		}
		if len(result.Removed) != 1 || result.Removed[0] != "a" {
			t.Fatalf("Removed = %v, want [a]", result.Removed)
		}
	})

	t.Run("unknown command returns error", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupCommandServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "a")

		_, err := svc.RemoveTags("missing", []string{"a"}, nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("RemoveTags() error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("unknown tag returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "untag-me", Template: "echo hi"})

		_, err := svc.RemoveTags("untag-me", []string{"missing"}, nil)
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("RemoveTags() error = %v, want ErrUnknownTagName", err)
		}
	})
}

// --- GetCommand / GetCommandOrNone / GetCommandByID ---

func TestCommandServiceGetCommand(t *testing.T) {
	t.Run("returns command from active profile", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		created := mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})

		got, err := svc.GetCommand("build", nil)
		if err != nil {
			t.Fatalf("GetCommand() error = %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("GetCommand() id = %d, want %d", got.ID, created.ID)
		}
	})

	t.Run("unknown alias returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.GetCommand("missing", nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("GetCommand() error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.GetCommand("build", strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetCommand() error = %v, want ErrProfileNotFound", err)
		}
	})
}

func TestCommandServiceGetCommandOrNone(t *testing.T) {
	t.Run("returns nil, nil when alias missing", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		cmd, err := svc.GetCommandOrNil("missing", nil)
		if err != nil {
			t.Fatalf("GetCommandOrNone() error = %v, want nil", err)
		}
		if cmd != nil {
			t.Fatalf("GetCommandOrNone() = %v, want nil", cmd)
		}
	})

	t.Run("returns command when found", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		created := mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})

		cmd, err := svc.GetCommandOrNil("build", nil)
		if err != nil {
			t.Fatalf("GetCommandOrNone() error = %v", err)
		}
		if cmd == nil || cmd.ID != created.ID {
			t.Fatalf("GetCommandOrNone() = %v, want id %d", cmd, created.ID)
		}
	})

	t.Run("does not swallow non-alias errors", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.GetCommandOrNil("build", strPtr("nope"))
		if err == nil {
			t.Fatalf("GetCommandOrNone() error = nil, want error resolving profile")
		}
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetCommandOrNone() error = %v, want ErrProfileNotFound", err)
		}
	})
}

func TestCommandServiceGetCommandByID(t *testing.T) {
	t.Run("returns command when found", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		created := mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})

		got, err := svc.GetCommandByID(created.ID, nil)
		if err != nil {
			t.Fatalf("GetCommandByID() error = %v", err)
		}
		if got.Alias != "build" {
			t.Fatalf("Alias = %q, want %q", got.Alias, "build")
		}
	})

	t.Run("unknown id returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.GetCommandByID(999999, nil)
		if !errors.Is(err, repository.ErrUnknownCommand) {
			t.Fatalf("GetCommandByID() error = %v, want ErrUnknownCommand", err)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		created := mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})

		_, err := svc.GetCommandByID(created.ID, strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetCommandByID() error = %v, want ErrProfileNotFound", err)
		}
	})
}

// --- ListCommands ---

func TestCommandServiceListCommands(t *testing.T) {
	t.Run("lists all commands with default ordering", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "zeta", Template: "echo z"})
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "alpha", Template: "echo a"})

		commands, err := svc.ListCommands("", nil, nil, nil)
		if err != nil {
			t.Fatalf("ListCommands() error = %v", err)
		}
		if len(commands) != 2 {
			t.Fatalf("len(commands) = %d, want 2", len(commands))
		}
		if commands[0].Alias != "alpha" {
			t.Fatalf("commands[0].Alias = %q, want %q", commands[0].Alias, "alpha")
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "a", Template: "echo a"})
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "b", Template: "echo b"})

		limit := 1
		commands, err := svc.ListCommands("", nil, &limit, nil)
		if err != nil {
			t.Fatalf("ListCommands() error = %v", err)
		}
		if len(commands) != 1 {
			t.Fatalf("len(commands) = %d, want 1", len(commands))
		}
	})

	t.Run("filters by tags", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupCommandServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "keep")
		cmdA := mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "has-tag", Template: "echo a"})
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "no-tag", Template: "echo b"})
		if _, err := svc.AddTags(cmdA.Alias, []string{"keep"}, nil); err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}

		commands, err := svc.ListCommands("", []string{"keep"}, nil, nil)
		if err != nil {
			t.Fatalf("ListCommands() error = %v", err)
		}
		if len(commands) != 1 || commands[0].Alias != "has-tag" {
			t.Fatalf("commands = %+v, want only [has-tag]", commands)
		}
	})

	t.Run("unknown tag filter returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.ListCommands("", []string{"missing"}, nil, nil)
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("ListCommands() error = %v, want ErrUnknownTagName", err)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.ListCommands("", nil, nil, strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("ListCommands() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("invalid order field returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.ListCommands("bogus", nil, nil, nil)
		if err == nil {
			t.Fatalf("ListCommands() error = nil, want error")
		}
	})
}

// --- SearchCommands ---

func TestCommandServiceSearchCommands(t *testing.T) {
	t.Run("uses default fields", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "apple", Template: "echo apple"})
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "banana", Template: "echo banana"})

		commands, err := svc.SearchCommands("apple", nil, nil, nil)
		if err != nil {
			t.Fatalf("SearchCommands() error = %v", err)
		}
		if len(commands) != 1 || commands[0].Alias != "apple" {
			t.Fatalf("commands = %+v, want only [apple]", commands)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "apple", Template: "echo apple"})
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "apple-two", Template: "echo apple"})

		limit := 1
		commands, err := svc.SearchCommands("apple", nil, &limit, nil)
		if err != nil {
			t.Fatalf("SearchCommands() error = %v", err)
		}
		if len(commands) != 1 {
			t.Fatalf("len(commands) = %d, want 1", len(commands))
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.SearchCommands("apple", nil, nil, strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("SearchCommands() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("invalid explicit field returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.SearchCommands("apple", []string{"bogus"}, nil, nil)
		if err == nil {
			t.Fatalf("SearchCommands() error = nil, want error")
		}
	})
}

// --- MoveCommand ---

func TestCommandServiceMoveCommand(t *testing.T) {
	t.Run("nil target profile returns ErrNoTargetProfile", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.MoveCommand("build", nil, nil)
		if !errors.Is(err, ErrNoTargetProfile) {
			t.Fatalf("MoveCommand() error = %v, want ErrNoTargetProfile", err)
		}
	})

	t.Run("moves command to target profile", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		target := mustCreateServiceProfile(t, db, "target-profile")
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})

		moved, err := svc.MoveCommand("build", strPtr("target-profile"), nil)
		if err != nil {
			t.Fatalf("MoveCommand() error = %v", err)
		}
		if moved.ProfileID != target.ID {
			t.Fatalf("ProfileID = %d, want %d", moved.ProfileID, target.ID)
		}

		got, err := svc.GetCommand("build", strPtr("target-profile"))
		if err != nil {
			t.Fatalf("GetCommand() after move error = %v", err)
		}
		if got.ID != moved.ID {
			t.Fatalf("GetCommand() id = %d, want %d", got.ID, moved.ID)
		}
	})

	t.Run("unknown target profile returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})

		_, err := svc.MoveCommand("build", strPtr("nope"), nil)
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("MoveCommand() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown source profile returns error", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})

		_, err := svc.MoveCommand("build", strPtr("target-profile"), strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("MoveCommand() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown alias returns error", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")

		_, err := svc.MoveCommand("missing", strPtr("target-profile"), nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("MoveCommand() error = %v, want ErrUnknownAlias", err)
		}
	})
}

// --- CopyCommand ---

func TestCommandServiceCopyCommand(t *testing.T) {
	t.Run("nil target profile returns ErrNoTargetProfile", func(t *testing.T) {
		svc, _, _, _, _ := setupCommandServiceTest(t)
		_, err := svc.CopyCommand("build", nil, nil, nil)
		if !errors.Is(err, ErrNoTargetProfile) {
			t.Fatalf("CopyCommand() error = %v, want ErrNoTargetProfile", err)
		}
	})

	t.Run("copies command with same alias into target profile", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		original := mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})

		copied, err := svc.CopyCommand("build", strPtr("target-profile"), nil, nil)
		if err != nil {
			t.Fatalf("CopyCommand() error = %v", err)
		}
		if copied.ID == original.ID {
			t.Fatalf("CopyCommand() returned same command, want a new one")
		}
		if copied.Alias != "build" {
			t.Fatalf("Alias = %q, want %q", copied.Alias, "build")
		}

		// original is untouched
		stillThere, err := svc.GetCommand("build", nil)
		if err != nil {
			t.Fatalf("GetCommand(original) error = %v", err)
		}
		if stillThere.ID != original.ID {
			t.Fatalf("original command id changed: got %d, want %d", stillThere.ID, original.ID)
		}
	})

	t.Run("copies command with a new alias", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})

		copied, err := svc.CopyCommand("build", strPtr("target-profile"), strPtr("build-copy"), nil)
		if err != nil {
			t.Fatalf("CopyCommand() error = %v", err)
		}
		if copied.Alias != "build-copy" {
			t.Fatalf("Alias = %q, want %q", copied.Alias, "build-copy")
		}
	})

	t.Run("copies command with no env set without panicking", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "no-env", Template: "echo hi"})

		copied, err := svc.CopyCommand("no-env", strPtr("target-profile"), nil, nil)
		if err != nil {
			t.Fatalf("CopyCommand() error = %v", err)
		}
		if copied.Env != nil {
			t.Fatalf("Env = %v, want nil", *copied.Env)
		}
	})

	t.Run("copies env correctly", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		env := map[string]string{"FOO": "bar"}
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "with-env", Template: "echo hi", Env: env})

		copied, err := svc.CopyCommand("with-env", strPtr("target-profile"), nil, nil)
		if err != nil {
			t.Fatalf("CopyCommand() error = %v", err)
		}
		if copied.Env == nil {
			t.Fatalf("Env = nil, want encoded json")
		}
		want, _ := json.Marshal(env)
		if *copied.Env != string(want) {
			t.Fatalf("Env = %q, want %q", *copied.Env, string(want))
		}
	})

	t.Run("malformed stored env returns error instead of panicking", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		cmd := mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "bad-env", Template: "echo hi"})

		if err := db.Model(&models.Command{}).Where("id = ?", cmd.ID).Update("env", "not-json").Error; err != nil {
			t.Fatalf("corrupt env: %v", err)
		}

		_, err := svc.CopyCommand("bad-env", strPtr("target-profile"), nil, nil)
		if err == nil {
			t.Fatalf("CopyCommand() error = nil, want json parse error")
		}
	})

	t.Run("unknown source profile returns error", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")

		_, err := svc.CopyCommand("build", strPtr("target-profile"), nil, strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("CopyCommand() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown alias returns error", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")

		_, err := svc.CopyCommand("missing", strPtr("target-profile"), nil, nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("CopyCommand() error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("alias conflict in target profile returns error", func(t *testing.T) {
		svc, _, _, _, db := setupCommandServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo hi"})
		mustCreateServiceCommand(t, svc, CreateCommandConfig{Alias: "build", Template: "echo other", ProfileName: strPtr("target-profile")})

		_, err := svc.CopyCommand("build", strPtr("target-profile"), nil, nil)
		if !errors.Is(err, repository.ErrAliasConflict) {
			t.Fatalf("CopyCommand() error = %v, want ErrAliasConflict", err)
		}
	})
}

// --- parseEnv ---

func TestParseEnv(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		got, err := parseEnv("")
		if err != nil {
			t.Fatalf("parseEnv(\"\") error = %v", err)
		}
		if got != nil {
			t.Fatalf("parseEnv(\"\") = %v, want nil", got)
		}
	})

	t.Run("valid json decodes to map", func(t *testing.T) {
		got, err := parseEnv(`{"A":"1","B":"2"}`)
		if err != nil {
			t.Fatalf("parseEnv() error = %v", err)
		}
		if got["A"] != "1" || got["B"] != "2" {
			t.Fatalf("parseEnv() = %v, want {A:1 B:2}", got)
		}
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := parseEnv("not-json")
		if err == nil {
			t.Fatalf("parseEnv(\"not-json\") error = nil, want error")
		}
	})
}
