package repository

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
)

func setupProfileRepositoryTest(t *testing.T) (ProfileRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:profile-repo-%d?mode=memory&cache=shared", time.Now().UnixNano())
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

	if err := db.AutoMigrate(&models.Profile{}, &models.ProfileState{}, &models.Command{}, &models.Variable{}); err != nil {
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

	return NewProfileRepository(db, validate.NewProfileValidator(nil)), db
}

func strPtr(v string) *string {
	return &v
}

func createProfile(t *testing.T, db *gorm.DB, name string, description *string) models.Profile {
	t.Helper()

	profile := models.Profile{Name: name, Description: description}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("create profile %q: %v", name, err)
	}
	return profile
}

func createCommand(t *testing.T, db *gorm.DB, profileID uint, alias string) {
	t.Helper()

	command := models.Command{
		Alias:     alias,
		Template:  "echo test",
		ProfileID: profileID,
	}
	if err := db.Create(&command).Error; err != nil {
		t.Fatalf("create command %q: %v", alias, err)
	}
}

func createVariable(t *testing.T, db *gorm.DB, profileID uint, name string) {
	t.Helper()

	variable := models.Variable{
		Name:      name,
		Value:     "value",
		ProfileID: profileID,
	}
	if err := db.Create(&variable).Error; err != nil {
		t.Fatalf("create variable %q: %v", name, err)
	}
}

func TestProfileRepositoryCreate(t *testing.T) {
	repo, _ := setupProfileRepositoryTest(t)

	t.Run("creates profile", func(t *testing.T) {
		description := "workspace profile"
		created, err := repo.Create("work", &description)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if created.ID == 0 {
			t.Fatalf("Create() returned zero ID")
		}
		if created.Name != "work" {
			t.Fatalf("Create() name = %q, want %q", created.Name, "work")
		}
		if created.Description == nil || *created.Description != description {
			t.Fatalf("Create() description mismatch")
		}

		fromDB, err := repo.GetByName("work")
		if err != nil {
			t.Fatalf("GetByName() after Create error = %v", err)
		}
		if fromDB.ID != created.ID {
			t.Fatalf("persisted profile id = %d, want %d", fromDB.ID, created.ID)
		}
	})

	t.Run("rejects duplicate profile name", func(t *testing.T) {
		if _, err := repo.Create(DefaultProfileName, nil); !errors.Is(err, ErrProfileNameExists) {
			t.Fatalf("Create() error = %v, want ErrProfileNameExists", err)
		}
	})

}

