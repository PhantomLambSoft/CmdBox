package services

import (
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

func setupVariableServiceTest(t *testing.T) (VariableService, repository.VariableRepository, repository.TagRepository, repository.ProfileRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:variable-service-%d?mode=memory&cache=shared", time.Now().UnixNano())
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
		&models.VariableTag{},
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
	variableRepo := repository.NewVariableRepository(db, profileRepo, validate.NewVariableValidator(nil))
	tagRepo := repository.NewTagRepository(db, validate.NewTagValidator(nil))

	svc := NewVariableService(db, variableRepo, tagRepo, profileRepo)
	return svc, variableRepo, tagRepo, profileRepo, db
}

func mustCreateServiceVariable(t *testing.T, svc VariableService, input CreateVariableConfig) *models.Variable {
	t.Helper()
	v, err := svc.CreateVariable(input)
	if err != nil {
		t.Fatalf("CreateVariable(%+v) error = %v", input, err)
	}
	return v
}

// --- CreateVariable ---

func TestVariableServiceCreateVariable(t *testing.T) {
	t.Run("creates variable using active profile", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		v, err := svc.CreateVariable(CreateVariableConfig{Name: "greeting", Value: "hello"})
		if err != nil {
			t.Fatalf("CreateVariable() error = %v", err)
		}
		if v.ID == 0 {
			t.Fatalf("CreateVariable() returned zero ID")
		}
	})

	t.Run("creates variable using named profile", func(t *testing.T) {
		svc, _, _, _, db := setupVariableServiceTest(t)
		other := mustCreateServiceProfile(t, db, "other-profile")

		v, err := svc.CreateVariable(CreateVariableConfig{Name: "deploy_target", Value: "prod", ProfileName: strPtr("other-profile")})
		if err != nil {
			t.Fatalf("CreateVariable() error = %v", err)
		}
		if v.ProfileID != other.ID {
			t.Fatalf("ProfileID = %d, want %d", v.ProfileID, other.ID)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.CreateVariable(CreateVariableConfig{Name: "greeting", Value: "hello", ProfileName: strPtr("nope")})
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("CreateVariable() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown tag returns error and does not create variable", func(t *testing.T) {
		svc, variableRepo, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.CreateVariable(CreateVariableConfig{Name: "tagged", Value: "x", Tags: []string{"missing-tag"}})
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("CreateVariable() error = %v, want ErrUnknownTagName", err)
		}
		if _, getErr := variableRepo.GetByName("tagged", nil); !errors.Is(getErr, repository.ErrUnKnownName) {
			t.Fatalf("variable was persisted despite tag lookup failure: err = %v", getErr)
		}
	})

	t.Run("attaches tags within the same transaction", func(t *testing.T) {
		svc, _, tagRepo, _, db := setupVariableServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "tag-a")
		mustCreateServiceTag(t, tagRepo, "tag-b")

		v := mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "tagged", Value: "x", Tags: []string{"tag-a", "tag-b"}})

		var count int64
		if err := db.Model(&models.VariableTag{}).Where("variable_id = ?", v.ID).Count(&count).Error; err != nil {
			t.Fatalf("count variable_tags: %v", err)
		}
		if count != 2 {
			t.Fatalf("variable_tags count = %d, want 2", count)
		}
	})

	t.Run("reserved name validation error propagates and rolls back", func(t *testing.T) {
		svc, variableRepo, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.CreateVariable(CreateVariableConfig{Name: "help", Value: "x"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("CreateVariable() error = %v, want ErrValidation", err)
		}
		if _, getErr := variableRepo.GetByName("help", nil); !errors.Is(getErr, repository.ErrUnKnownName) {
			t.Fatalf("variable was persisted despite validation failure: err = %v", getErr)
		}
	})

	t.Run("duplicate name error rolls back tag attachment", func(t *testing.T) {
		svc, _, tagRepo, _, db := setupVariableServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "dup-tag")
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "dup", Value: "one"})

		_, err := svc.CreateVariable(CreateVariableConfig{Name: "dup", Value: "two", Tags: []string{"dup-tag"}})
		if !errors.Is(err, repository.ErrNameConflict) {
			t.Fatalf("CreateVariable() error = %v, want ErrNameConflict", err)
		}

		var count int64
		if err := db.Model(&models.VariableTag{}).Count(&count).Error; err != nil {
			t.Fatalf("count variable_tags: %v", err)
		}
		if count != 0 {
			t.Fatalf("variable_tags count = %d, want 0 after rollback", count)
		}
	})
}

