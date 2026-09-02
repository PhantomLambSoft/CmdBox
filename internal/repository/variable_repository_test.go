package repository

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
)

func setupVariableRepositoryTest(t *testing.T) (VariableRepository, ProfileRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:variable-repo-%d?mode=memory&cache=shared", time.Now().UnixNano())
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
	// declared on the models (e.g. VariableTag.TagID) are actually enforced in tests.
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
	variableRepo := NewVariableRepository(db, profileRepo, validate.NewVariableValidator(nil))
	return variableRepo, profileRepo, db
}

func mustCreateVariable(t *testing.T, repo VariableRepository, input VariableCreateConfig) *models.Variable {
	t.Helper()
	v, err := repo.Create(input)
	if err != nil {
		t.Fatalf("Create(%+v) error = %v", input, err)
	}
	return v
}

// --- Create ---

func TestVariableRepositoryCreate(t *testing.T) {
	repo, _, db := setupVariableRepositoryTest(t)

	t.Run("creates variable using active profile", func(t *testing.T) {
		v, err := repo.Create(VariableCreateConfig{Name: "greeting", Value: "hello"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if v.ID == 0 {
			t.Fatalf("Create() returned zero ID")
		}
		if v.Name != "greeting" {
			t.Fatalf("Name = %q, want %q", v.Name, "greeting")
		}
		if v.Value != "hello" {
			t.Fatalf("Value = %q, want %q", v.Value, "hello")
		}
		if v.ProfileID == 0 {
			t.Fatalf("ProfileID = 0, want active profile id")
		}

		fromDB, err := repo.GetByID(v.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() after Create error = %v", err)
		}
		if fromDB.Name != "greeting" {
			t.Fatalf("persisted name = %q, want %q", fromDB.Name, "greeting")
		}
	})

	t.Run("creates variable with explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "explicit-profile")
		v, err := repo.Create(VariableCreateConfig{Name: "deploy_target", Value: "prod", ProfileID: &other.ID})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if v.ProfileID != other.ID {
			t.Fatalf("ProfileID = %d, want %d", v.ProfileID, other.ID)
		}
	})

	t.Run("trims name whitespace", func(t *testing.T) {
		v, err := repo.Create(VariableCreateConfig{Name: "  spaced  ", Value: "value"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if v.Name != "spaced" {
			t.Fatalf("Name = %q, want %q", v.Name, "spaced")
		}
	})

	t.Run("does not trim value whitespace", func(t *testing.T) {
		v, err := repo.Create(VariableCreateConfig{Name: "raw_value", Value: "  padded  "})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if v.Value != "  padded  " {
			t.Fatalf("Value = %q, want %q (unchanged)", v.Value, "  padded  ")
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := repo.Create(VariableCreateConfig{Name: "", Value: "hello"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects reserved name", func(t *testing.T) {
		_, err := repo.Create(VariableCreateConfig{Name: "help", Value: "hello"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects empty value", func(t *testing.T) {
		_, err := repo.Create(VariableCreateConfig{Name: "empty_value", Value: ""})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects self-referencing value", func(t *testing.T) {
		_, err := repo.Create(VariableCreateConfig{Name: "loop", Value: "prefix <loop> suffix"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects duplicate name in same profile", func(t *testing.T) {
		if _, err := repo.Create(VariableCreateConfig{Name: "dup", Value: "one"}); err != nil {
			t.Fatalf("Create() first error = %v", err)
		}
		_, err := repo.Create(VariableCreateConfig{Name: "dup", Value: "two"})
		if !errors.Is(err, ErrNameConflict) {
			t.Fatalf("Create() second error = %v, want ErrNameConflict", err)
		}
	})

	t.Run("allows same name across different profiles", func(t *testing.T) {
		profileA := createTestProfile(t, db, "var-profile-a")
		profileB := createTestProfile(t, db, "var-profile-b")

		if _, err := repo.Create(VariableCreateConfig{Name: "shared", Value: "a", ProfileID: &profileA.ID}); err != nil {
			t.Fatalf("Create() profile A error = %v", err)
		}
		if _, err := repo.Create(VariableCreateConfig{Name: "shared", Value: "b", ProfileID: &profileB.ID}); err != nil {
			t.Fatalf("Create() profile B error = %v", err)
		}
	})

	t.Run("resolves active variable profile independently of active command profile", func(t *testing.T) {
		commandProfile := createTestProfile(t, db, "active-command-profile")
		variableProfile := createTestProfile(t, db, "active-variable-profile")

		if _, err := profileRepoSetActive(db, &commandProfile.ID, &variableProfile.ID, nil); err != nil {
			t.Fatalf("set active profiles: %v", err)
		}

		v, err := repo.Create(VariableCreateConfig{Name: "profile_scoped", Value: "x"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if v.ProfileID != variableProfile.ID {
			t.Fatalf("ProfileID = %d, want active *variable* profile id %d (not active command profile %d)", v.ProfileID, variableProfile.ID, commandProfile.ID)
		}
	})

	t.Run("propagates error resolving active profile", func(t *testing.T) {
		if err := db.Delete(&models.ProfileState{}, 1).Error; err != nil {
			t.Fatalf("delete profile state: %v", err)
		}
		_, err := repo.Create(VariableCreateConfig{Name: "no-state", Value: "x"})
		if err == nil {
			t.Fatalf("Create() error = nil, want error resolving profile state")
		}
	})
}

// profileRepoSetActive is a small wrapper so tests can set active profiles without
// pulling in every ProfileRepository test helper.
func profileRepoSetActive(db *gorm.DB, commandProfileID, variableProfileID, settingsProfileID *uint) (*models.ProfileState, error) {
	profileRepo := NewProfileRepository(db, validate.NewProfileValidator(nil))
	return profileRepo.SetActiveProfile(commandProfileID, variableProfileID, settingsProfileID)
}

// --- GetByName ---

func TestVariableRepositoryGetByName(t *testing.T) {
	repo, _, db := setupVariableRepositoryTest(t)
	v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "greeting", Value: "hello"})

	t.Run("returns variable scoped to active profile", func(t *testing.T) {
		got, err := repo.GetByName("greeting", nil)
		if err != nil {
			t.Fatalf("GetByName() error = %v", err)
		}
		if got.ID != v.ID {
			t.Fatalf("GetByName() id = %d, want %d", got.ID, v.ID)
		}
	})

	t.Run("lookup is case-sensitive", func(t *testing.T) {
		_, err := repo.GetByName("GREETING", nil)
		if !errors.Is(err, ErrUnKnownName) {
			t.Fatalf("GetByName(\"GREETING\") error = %v, want ErrUnKnownName", err)
		}
	})

	t.Run("returns ErrUnKnownName when missing", func(t *testing.T) {
		_, err := repo.GetByName("missing", nil)
		if !errors.Is(err, ErrUnKnownName) {
			t.Fatalf("GetByName() error = %v, want ErrUnKnownName", err)
		}
	})

	t.Run("scopes lookup to explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "other-profile")
		otherVar, err := repo.Create(VariableCreateConfig{Name: "greeting", Value: "hi", ProfileID: &other.ID})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repo.GetByName("greeting", &other.ID)
		if err != nil {
			t.Fatalf("GetByName(scoped) error = %v", err)
		}
		if got.ID != otherVar.ID {
			t.Fatalf("GetByName(scoped) id = %d, want %d", got.ID, otherVar.ID)
		}

		// still resolves to the active-profile version when no profile id is given.
		got, err = repo.GetByName("greeting", nil)
		if err != nil {
			t.Fatalf("GetByName(active) error = %v", err)
		}
		if got.ID != v.ID {
			t.Fatalf("GetByName(active) id = %d, want %d", got.ID, v.ID)
		}
	})

	t.Run("preserves and retrieves mixed-case name exactly", func(t *testing.T) {
		mixed := mustCreateVariable(t, repo, VariableCreateConfig{Name: "MixedCase", Value: "x"})

		got, err := repo.GetByName("MixedCase", nil)
		if err != nil {
			t.Fatalf("GetByName(\"MixedCase\") error = %v", err)
		}
		if got.ID != mixed.ID {
			t.Fatalf("GetByName(\"MixedCase\") id = %d, want %d", got.ID, mixed.ID)
		}

		_, err = repo.GetByName("mixedcase", nil)
		if !errors.Is(err, ErrUnKnownName) {
			t.Fatalf("GetByName(\"mixedcase\") error = %v, want ErrUnKnownName", err)
		}
	})

	t.Run("names differing only by case coexist in the same profile", func(t *testing.T) {
		capitalized := mustCreateVariable(t, repo, VariableCreateConfig{Name: "Greeting", Value: "x"})

		lower, err := repo.GetByName("greeting", nil)
		if err != nil {
			t.Fatalf("GetByName(\"greeting\") error = %v", err)
		}
		if lower.ID != v.ID {
			t.Fatalf("GetByName(\"greeting\") id = %d, want %d", lower.ID, v.ID)
		}

		upper, err := repo.GetByName("Greeting", nil)
		if err != nil {
			t.Fatalf("GetByName(\"Greeting\") error = %v", err)
		}
		if upper.ID != capitalized.ID {
			t.Fatalf("GetByName(\"Greeting\") id = %d, want %d", upper.ID, capitalized.ID)
		}
	})
}

// --- GetByID ---

func TestVariableRepositoryGetByID(t *testing.T) {
	repo, _, db := setupVariableRepositoryTest(t)
	v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "greeting", Value: "hello"})

	t.Run("returns variable scoped to active profile", func(t *testing.T) {
		got, err := repo.GetByID(v.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got.Name != "greeting" {
			t.Fatalf("Name = %q, want %q", got.Name, "greeting")
		}
	})

	t.Run("returns ErrUnknownVariable when missing", func(t *testing.T) {
		_, err := repo.GetByID(v.ID+999999, nil)
		if !errors.Is(err, ErrUnknownVariable) {
			t.Fatalf("GetByID() error = %v, want ErrUnknownVariable", err)
		}
	})

	t.Run("returns ErrUnknownVariable when scoped to wrong profile", func(t *testing.T) {
		other := createTestProfile(t, db, "wrong-profile")
		_, err := repo.GetByID(v.ID, &other.ID)
		if !errors.Is(err, ErrUnknownVariable) {
			t.Fatalf("GetByID(wrong profile) error = %v, want ErrUnknownVariable", err)
		}
	})

	t.Run("resolves active variable profile independently of active command profile", func(t *testing.T) {
		commandProfile := createTestProfile(t, db, "get-active-command-profile")
		variableProfile := createTestProfile(t, db, "get-active-variable-profile")
		scoped, err := repo.Create(VariableCreateConfig{Name: "scoped-var", Value: "x", ProfileID: &variableProfile.ID})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if _, err := profileRepoSetActive(db, &commandProfile.ID, &variableProfile.ID, nil); err != nil {
			t.Fatalf("set active profiles: %v", err)
		}

		got, err := repo.GetByID(scoped.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() error = %v, want lookup against active variable profile to succeed", err)
		}
		if got.ID != scoped.ID {
			t.Fatalf("GetByID() id = %d, want %d", got.ID, scoped.ID)
		}
	})
}

// --- Update ---

func TestVariableRepositoryUpdateValidation(t *testing.T) {
	repo, _, _ := setupVariableRepositoryTest(t)
	v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "greeting", Value: "hello"})

	t.Run("nil variable returns ErrNoUpdateTarget", func(t *testing.T) {
		_, err := repo.Update(nil, VariableUpdateConfig{})
		if !errors.Is(err, ErrNoUpdateTarget) {
			t.Fatalf("Update(nil) error = %v, want ErrNoUpdateTarget", err)
		}
	})

	t.Run("rejects reserved name on rename", func(t *testing.T) {
		fresh := *v
		newName := "help"
		_, err := repo.Update(&fresh, VariableUpdateConfig{Name: &newName})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Update() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects empty value on update", func(t *testing.T) {
		fresh := *v
		newValue := "   "
		_, err := repo.Update(&fresh, VariableUpdateConfig{Value: &newValue})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Update() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects self-referencing value after merge", func(t *testing.T) {
		fresh := *v
		newValue := "run <greeting> now"
		_, err := repo.Update(&fresh, VariableUpdateConfig{Value: &newValue})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Update() error = %v, want ErrValidation", err)
		}
	})
}

func TestVariableRepositoryUpdateFields(t *testing.T) {
	repo, _, db := setupVariableRepositoryTest(t)

	t.Run("updates name and value", func(t *testing.T) {
		v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "rename-me", Value: "old"})
		newName := "renamed"
		newValue := "new"
		updated, err := repo.Update(v, VariableUpdateConfig{Name: &newName, Value: &newValue})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Name != "renamed" || updated.Value != "new" {
			t.Fatalf("Update() = %+v, want name=renamed value=new", updated)
		}

		persisted, err := repo.GetByID(v.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if persisted.Name != "renamed" {
			t.Fatalf("persisted name = %q, want %q", persisted.Name, "renamed")
		}
	})

	t.Run("trims renamed name whitespace", func(t *testing.T) {
		v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "trim-me", Value: "hi"})
		newName := "  trimmed  "
		updated, err := repo.Update(v, VariableUpdateConfig{Name: &newName})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Name != "trimmed" {
			t.Fatalf("Name = %q, want %q", updated.Name, "trimmed")
		}
	})

	t.Run("trims updated value whitespace", func(t *testing.T) {
		v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "trim-value", Value: "hi"})
		newValue := "  trimmed value  "
		updated, err := repo.Update(v, VariableUpdateConfig{Value: &newValue})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Value != "trimmed value" {
			t.Fatalf("Value = %q, want %q", updated.Value, "trimmed value")
		}
	})

	t.Run("no-op update leaves fields unchanged", func(t *testing.T) {
		v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "noop-me", Value: "kept"})
		updated, err := repo.Update(v, VariableUpdateConfig{})
		if err != nil {
			t.Fatalf("Update(no-op) error = %v", err)
		}
		if updated.Name != "noop-me" || updated.Value != "kept" {
			t.Fatalf("Update(no-op) changed fields: %+v", updated)
		}
	})

	t.Run("rejects rename to name already used in same profile", func(t *testing.T) {
		_ = mustCreateVariable(t, repo, VariableCreateConfig{Name: "taken", Value: "x"})
		v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "renaming", Value: "x"})

		newName := "taken"
		_, err := repo.Update(v, VariableUpdateConfig{Name: &newName})
		if !errors.Is(err, ErrNameConflict) {
			t.Fatalf("Update() error = %v, want ErrNameConflict", err)
		}
	})

	t.Run("allows rename to name used in a different profile", func(t *testing.T) {
		other := createTestProfile(t, db, "rename-profile")
		if _, err := repo.Create(VariableCreateConfig{Name: "cross-profile", Value: "x", ProfileID: &other.ID}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "renaming2", Value: "x"})

		newName := "cross-profile"
		updated, err := repo.Update(v, VariableUpdateConfig{Name: &newName})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Name != "cross-profile" {
			t.Fatalf("Name = %q, want %q", updated.Name, "cross-profile")
		}
	})

	t.Run("moves variable to a new profile", func(t *testing.T) {
		origin := createTestProfile(t, db, "move-origin")
		target := createTestProfile(t, db, "move-target")
		v, err := repo.Create(VariableCreateConfig{Name: "movable", Value: "x", ProfileID: &origin.ID})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		updated, err := repo.Update(v, VariableUpdateConfig{ProfileID: &target.ID})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.ProfileID != target.ID {
			t.Fatalf("ProfileID = %d, want %d", updated.ProfileID, target.ID)
		}

		if _, err := repo.GetByID(v.ID, &origin.ID); !errors.Is(err, ErrUnknownVariable) {
			t.Fatalf("GetByID(origin) error = %v, want ErrUnknownVariable", err)
		}
		persisted, err := repo.GetByID(v.ID, &target.ID)
		if err != nil {
			t.Fatalf("GetByID(target) error = %v", err)
		}
		if persisted.Name != "movable" {
			t.Fatalf("persisted name = %q, want %q", persisted.Name, "movable")
		}
	})

	t.Run("nil ProfileID leaves profile unchanged", func(t *testing.T) {
		origin := createTestProfile(t, db, "stay-origin")
		v, err := repo.Create(VariableCreateConfig{Name: "stationary", Value: "x", ProfileID: &origin.ID})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		newValue := "y"
		updated, err := repo.Update(v, VariableUpdateConfig{Value: &newValue})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.ProfileID != origin.ID {
			t.Fatalf("ProfileID = %d, want unchanged %d", updated.ProfileID, origin.ID)
		}
	})

	t.Run("rejects move that conflicts with existing name in target profile", func(t *testing.T) {
		target := createTestProfile(t, db, "conflict-target")
		if _, err := repo.Create(VariableCreateConfig{Name: "taken-in-target", Value: "x", ProfileID: &target.ID}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "taken-in-target", Value: "y"})

		_, err := repo.Update(v, VariableUpdateConfig{ProfileID: &target.ID})
		if !errors.Is(err, ErrNameConflict) {
			t.Fatalf("Update() error = %v, want ErrNameConflict", err)
		}
	})
}