func TestProfileRepositoryGetByNameAndID(t *testing.T) {
	repo, db := setupProfileRepositoryTest(t)
	profile := createProfile(t, db, "dev", strPtr("development"))

	t.Run("GetByName returns profile", func(t *testing.T) {
		got, err := repo.GetByName("dev")
		if err != nil {
			t.Fatalf("GetByName() error = %v", err)
		}
		if got.ID != profile.ID {
			t.Fatalf("GetByName() id = %d, want %d", got.ID, profile.ID)
		}
	})

	t.Run("GetByName maps not found", func(t *testing.T) {
		_, err := repo.GetByName("missing")
		if !errors.Is(err, ErrProfileNotFound) {
			t.Fatalf("GetByName() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("GetByID returns profile", func(t *testing.T) {
		got, err := repo.GetByID(profile.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got.Name != "dev" {
			t.Fatalf("GetByID() name = %q, want %q", got.Name, "dev")
		}
	})

	t.Run("GetByID maps not found", func(t *testing.T) {
		_, err := repo.GetByID(profile.ID + 999)
		if !errors.Is(err, ErrProfileNotFound) {
			t.Fatalf("GetByID() error = %v, want ErrProfileNotFound", err)
		}
	})
}

func TestProfileRepositoryList(t *testing.T) {
	repo, db := setupProfileRepositoryTest(t)

	profiles, err := repo.ListAll()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(profiles) != 1 || profiles[0].Name != DefaultProfileName {
		t.Fatalf("List() default result = %#v, want only default", profiles)
	}

	createProfile(t, db, "zeta", nil)
	createProfile(t, db, "alpha", nil)

	profiles, err = repo.ListAll()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	gotNames := []string{profiles[0].Name, profiles[1].Name, profiles[2].Name}
	wantNames := []string{"alpha", DefaultProfileName, "zeta"}
	if strings.Join(gotNames, ",") != strings.Join(wantNames, ",") {
		t.Fatalf("List() order = %v, want %v", gotNames, wantNames)
	}
}

func TestProfileRepositoryUpdate(t *testing.T) {
	repo, db := setupProfileRepositoryTest(t)
	target := createProfile(t, db, "dev", strPtr("old description"))
	_ = createProfile(t, db, "qa", nil)

	t.Run("returns not found", func(t *testing.T) {
		newName := "new-name"
		_, err := repo.Update(target.ID+12345, &newName, nil)
		if !errors.Is(err, ErrProfileNotFound) {
			t.Fatalf("Update() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("prevents renaming default profile", func(t *testing.T) {
		newName := "renamed-default"
		_, err := repo.Update(1, &newName, nil)
		if !errors.Is(err, ErrDefaultProfileName) {
			t.Fatalf("Update() error = %v, want ErrDefaultProfileName", err)
		}
	})

	t.Run("prevents duplicate name", func(t *testing.T) {
		newName := "qa"
		_, err := repo.Update(target.ID, &newName, nil)
		if !errors.Is(err, ErrProfileNameExists) {
			t.Fatalf("Update() error = %v, want ErrProfileNameExists", err)
		}
	})

	t.Run("updates name and description", func(t *testing.T) {
		newName := "dev-renamed"
		newDescription := "new description"
		updated, err := repo.Update(target.ID, &newName, &newDescription)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Name != newName {
			t.Fatalf("Update() name = %q, want %q", updated.Name, newName)
		}
		if updated.Description == nil || *updated.Description != newDescription {
			t.Fatalf("Update() description mismatch")
		}

		persisted, err := repo.GetByID(target.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if persisted.Name != newName {
			t.Fatalf("persisted name = %q, want %q", persisted.Name, newName)
		}
	})

	t.Run("leaves description unchanged when description argument is nil", func(t *testing.T) {
		updated, err := repo.Update(target.ID, nil, nil)
		if err != nil {
			t.Fatalf("Update() with nil description pointer error = %v", err)
		}
		if updated.Description == nil || *updated.Description != "new description" {
			t.Fatalf("Update() description = %v, want %q", updated.Description, "new description")
		}
	})

	t.Run("default profile unchanged when name is same", func(t *testing.T) {
		sameName := DefaultProfileName
		description := "updated default description"
		updated, err := repo.Update(1, &sameName, &description)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Name != DefaultProfileName {
			t.Fatalf("default profile name changed: %q", updated.Name)
		}
		if updated.Description == nil || *updated.Description != description {
			t.Fatalf("default profile description not updated")
		}
	})
}

func TestProfileRepositoryDelete(t *testing.T) {
	repo, db := setupProfileRepositoryTest(t)

	t.Run("returns not found", func(t *testing.T) {
		err := repo.Delete(999999, false)
		if !errors.Is(err, ErrProfileNotFound) {
			t.Fatalf("Delete() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("prevents deleting default profile", func(t *testing.T) {
		err := repo.Delete(1, false)
		if !errors.Is(err, ErrDefaultProfileDelete) {
			t.Fatalf("Delete() error = %v, want ErrDefaultProfileDelete", err)
		}
	})

	t.Run("prevents deleting active profile for each active slot", func(t *testing.T) {
		cases := []struct {
			name   string
			apply  func(id uint)
			target string
		}{
			{
				name: "active command profile",
				apply: func(id uint) {
					_ = db.Model(&models.ProfileState{}).Where("id = ?", 1).Update("active_command_profile_id", id)
				},
				target: "cmd",
			},
			{
				name: "active variable profile",
				apply: func(id uint) {
					_ = db.Model(&models.ProfileState{}).Where("id = ?", 1).Update("active_variable_profile_id", id)
				},
				target: "var",
			},
			{
				name: "active settings profile",
				apply: func(id uint) {
					_ = db.Model(&models.ProfileState{}).Where("id = ?", 1).Update("active_settings_profile_id", id)
				},
				target: "settings",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				profile := createProfile(t, db, "active-"+tc.target, nil)
				tc.apply(profile.ID)

				err := repo.Delete(profile.ID, false)
				if !errors.Is(err, ErrProfileInUse) {
					t.Fatalf("Delete() error = %v, want ErrProfileInUse", err)
				}

				_ = db.Model(&models.ProfileState{}).Where("id = ?", 1).Updates(map[string]any{
					"active_command_profile_id":  1,
					"active_variable_profile_id": 1,
					"active_settings_profile_id": 1,
				})
			})
		}
	})

	t.Run("blocks deletion when profile has commands or variables and force is false", func(t *testing.T) {
		profile := createProfile(t, db, "with-content", nil)
		createCommand(t, db, profile.ID, "test-command")
		createVariable(t, db, profile.ID, "test-variable")

		err := repo.Delete(profile.ID, false)
		if err == nil {
			t.Fatalf("Delete() error = nil, want ErrProfileHasContent")
		}
		if !errors.Is(err, ErrProfileHasContent) {
			t.Fatalf("Delete() error = %v, want ErrProfileHasContent", err)
		}
	})

	t.Run("force delete ignores content", func(t *testing.T) {
		profile := createProfile(t, db, "force-delete", nil)
		createCommand(t, db, profile.ID, "force-command")
		createVariable(t, db, profile.ID, "force-variable")

		if err := repo.Delete(profile.ID, true); err != nil {
			t.Fatalf("Delete(force=true) error = %v", err)
		}

		_, err := repo.GetByID(profile.ID)
		if !errors.Is(err, ErrProfileNotFound) {
			t.Fatalf("GetByID() after delete error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("deletes profile without content", func(t *testing.T) {
		profile := createProfile(t, db, "plain-delete", nil)

		if err := repo.Delete(profile.ID, false); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err := repo.GetByID(profile.ID)
		if !errors.Is(err, ErrProfileNotFound) {
			t.Fatalf("GetByID() after delete error = %v, want ErrProfileNotFound", err)
		}
	})
}

func TestProfileRepositoryTouchLastUsed(t *testing.T) {
	repo, db := setupProfileRepositoryTest(t)
	profile := createProfile(t, db, "touch-me", nil)

	if err := repo.RecordUse(profile.ID); err != nil {
		t.Fatalf("TouchLastUsed() error = %v", err)
	}

	updated, err := repo.GetByID(profile.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if updated.LastUsed == nil {
		t.Fatalf("TouchLastUsed() did not set LastUsed")
	}
	if updated.LastUsed.Before(updated.CreatedAt) {
		t.Fatalf("TouchLastUsed() LastUsed before CreatedAt")
	}

	if err := repo.RecordUse(profile.ID + 111111); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("TouchLastUsed() missing error = %v, want ErrProfileNotFound", err)
	}
}

func TestProfileRepositoryGetState(t *testing.T) {
	repo, db := setupProfileRepositoryTest(t)

	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if state.ID != 1 {
		t.Fatalf("GetState() id = %d, want 1", state.ID)
	}
	if !state.Linked() {
		t.Fatalf("GetState() expected linked state")
	}

	if err := db.Delete(&models.ProfileState{}, 1).Error; err != nil {
		t.Fatalf("delete state row error = %v", err)
	}

	_, err = repo.GetState()
	if err == nil || !strings.Contains(err.Error(), "loading profile state") {
		t.Fatalf("GetState() missing state error = %v, want wrapped loading error", err)
	}
}

func TestProfileRepositorySetActiveProfile(t *testing.T) {
	repo, db := setupProfileRepositoryTest(t)
	commandProfile := createProfile(t, db, "active-command", nil)
	variableProfile := createProfile(t, db, "active-variable", nil)
	settingsProfile := createProfile(t, db, "active-settings", nil)

	t.Run("no-op when all args nil", func(t *testing.T) {
		stateBefore, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState() error = %v", err)
		}

		stateAfter, err := repo.SetActiveProfile(nil, nil, nil)
		if err != nil {
			t.Fatalf("SetActiveProfile(nil,nil,nil) error = %v", err)
		}

		if stateAfter.ID != stateBefore.ID ||
			stateAfter.ActiveCommandProfileID != stateBefore.ActiveCommandProfileID ||
			stateAfter.ActiveVariableProfileID != stateBefore.ActiveVariableProfileID ||
			stateAfter.ActiveSettingsProfileID != stateBefore.ActiveSettingsProfileID {
			t.Fatalf("SetActiveProfile(nil,nil,nil) changed active ids: before=%+v after=%+v", *stateBefore, *stateAfter)
		}
	})

	t.Run("updates each field independently", func(t *testing.T) {
		state, err := repo.SetActiveProfile(&commandProfile.ID, nil, nil)
		if err != nil {
			t.Fatalf("SetActiveProfile(command) error = %v", err)
		}
		if state.ActiveCommandProfileID != commandProfile.ID {
			t.Fatalf("active command id = %d, want %d", state.ActiveCommandProfileID, commandProfile.ID)
		}

		state, err = repo.SetActiveProfile(nil, &variableProfile.ID, nil)
		if err != nil {
			t.Fatalf("SetActiveProfile(variable) error = %v", err)
		}
		if state.ActiveVariableProfileID != variableProfile.ID {
			t.Fatalf("active variable id = %d, want %d", state.ActiveVariableProfileID, variableProfile.ID)
		}

		state, err = repo.SetActiveProfile(nil, nil, &settingsProfile.ID)
		if err != nil {
			t.Fatalf("SetActiveProfile(settings) error = %v", err)
		}
		if state.ActiveSettingsProfileID != settingsProfile.ID {
			t.Fatalf("active settings id = %d, want %d", state.ActiveSettingsProfileID, settingsProfile.ID)
		}
	})

	t.Run("updates multiple fields in one call", func(t *testing.T) {
		state, err := repo.SetActiveProfile(&commandProfile.ID, &variableProfile.ID, &settingsProfile.ID)
		if err != nil {
			t.Fatalf("SetActiveProfile(all) error = %v", err)
		}
		if state.ActiveCommandProfileID != commandProfile.ID || state.ActiveVariableProfileID != variableProfile.ID || state.ActiveSettingsProfileID != settingsProfile.ID {
			t.Fatalf("SetActiveProfile(all) unexpected state = %+v", *state)
		}

		persisted, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState() error = %v", err)
		}
		if persisted.ID != state.ID ||
			persisted.ActiveCommandProfileID != state.ActiveCommandProfileID ||
			persisted.ActiveVariableProfileID != state.ActiveVariableProfileID ||
			persisted.ActiveSettingsProfileID != state.ActiveSettingsProfileID {
			t.Fatalf("persisted state mismatch: got %+v want %+v", *persisted, *state)
		}
	})
}

func TestProfileRepositoryGetActiveProfiles(t *testing.T) {
	repo, db := setupProfileRepositoryTest(t)
	commandProfile := createProfile(t, db, "active-command", nil)
	variableProfile := createProfile(t, db, "active-variable", nil)
	settingsProfile := createProfile(t, db, "active-settings", nil)

	t.Run("returns the default profile for each slot before any change", func(t *testing.T) {
		cmd, err := repo.GetActiveCommandProfile()
		if err != nil {
			t.Fatalf("GetActiveCommandProfile() error = %v", err)
		}
		if cmd.ID != 1 || cmd.Name != DefaultProfileName {
			t.Fatalf("GetActiveCommandProfile() = %+v, want id=1 name=%q", cmd, DefaultProfileName)
		}

		v, err := repo.GetActiveVariableProfile()
		if err != nil {
			t.Fatalf("GetActiveVariableProfile() error = %v", err)
		}
		if v.ID != 1 || v.Name != DefaultProfileName {
			t.Fatalf("GetActiveVariableProfile() = %+v, want id=1 name=%q", v, DefaultProfileName)
		}

		s, err := repo.GetActiveSettingsProfile()
		if err != nil {
			t.Fatalf("GetActiveSettingsProfile() error = %v", err)
		}
		if s.ID != 1 || s.Name != DefaultProfileName {
			t.Fatalf("GetActiveSettingsProfile() = %+v, want id=1 name=%q", s, DefaultProfileName)
		}
	})

	t.Run("reflects changes made via SetActiveProfile", func(t *testing.T) {
		if _, err := repo.SetActiveProfile(&commandProfile.ID, &variableProfile.ID, &settingsProfile.ID); err != nil {
			t.Fatalf("SetActiveProfile() error = %v", err)
		}

		cmd, err := repo.GetActiveCommandProfile()
		if err != nil {
			t.Fatalf("GetActiveCommandProfile() error = %v", err)
		}
		if cmd.ID != commandProfile.ID || cmd.Name != "active-command" {
			t.Fatalf("GetActiveCommandProfile() = %+v, want id=%d name=active-command", cmd, commandProfile.ID)
		}

		v, err := repo.GetActiveVariableProfile()
		if err != nil {
			t.Fatalf("GetActiveVariableProfile() error = %v", err)
		}
		if v.ID != variableProfile.ID || v.Name != "active-variable" {
			t.Fatalf("GetActiveVariableProfile() = %+v, want id=%d name=active-variable", v, variableProfile.ID)
		}

		s, err := repo.GetActiveSettingsProfile()
		if err != nil {
			t.Fatalf("GetActiveSettingsProfile() error = %v", err)
		}
		if s.ID != settingsProfile.ID || s.Name != "active-settings" {
			t.Fatalf("GetActiveSettingsProfile() = %+v, want id=%d name=active-settings", s, settingsProfile.ID)
		}
	})

	t.Run("wraps the error when profile state is missing", func(t *testing.T) {
		if err := db.Delete(&models.ProfileState{}, 1).Error; err != nil {
			t.Fatalf("delete state row error = %v", err)
		}

		if _, err := repo.GetActiveCommandProfile(); err == nil || !strings.Contains(err.Error(), "loading profile state") {
			t.Fatalf("GetActiveCommandProfile() error = %v, want wrapped loading error", err)
		}
		if _, err := repo.GetActiveVariableProfile(); err == nil || !strings.Contains(err.Error(), "loading profile state") {
			t.Fatalf("GetActiveVariableProfile() error = %v, want wrapped loading error", err)
		}
		if _, err := repo.GetActiveSettingsProfile(); err == nil || !strings.Contains(err.Error(), "loading profile state") {
			t.Fatalf("GetActiveSettingsProfile() error = %v, want wrapped loading error", err)
		}
	})
}

// --- WithTx ---

func TestProfileRepositoryWithTx(t *testing.T) {
	t.Run("operations persist after commit", func(t *testing.T) {
		repo, db := setupProfileRepositoryTest(t)
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		txRepo := repo.WithTx(tx)

		if _, err := txRepo.Create("tx-commit", nil); err != nil {
			t.Fatalf("Create() within tx error = %v", err)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit tx: %v", err)
		}

		if _, err := repo.GetByName("tx-commit"); err != nil {
			t.Fatalf("GetByName() after commit error = %v", err)
		}
	})

	t.Run("operations discarded on rollback", func(t *testing.T) {
		repo, db := setupProfileRepositoryTest(t)
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		txRepo := repo.WithTx(tx)

		if _, err := txRepo.Create("tx-rollback", nil); err != nil {
			t.Fatalf("Create() within tx error = %v", err)
		}
		if err := tx.Rollback().Error; err != nil {
			t.Fatalf("rollback tx: %v", err)
		}

		if _, err := repo.GetByName("tx-rollback"); !errors.Is(err, ErrProfileNotFound) {
			t.Fatalf("GetByName() after rollback error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("preserves validator", func(t *testing.T) {
		repo, db := setupProfileRepositoryTest(t)
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		txRepo := repo.WithTx(tx)

		_, err := txRepo.Create("has space", nil)
		if err == nil {
			t.Fatalf("Create() within tx error = nil, want validation error")
		}
		if err := tx.Rollback().Error; err != nil {
			t.Fatalf("rollback tx: %v", err)
		}
	})

	t.Run("reads within the transaction see uncommitted writes", func(t *testing.T) {
		repo, db := setupProfileRepositoryTest(t)
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		txRepo := repo.WithTx(tx)

		created, err := txRepo.Create("tx-visible", nil)
		if err != nil {
			t.Fatalf("Create() within tx error = %v", err)
		}

		got, err := txRepo.GetByID(created.ID)
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