// --- UpdateVariable ---

func TestVariableServiceUpdateVariable(t *testing.T) {
	t.Run("updates existing variable", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "rename-me", Value: "old"})

		newName := "renamed"
		updated, err := svc.UpdateVariable("rename-me", nil, UpdateVariableConfig{NewName: &newName})
		if err != nil {
			t.Fatalf("UpdateVariable() error = %v", err)
		}
		if updated.Name != "renamed" {
			t.Fatalf("Name = %q, want %q", updated.Name, "renamed")
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.UpdateVariable("missing", nil, UpdateVariableConfig{})
		if !errors.Is(err, repository.ErrUnKnownName) {
			t.Fatalf("UpdateVariable() error = %v, want ErrUnKnownName", err)
		}
	})

	t.Run("propagates validation error from repository", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "renaming", Value: "x"})

		newName := "help"
		_, err := svc.UpdateVariable("renaming", nil, UpdateVariableConfig{NewName: &newName})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("UpdateVariable() error = %v, want ErrValidation", err)
		}
	})

	t.Run("updates value", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "value-me", Value: "old"})

		newValue := "new"
		updated, err := svc.UpdateVariable("value-me", nil, UpdateVariableConfig{Value: &newValue})
		if err != nil {
			t.Fatalf("UpdateVariable() error = %v", err)
		}
		if updated.Value != "new" {
			t.Fatalf("Value = %q, want %q", updated.Value, "new")
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "build", Value: "x"})

		_, err := svc.UpdateVariable("build", strPtr("nope"), UpdateVariableConfig{})
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("UpdateVariable() error = %v, want ErrProfileNotFound", err)
		}
	})
}

// --- DeleteVariable ---

func TestVariableServiceDeleteVariable(t *testing.T) {
	t.Run("deletes existing variable", func(t *testing.T) {
		svc, variableRepo, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "delete-me", Value: "x"})

		if err := svc.DeleteVariable("delete-me", nil); err != nil {
			t.Fatalf("DeleteVariable() error = %v", err)
		}
		if _, err := variableRepo.GetByName("delete-me", nil); !errors.Is(err, repository.ErrUnKnownName) {
			t.Fatalf("variable still exists after delete: err = %v", err)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		err := svc.DeleteVariable("missing", nil)
		if !errors.Is(err, repository.ErrUnKnownName) {
			t.Fatalf("DeleteVariable() error = %v, want ErrUnKnownName", err)
		}
	})
}

// --- AddTags / RemoveTags ---

func TestVariableServiceAddTags(t *testing.T) {
	t.Run("adds new tags", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "tag-me", Value: "x"})
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

	t.Run("unknown variable returns error", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupVariableServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "a")

		_, err := svc.AddTags("missing", []string{"a"}, nil)
		if !errors.Is(err, repository.ErrUnKnownName) {
			t.Fatalf("AddTags() error = %v, want ErrUnKnownName", err)
		}
	})

	t.Run("unknown tag returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "tag-me", Value: "x"})

		_, err := svc.AddTags("tag-me", []string{"missing"}, nil)
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("AddTags() error = %v, want ErrUnknownTagName", err)
		}
	})
}