// --- Delete ---

func TestVariableRepositoryDelete(t *testing.T) {
	repo, _, _ := setupVariableRepositoryTest(t)

	t.Run("nil variable is a no-op", func(t *testing.T) {
		if err := repo.Delete(nil); err != nil {
			t.Fatalf("Delete(nil) error = %v", err)
		}
	})

	t.Run("deletes existing variable", func(t *testing.T) {
		v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "delete-me", Value: "x"})
		if err := repo.Delete(v); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		_, err := repo.GetByID(v.ID, nil)
		if !errors.Is(err, ErrUnknownVariable) {
			t.Fatalf("GetByID() after delete error = %v, want ErrUnknownVariable", err)
		}
	})
}

// --- AddTags / RemoveTags ---

func TestVariableRepositoryAddTags(t *testing.T) {
	repo, _, db := setupVariableRepositoryTest(t)
	v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "tag-me", Value: "x"})
	tagA := createTestTag(t, db, "var-tag-a")
	tagB := createTestTag(t, db, "var-tag-b")

	t.Run("empty tags is a no-op", func(t *testing.T) {
		result, err := repo.AddTags(v, nil)
		if err != nil {
			t.Fatalf("AddTags(nil) error = %v", err)
		}
		if len(result.Added) != 0 || len(result.Existing) != 0 {
			t.Fatalf("AddTags(nil) = %+v, want zero value", result)
		}
	})

	t.Run("adds new tags", func(t *testing.T) {
		result, err := repo.AddTags(v, []models.Tag{tagA, tagB})
		if err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}
		if len(result.Added) != 2 || len(result.Existing) != 0 {
			t.Fatalf("AddTags() = %+v, want both added", result)
		}

		var count int64
		if err := db.Model(&models.VariableTag{}).Where("variable_id = ?", v.ID).Count(&count).Error; err != nil {
			t.Fatalf("count variable_tags: %v", err)
		}
		if count != 2 {
			t.Fatalf("variable_tags count = %d, want 2", count)
		}
	})

	t.Run("re-adding tags reports them as existing", func(t *testing.T) {
		result, err := repo.AddTags(v, []models.Tag{tagA, tagB})
		if err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}
		if len(result.Added) != 0 || len(result.Existing) != 2 {
			t.Fatalf("AddTags() = %+v, want both existing", result)
		}
	})

	t.Run("mix of new and existing tags", func(t *testing.T) {
		tagC := createTestTag(t, db, "var-tag-c")
		result, err := repo.AddTags(v, []models.Tag{tagA, tagC})
		if err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}
		if len(result.Added) != 1 || result.Added[0] != "var-tag-c" {
			t.Fatalf("Added = %v, want [var-tag-c]", result.Added)
		}
		if len(result.Existing) != 1 || result.Existing[0] != "var-tag-a" {
			t.Fatalf("Existing = %v, want [var-tag-a]", result.Existing)
		}
	})
}