func TestVariableServiceRemoveTags(t *testing.T) {
	t.Run("removes attached tags", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "untag-me", Value: "x"})
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

	t.Run("unknown variable returns error", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupVariableServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "a")

		_, err := svc.RemoveTags("missing", []string{"a"}, nil)
		if !errors.Is(err, repository.ErrUnKnownName) {
			t.Fatalf("RemoveTags() error = %v, want ErrUnKnownName", err)
		}
	})

	t.Run("unknown tag returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "untag-me", Value: "x"})

		_, err := svc.RemoveTags("untag-me", []string{"missing"}, nil)
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("RemoveTags() error = %v, want ErrUnknownTagName", err)
		}
	})
}

// --- GetVariable / GetCommandOrNil / GetVariableByID ---

func TestVariableServiceGetVariable(t *testing.T) {
	t.Run("returns variable from active profile", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		created := mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})

		got, err := svc.GetVariable("greeting", nil)
		if err != nil {
			t.Fatalf("GetVariable() error = %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("GetVariable() id = %d, want %d", got.ID, created.ID)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.GetVariable("missing", nil)
		if !errors.Is(err, repository.ErrUnKnownName) {
			t.Fatalf("GetVariable() error = %v, want ErrUnKnownName", err)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.GetVariable("greeting", strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetVariable() error = %v, want ErrProfileNotFound", err)
		}
	})
}

func TestVariableServiceGetVariableOrNil(t *testing.T) {
	t.Run("returns nil, nil when name missing", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		v, err := svc.GetVariableOrNil("missing", nil)
		if err != nil {
			t.Fatalf("GetVariableOrNil() error = %v, want nil", err)
		}
		if v != nil {
			t.Fatalf("GetVariableOrNil() = %v, want nil", v)
		}
	})

	t.Run("returns variable when found", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		created := mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})

		v, err := svc.GetVariableOrNil("greeting", nil)
		if err != nil {
			t.Fatalf("GetVariableOrNil() error = %v", err)
		}
		if v == nil || v.ID != created.ID {
			t.Fatalf("GetVariableOrNil() = %v, want id %d", v, created.ID)
		}
	})

	t.Run("does not swallow non-name errors", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.GetVariableOrNil("greeting", strPtr("nope"))
		if err == nil {
			t.Fatalf("GetVariableOrNil() error = nil, want error resolving profile")
		}
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetVariableOrNil() error = %v, want ErrProfileNotFound", err)
		}
	})
}

func TestVariableServiceGetVariableByID(t *testing.T) {
	t.Run("returns variable when found", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		created := mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})

		got, err := svc.GetVariableByID(created.ID, nil)
		if err != nil {
			t.Fatalf("GetVariableByID() error = %v", err)
		}
		if got.Name != "greeting" {
			t.Fatalf("Name = %q, want %q", got.Name, "greeting")
		}
	})

	t.Run("unknown id returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.GetVariableByID(999999, nil)
		if !errors.Is(err, repository.ErrUnknownVariable) {
			t.Fatalf("GetVariableByID() error = %v, want ErrUnknownVariable", err)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		created := mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})

		_, err := svc.GetVariableByID(created.ID, strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetVariableByID() error = %v, want ErrProfileNotFound", err)
		}
	})
}

// --- ListVariables ---