func TestVariableRepositoryAddTagsPropagatesTransactionError(t *testing.T) {
	repo, _, _ := setupVariableRepositoryTest(t)
	v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "tag-fail", Value: "x"})

	// Reference a tag id that was never persisted; the foreign key on
	// VariableTag.TagID makes the insert fail, and the error should be wrapped.
	bogusTag := models.Tag{ID: 999999, Name: "bogus"}
	_, err := repo.AddTags(v, []models.Tag{bogusTag})
	if !errors.Is(err, ErrTagAttachFailed) {
		t.Fatalf("AddTags() error = %v, want ErrTagAttachFailed", err)
	}
}

func TestVariableRepositoryRemoveTags(t *testing.T) {
	repo, _, db := setupVariableRepositoryTest(t)
	v := mustCreateVariable(t, repo, VariableCreateConfig{Name: "untag-me", Value: "x"})
	tagA := createTestTag(t, db, "var-remove-a")
	tagB := createTestTag(t, db, "var-remove-b")

	t.Run("empty tags is a no-op", func(t *testing.T) {
		result, err := repo.RemoveTags(v, nil)
		if err != nil {
			t.Fatalf("RemoveTags(nil) error = %v", err)
		}
		if len(result.Removed) != 0 || len(result.NotAttached) != 0 {
			t.Fatalf("RemoveTags(nil) = %+v, want zero value", result)
		}
	})

	t.Run("removing unattached tags", func(t *testing.T) {
		result, err := repo.RemoveTags(v, []models.Tag{tagA, tagB})
		if err != nil {
			t.Fatalf("RemoveTags() error = %v", err)
		}
		if len(result.Removed) != 0 || len(result.NotAttached) != 2 {
			t.Fatalf("RemoveTags() = %+v, want both not attached", result)
		}
	})

	t.Run("removes attached tags and reports mix", func(t *testing.T) {
		if _, err := repo.AddTags(v, []models.Tag{tagA}); err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}

		result, err := repo.RemoveTags(v, []models.Tag{tagA, tagB})
		if err != nil {
			t.Fatalf("RemoveTags() error = %v", err)
		}
		if len(result.Removed) != 1 || result.Removed[0] != "var-remove-a" {
			t.Fatalf("Removed = %v, want [var-remove-a]", result.Removed)
		}
		if len(result.NotAttached) != 1 || result.NotAttached[0] != "var-remove-b" {
			t.Fatalf("NotAttached = %v, want [var-remove-b]", result.NotAttached)
		}

		var count int64
		if err := db.Model(&models.VariableTag{}).Where("variable_id = ? AND tag_id = ?", v.ID, tagA.ID).Count(&count).Error; err != nil {
			t.Fatalf("count variable_tags: %v", err)
		}
		if count != 0 {
			t.Fatalf("variable_tags count for removed tag = %d, want 0", count)
		}
	})
}