func TestVariableServiceListVariables(t *testing.T) {
	t.Run("lists all variables with default ordering", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "zeta", Value: "z"})
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "alpha", Value: "a"})

		variables, err := svc.ListVariables("", nil, nil, nil)
		if err != nil {
			t.Fatalf("ListVariables() error = %v", err)
		}
		if len(variables) != 2 {
			t.Fatalf("len(variables) = %d, want 2", len(variables))
		}
		if variables[0].Name != "alpha" {
			t.Fatalf("variables[0].Name = %q, want %q", variables[0].Name, "alpha")
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "a", Value: "1"})
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "b", Value: "2"})

		limit := 1
		variables, err := svc.ListVariables("", nil, &limit, nil)
		if err != nil {
			t.Fatalf("ListVariables() error = %v", err)
		}
		if len(variables) != 1 {
			t.Fatalf("len(variables) = %d, want 1", len(variables))
		}
	})

	t.Run("filters by tags", func(t *testing.T) {
		svc, _, tagRepo, _, _ := setupVariableServiceTest(t)
		mustCreateServiceTag(t, tagRepo, "keep")
		varA := mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "has-tag", Value: "a"})
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "no-tag", Value: "b"})
		if _, err := svc.AddTags(varA.Name, []string{"keep"}, nil); err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}

		variables, err := svc.ListVariables("", []string{"keep"}, nil, nil)
		if err != nil {
			t.Fatalf("ListVariables() error = %v", err)
		}
		if len(variables) != 1 || variables[0].Name != "has-tag" {
			t.Fatalf("variables = %+v, want only [has-tag]", variables)
		}
	})

	t.Run("unknown tag filter returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.ListVariables("", []string{"missing"}, nil, nil)
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("ListVariables() error = %v, want ErrUnknownTagName", err)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.ListVariables("", nil, nil, strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("ListVariables() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("invalid order field returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.ListVariables("bogus", nil, nil, nil)
		if err == nil {
			t.Fatalf("ListVariables() error = nil, want error")
		}
	})
}

// --- SearchVariables ---

func TestVariableServiceSearchVariables(t *testing.T) {
	t.Run("uses default fields", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "apple", Value: "a fruit"})
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "banana", Value: "a yellow fruit"})

		variables, err := svc.SearchVariables("apple", nil, nil, nil)
		if err != nil {
			t.Fatalf("SearchVariables() error = %v", err)
		}
		if len(variables) != 1 || variables[0].Name != "apple" {
			t.Fatalf("variables = %+v, want only [apple]", variables)
		}
	})

	t.Run("respects explicit fields", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "apple", Value: "a fruit"})
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "yellow-apple", Value: "unrelated"})

		variables, err := svc.SearchVariables("apple", []string{"name"}, nil, nil)
		if err != nil {
			t.Fatalf("SearchVariables() error = %v", err)
		}
		wantNames := map[string]bool{"apple": true, "yellow-apple": true}
		if len(variables) != 2 {
			t.Fatalf("len(variables) = %d, want 2; variables = %+v", len(variables), variables)
		}
		for _, v := range variables {
			if !wantNames[v.Name] {
				t.Fatalf("unexpected variable in results: %q", v.Name)
			}
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "apple", Value: "a fruit"})
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "apple-two", Value: "a fruit"})

		limit := 1
		variables, err := svc.SearchVariables("apple", nil, &limit, nil)
		if err != nil {
			t.Fatalf("SearchVariables() error = %v", err)
		}
		if len(variables) != 1 {
			t.Fatalf("len(variables) = %d, want 1", len(variables))
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.SearchVariables("apple", nil, nil, strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("SearchVariables() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("invalid explicit field returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.SearchVariables("apple", []string{"bogus"}, nil, nil)
		if err == nil {
			t.Fatalf("SearchVariables() error = nil, want error")
		}
	})
}

// --- MoveVariable ---

func TestVariableServiceMoveVariable(t *testing.T) {
	t.Run("nil target profile returns ErrNoTargetProfile", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.MoveVariable("greeting", nil, nil)
		if !errors.Is(err, ErrNoTargetProfile) {
			t.Fatalf("MoveVariable() error = %v, want ErrNoTargetProfile", err)
		}
	})

	t.Run("moves variable to target profile", func(t *testing.T) {
		svc, _, _, _, db := setupVariableServiceTest(t)
		target := mustCreateServiceProfile(t, db, "target-profile")
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})

		moved, err := svc.MoveVariable("greeting", strPtr("target-profile"), nil)
		if err != nil {
			t.Fatalf("MoveVariable() error = %v", err)
		}
		if moved.ProfileID != target.ID {
			t.Fatalf("ProfileID = %d, want %d", moved.ProfileID, target.ID)
		}

		got, err := svc.GetVariable("greeting", strPtr("target-profile"))
		if err != nil {
			t.Fatalf("GetVariable() after move error = %v", err)
		}
		if got.ID != moved.ID {
			t.Fatalf("GetVariable() id = %d, want %d", got.ID, moved.ID)
		}
	})

	t.Run("unknown target profile returns error", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})

		_, err := svc.MoveVariable("greeting", strPtr("nope"), nil)
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("MoveVariable() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown source profile returns error", func(t *testing.T) {
		svc, _, _, _, db := setupVariableServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})

		_, err := svc.MoveVariable("greeting", strPtr("target-profile"), strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("MoveVariable() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _, _, db := setupVariableServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")

		_, err := svc.MoveVariable("missing", strPtr("target-profile"), nil)
		if !errors.Is(err, repository.ErrUnKnownName) {
			t.Fatalf("MoveVariable() error = %v, want ErrUnKnownName", err)
		}
	})
}

// --- CopyVariable ---

func TestVariableServiceCopyVariable(t *testing.T) {
	t.Run("nil target profile returns ErrNoTargetProfile", func(t *testing.T) {
		svc, _, _, _, _ := setupVariableServiceTest(t)
		_, err := svc.CopyVariable("greeting", nil, nil, nil)
		if !errors.Is(err, ErrNoTargetProfile) {
			t.Fatalf("CopyVariable() error = %v, want ErrNoTargetProfile", err)
		}
	})

	t.Run("copies variable with same name into target profile", func(t *testing.T) {
		svc, _, _, _, db := setupVariableServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		original := mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})

		copied, err := svc.CopyVariable("greeting", strPtr("target-profile"), nil, nil)
		if err != nil {
			t.Fatalf("CopyVariable() error = %v", err)
		}
		if copied.ID == original.ID {
			t.Fatalf("CopyVariable() returned same variable, want a new one")
		}
		if copied.Name != "greeting" {
			t.Fatalf("Name = %q, want %q", copied.Name, "greeting")
		}
		if copied.Value != "hello" {
			t.Fatalf("Value = %q, want %q", copied.Value, "hello")
		}

		// original is untouched
		stillThere, err := svc.GetVariable("greeting", nil)
		if err != nil {
			t.Fatalf("GetVariable(original) error = %v", err)
		}
		if stillThere.ID != original.ID {
			t.Fatalf("original variable id changed: got %d, want %d", stillThere.ID, original.ID)
		}
	})

	t.Run("copies variable with a new name", func(t *testing.T) {
		svc, _, _, _, db := setupVariableServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})

		copied, err := svc.CopyVariable("greeting", strPtr("target-profile"), strPtr("greeting-copy"), nil)
		if err != nil {
			t.Fatalf("CopyVariable() error = %v", err)
		}
		if copied.Name != "greeting-copy" {
			t.Fatalf("Name = %q, want %q", copied.Name, "greeting-copy")
		}
	})

	t.Run("unknown source profile returns error", func(t *testing.T) {
		svc, _, _, _, db := setupVariableServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")

		_, err := svc.CopyVariable("greeting", strPtr("target-profile"), nil, strPtr("nope"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("CopyVariable() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _, _, db := setupVariableServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")

		_, err := svc.CopyVariable("missing", strPtr("target-profile"), nil, nil)
		if !errors.Is(err, repository.ErrUnKnownName) {
			t.Fatalf("CopyVariable() error = %v, want ErrUnKnownName", err)
		}
	})

	t.Run("name conflict in target profile returns error", func(t *testing.T) {
		svc, _, _, _, db := setupVariableServiceTest(t)
		mustCreateServiceProfile(t, db, "target-profile")
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "hello"})
		mustCreateServiceVariable(t, svc, CreateVariableConfig{Name: "greeting", Value: "other", ProfileName: strPtr("target-profile")})

		_, err := svc.CopyVariable("greeting", strPtr("target-profile"), nil, nil)
		if !errors.Is(err, repository.ErrNameConflict) {
			t.Fatalf("CopyVariable() error = %v, want ErrNameConflict", err)
		}
	})
}