// --- ListAll ---

func TestVariableRepositoryListAll(t *testing.T) {
	repo, _, db := setupVariableRepositoryTest(t)
	_ = mustCreateVariable(t, repo, VariableCreateConfig{Name: "zeta", Value: "z"})
	_ = mustCreateVariable(t, repo, VariableCreateConfig{Name: "alpha", Value: "a"})
	_ = mustCreateVariable(t, repo, VariableCreateConfig{Name: "mid", Value: "m"})

	t.Run("defaults to name order and limit 25", func(t *testing.T) {
		variables, err := repo.ListAll("", 0, nil)
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(variables) != 3 {
			t.Fatalf("len(variables) = %d, want 3", len(variables))
		}
		wantOrder := []string{"alpha", "mid", "zeta"}
		for i, want := range wantOrder {
			if variables[i].Name != want {
				t.Fatalf("variables[%d].Name = %q, want %q", i, variables[i].Name, want)
			}
		}
	})

	t.Run("respects explicit order and limit", func(t *testing.T) {
		variables, err := repo.ListAll("-name", 2, nil)
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(variables) != 2 {
			t.Fatalf("len(variables) = %d, want 2", len(variables))
		}
		if variables[0].Name != "zeta" || variables[1].Name != "mid" {
			t.Fatalf("variables = %+v, want [zeta, mid]", variables)
		}
	})

	t.Run("scopes to explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "list-all-other")
		if _, err := repo.Create(VariableCreateConfig{Name: "in-other", Value: "x", ProfileID: &other.ID}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		variables, err := repo.ListAll("", 0, &other.ID)
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(variables) != 1 || variables[0].Name != "in-other" {
			t.Fatalf("variables = %+v, want only [in-other]", variables)
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

func TestVariableRepositoryListByTags(t *testing.T) {
	repo, _, db := setupVariableRepositoryTest(t)
	tagA := createTestTag(t, db, "list-var-tag-a")
	tagB := createTestTag(t, db, "list-var-tag-b")

	varA := mustCreateVariable(t, repo, VariableCreateConfig{Name: "has-a", Value: "a"})
	varB := mustCreateVariable(t, repo, VariableCreateConfig{Name: "has-b", Value: "b"})
	varBoth := mustCreateVariable(t, repo, VariableCreateConfig{Name: "has-both", Value: "both"})
	_ = mustCreateVariable(t, repo, VariableCreateConfig{Name: "has-neither", Value: "none"})

	if _, err := repo.AddTags(varA, []models.Tag{tagA}); err != nil {
		t.Fatalf("AddTags() error = %v", err)
	}
	if _, err := repo.AddTags(varB, []models.Tag{tagB}); err != nil {
		t.Fatalf("AddTags() error = %v", err)
	}
	if _, err := repo.AddTags(varBoth, []models.Tag{tagA, tagB}); err != nil {
		t.Fatalf("AddTags() error = %v", err)
	}

	t.Run("empty tags returns nil without error", func(t *testing.T) {
		variables, err := repo.ListByTags(nil, "", 0, nil)
		if err != nil {
			t.Fatalf("ListByTags(nil) error = %v", err)
		}
		if variables != nil {
			t.Fatalf("ListByTags(nil) = %v, want nil", variables)
		}
	})

	t.Run("returns variables matching any tag without duplicates", func(t *testing.T) {
		variables, err := repo.ListByTags([]models.Tag{tagA, tagB}, "name", 0, nil)
		if err != nil {
			t.Fatalf("ListByTags() error = %v", err)
		}
		if len(variables) != 3 {
			t.Fatalf("len(variables) = %d, want 3 (no duplicates); variables = %+v", len(variables), variables)
		}
		wantNames := map[string]bool{"has-a": true, "has-b": true, "has-both": true}
		for _, v := range variables {
			if !wantNames[v.Name] {
				t.Fatalf("unexpected variable in results: %q", v.Name)
			}
		}
	})

	t.Run("scopes to explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "list-var-tags-other")
		otherVar, err := repo.Create(VariableCreateConfig{Name: "other-tagged", Value: "x", ProfileID: &other.ID})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if _, err := repo.AddTags(otherVar, []models.Tag{tagA}); err != nil {
			t.Fatalf("AddTags() error = %v", err)
		}

		variables, err := repo.ListByTags([]models.Tag{tagA}, "", 0, &other.ID)
		if err != nil {
			t.Fatalf("ListByTags() error = %v", err)
		}
		if len(variables) != 1 || variables[0].Name != "other-tagged" {
			t.Fatalf("variables = %+v, want only [other-tagged]", variables)
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

func TestVariableRepositorySearch(t *testing.T) {
	repo, _, db := setupVariableRepositoryTest(t)
	_ = mustCreateVariable(t, repo, VariableCreateConfig{Name: "apple", Value: "a fruit"})
	_ = mustCreateVariable(t, repo, VariableCreateConfig{Name: "green-apple", Value: "a green fruit"})
	_ = mustCreateVariable(t, repo, VariableCreateConfig{Name: "banana", Value: "a yellow fruit"})

	t.Run("uses default fields when none provided", func(t *testing.T) {
		variables, err := repo.Search("apple", nil, 0, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(variables) != 2 {
			t.Fatalf("len(variables) = %d, want 2; variables = %+v", len(variables), variables)
		}
		if variables[0].Name != "apple" {
			t.Fatalf("variables[0].Name = %q, want %q (closest match first)", variables[0].Name, "apple")
		}
	})

	t.Run("matches against value field by default", func(t *testing.T) {
		variables, err := repo.Search("yellow", nil, 0, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(variables) != 1 || variables[0].Name != "banana" {
			t.Fatalf("variables = %+v, want only [banana]", variables)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		variables, err := repo.Search("apple", nil, 1, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(variables) != 1 {
			t.Fatalf("len(variables) = %d, want 1", len(variables))
		}
	})

	t.Run("scopes to explicit profile id", func(t *testing.T) {
		other := createTestProfile(t, db, "search-var-other")
		if _, err := repo.Create(VariableCreateConfig{Name: "apple-elsewhere", Value: "x", ProfileID: &other.ID}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		variables, err := repo.Search("apple", nil, 0, &other.ID)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(variables) != 1 || variables[0].Name != "apple-elsewhere" {
			t.Fatalf("variables = %+v, want only [apple-elsewhere]", variables)
		}
	})

	t.Run("invalid explicit field returns error", func(t *testing.T) {
		_, err := repo.Search("apple", []string{"bogus"}, 0, nil)
		if err == nil {
			t.Fatalf("Search() error = nil, want error")
		}
	})

	t.Run("no matches returns empty slice", func(t *testing.T) {
		variables, err := repo.Search("nonexistent", nil, 0, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(variables) != 0 {
			t.Fatalf("variables = %+v, want empty", variables)
		}
	})
}

// --- WithTx ---

func TestVariableRepositoryWithTx(t *testing.T) {
	t.Run("operations persist after commit", func(t *testing.T) {
		repo, _, db := setupVariableRepositoryTest(t)
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		txRepo := repo.WithTx(tx)

		if _, err := txRepo.Create(VariableCreateConfig{Name: "tx-commit", Value: "hi"}); err != nil {
			t.Fatalf("Create() within tx error = %v", err)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit tx: %v", err)
		}

		if _, err := repo.GetByName("tx-commit", nil); err != nil {
			t.Fatalf("GetByName() after commit error = %v", err)
		}
	})

	t.Run("operations discarded on rollback", func(t *testing.T) {
		repo, _, db := setupVariableRepositoryTest(t)
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		txRepo := repo.WithTx(tx)

		if _, err := txRepo.Create(VariableCreateConfig{Name: "tx-rollback", Value: "hi"}); err != nil {
			t.Fatalf("Create() within tx error = %v", err)
		}
		if err := tx.Rollback().Error; err != nil {
			t.Fatalf("rollback tx: %v", err)
		}

		if _, err := repo.GetByName("tx-rollback", nil); !errors.Is(err, ErrUnKnownName) {
			t.Fatalf("GetByName() after rollback error = %v, want ErrUnKnownName", err)
		}
	})

	t.Run("preserves validator", func(t *testing.T) {
		repo, _, db := setupVariableRepositoryTest(t)
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		txRepo := repo.WithTx(tx)

		_, err := txRepo.Create(VariableCreateConfig{Name: "help", Value: "hi"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() within tx error = %v, want ErrValidation", err)
		}
		if err := tx.Rollback().Error; err != nil {
			t.Fatalf("rollback tx: %v", err)
		}
	})

	t.Run("preserves profile resolution", func(t *testing.T) {
		repo, _, db := setupVariableRepositoryTest(t)
		other := createTestProfile(t, db, "tx-var-profile")

		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		txRepo := repo.WithTx(tx)

		v, err := txRepo.Create(VariableCreateConfig{Name: "tx-scoped", Value: "hi", ProfileID: &other.ID})
		if err != nil {
			t.Fatalf("Create() within tx error = %v", err)
		}
		if v.ProfileID != other.ID {
			t.Fatalf("ProfileID = %d, want %d", v.ProfileID, other.ID)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit tx: %v", err)
		}
	})

	t.Run("reads within the transaction see uncommitted writes", func(t *testing.T) {
		repo, _, db := setupVariableRepositoryTest(t)
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		txRepo := repo.WithTx(tx)

		created, err := txRepo.Create(VariableCreateConfig{Name: "tx-visible", Value: "hi"})
		if err != nil {
			t.Fatalf("Create() within tx error = %v", err)
		}

		got, err := txRepo.GetByID(created.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() within tx error = %v", err)
		}
		if got.Name != "tx-visible" {
			t.Fatalf("GetByID() within tx = %+v, want name=tx-visible", got)
		}

		if err := tx.Rollback().Error; err != nil {
			t.Fatalf("rollback tx: %v", err)
		}
	})
}

// --- misc ---

func TestNewVariableRepositoryDefaultValidator(t *testing.T) {
	_, profileRepo, db := setupVariableRepositoryTest(t)
	repo := NewVariableRepository(db, profileRepo, nil)

	_, err := repo.Create(VariableCreateConfig{Name: "help", Value: "hi"})
	if !errors.Is(err, validate.ErrValidation) {
		t.Fatalf("Create() with default validator error = %v, want ErrValidation (reserved name)", err)
	}
}
