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

func setupProfileServiceTest(t *testing.T) (ProfileService, repository.ProfileRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:profile-service-%d?mode=memory&cache=shared", time.Now().UnixNano())
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
	svc := NewProfileService(profileRepo)
	return svc, profileRepo, db
}

func mustCreateServiceProfileViaSvc(t *testing.T, svc ProfileService, name string, description *string) *models.Profile {
	t.Helper()
	p, err := svc.CreateProfile(name, description)
	if err != nil {
		t.Fatalf("CreateProfile(%q) error = %v", name, err)
	}
	return p
}

// --- CreateProfile ---

func TestProfileServiceCreateProfile(t *testing.T) {
	t.Run("creates profile", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		p, err := svc.CreateProfile("work", nil)
		if err != nil {
			t.Fatalf("CreateProfile() error = %v", err)
		}
		if p.ID == 0 {
			t.Fatalf("CreateProfile() returned zero ID")
		}
		if p.Name != "work" {
			t.Fatalf("Name = %q, want %q", p.Name, "work")
		}
	})

	t.Run("creates profile with description", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		description := "workspace profile"
		p, err := svc.CreateProfile("work", &description)
		if err != nil {
			t.Fatalf("CreateProfile() error = %v", err)
		}
		if p.Description == nil || *p.Description != description {
			t.Fatalf("Description = %v, want %q", p.Description, description)
		}
	})

	t.Run("duplicate name returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "dup", nil)

		_, err := svc.CreateProfile("dup", nil)
		if !errors.Is(err, repository.ErrProfileNameExists) {
			t.Fatalf("CreateProfile() error = %v, want ErrProfileNameExists", err)
		}
	})

	t.Run("invalid name returns validation error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		_, err := svc.CreateProfile("has space", nil)
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("CreateProfile() error = %v, want ErrValidation", err)
		}
	})
}

// --- UpdateProfile ---

func TestProfileServiceUpdateProfile(t *testing.T) {
	t.Run("updates name", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "rename-me", nil)

		newName := "renamed"
		updated, err := svc.UpdateProfile("rename-me", UpdateProfileConfig{NewName: &newName})
		if err != nil {
			t.Fatalf("UpdateProfile() error = %v", err)
		}
		if updated.Name != "renamed" {
			t.Fatalf("Name = %q, want %q", updated.Name, "renamed")
		}
	})

	t.Run("updates description", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "described", nil)

		newDescription := "a new description"
		updated, err := svc.UpdateProfile("described", UpdateProfileConfig{Description: &newDescription})
		if err != nil {
			t.Fatalf("UpdateProfile() error = %v", err)
		}
		if updated.Description == nil || *updated.Description != newDescription {
			t.Fatalf("Description = %v, want %q", updated.Description, newDescription)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		_, err := svc.UpdateProfile("missing", UpdateProfileConfig{})
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("UpdateProfile() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("prevents renaming default profile", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		newName := "renamed-default"
		_, err := svc.UpdateProfile(repository.DefaultProfileName, UpdateProfileConfig{NewName: &newName})
		if !errors.Is(err, repository.ErrDefaultProfileName) {
			t.Fatalf("UpdateProfile() error = %v, want ErrDefaultProfileName", err)
		}
	})

	t.Run("name conflict returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "taken", nil)
		mustCreateServiceProfileViaSvc(t, svc, "renaming", nil)

		newName := "taken"
		_, err := svc.UpdateProfile("renaming", UpdateProfileConfig{NewName: &newName})
		if !errors.Is(err, repository.ErrProfileNameExists) {
			t.Fatalf("UpdateProfile() error = %v, want ErrProfileNameExists", err)
		}
	})
}

// --- DeleteProfile ---

func TestProfileServiceDeleteProfile(t *testing.T) {
	t.Run("deletes existing profile", func(t *testing.T) {
		svc, profileRepo, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "delete-me", nil)

		if err := svc.DeleteProfile("delete-me", false); err != nil {
			t.Fatalf("DeleteProfile() error = %v", err)
		}
		if _, err := profileRepo.GetByName("delete-me"); !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("profile still exists after delete: err = %v", err)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		err := svc.DeleteProfile("missing", false)
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("DeleteProfile() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("prevents deleting default profile", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		err := svc.DeleteProfile(repository.DefaultProfileName, false)
		if !errors.Is(err, repository.ErrDefaultProfileDelete) {
			t.Fatalf("DeleteProfile() error = %v, want ErrDefaultProfileDelete", err)
		}
	})

	t.Run("prevents deleting active profile", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "active-one", nil)
		if _, err := svc.SwitchProfile("active-one"); err != nil {
			t.Fatalf("SwitchProfile() error = %v", err)
		}

		err := svc.DeleteProfile("active-one", false)
		if !errors.Is(err, repository.ErrProfileInUse) {
			t.Fatalf("DeleteProfile() error = %v, want ErrProfileInUse", err)
		}
	})

	t.Run("blocks deletion when profile has content and force is false", func(t *testing.T) {
		svc, _, db := setupProfileServiceTest(t)
		profile := mustCreateServiceProfileViaSvc(t, svc, "with-content", nil)
		if err := db.Create(&models.Command{Alias: "cmd", Template: "echo hi", ProfileID: profile.ID}).Error; err != nil {
			t.Fatalf("create command: %v", err)
		}

		err := svc.DeleteProfile("with-content", false)
		if !errors.Is(err, repository.ErrProfileHasContent) {
			t.Fatalf("DeleteProfile() error = %v, want ErrProfileHasContent", err)
		}
	})

	t.Run("force delete ignores content", func(t *testing.T) {
		svc, profileRepo, db := setupProfileServiceTest(t)
		profile := mustCreateServiceProfileViaSvc(t, svc, "force-delete", nil)
		if err := db.Create(&models.Command{Alias: "cmd", Template: "echo hi", ProfileID: profile.ID}).Error; err != nil {
			t.Fatalf("create command: %v", err)
		}

		if err := svc.DeleteProfile("force-delete", true); err != nil {
			t.Fatalf("DeleteProfile(force=true) error = %v", err)
		}
		if _, err := profileRepo.GetByName("force-delete"); !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("profile still exists after force delete: err = %v", err)
		}
	})
}

// --- GetProfile ---

func TestProfileServiceGetProfile(t *testing.T) {
	t.Run("returns profile when found", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		created := mustCreateServiceProfileViaSvc(t, svc, "work", nil)

		got, err := svc.GetProfile("work")
		if err != nil {
			t.Fatalf("GetProfile() error = %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("GetProfile() id = %d, want %d", got.ID, created.ID)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		_, err := svc.GetProfile("missing")
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetProfile() error = %v, want ErrProfileNotFound", err)
		}
	})
}

// --- ListProfiles ---

func TestProfileServiceListProfiles(t *testing.T) {
	t.Run("lists all profiles with default ordering", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "zeta", nil)
		mustCreateServiceProfileViaSvc(t, svc, "alpha", nil)

		profiles, err := svc.ListProfiles("", nil)
		if err != nil {
			t.Fatalf("ListProfiles() error = %v", err)
		}
		if len(profiles) != 3 {
			t.Fatalf("len(profiles) = %d, want 3", len(profiles))
		}
		if profiles[0].Name != "alpha" {
			t.Fatalf("profiles[0].Name = %q, want %q", profiles[0].Name, "alpha")
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "a", nil)
		mustCreateServiceProfileViaSvc(t, svc, "b", nil)

		limit := 1
		profiles, err := svc.ListProfiles("", &limit)
		if err != nil {
			t.Fatalf("ListProfiles() error = %v", err)
		}
		if len(profiles) != 1 {
			t.Fatalf("len(profiles) = %d, want 1", len(profiles))
		}
	})

	t.Run("invalid order field returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		_, err := svc.ListProfiles("bogus", nil)
		if err == nil {
			t.Fatalf("ListProfiles() error = nil, want error")
		}
	})
}

// --- SearchProfiles ---

func TestProfileServiceSearchProfiles(t *testing.T) {
	t.Run("uses default fields", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "apple", nil)
		mustCreateServiceProfileViaSvc(t, svc, "banana", nil)

		profiles, err := svc.SearchProfiles("apple", nil, nil)
		if err != nil {
			t.Fatalf("SearchProfiles() error = %v", err)
		}
		if len(profiles) != 1 || profiles[0].Name != "apple" {
			t.Fatalf("profiles = %+v, want only [apple]", profiles)
		}
	})

	t.Run("respects explicit fields", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		description := "a fruit"
		mustCreateServiceProfileViaSvc(t, svc, "apple", &description)
		mustCreateServiceProfileViaSvc(t, svc, "unrelated", &description)

		profiles, err := svc.SearchProfiles("apple", []string{"name"}, nil)
		if err != nil {
			t.Fatalf("SearchProfiles() error = %v", err)
		}
		if len(profiles) != 1 || profiles[0].Name != "apple" {
			t.Fatalf("profiles = %+v, want only [apple]", profiles)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "apple", nil)
		mustCreateServiceProfileViaSvc(t, svc, "apple-two", nil)

		limit := 1
		profiles, err := svc.SearchProfiles("apple", nil, &limit)
		if err != nil {
			t.Fatalf("SearchProfiles() error = %v", err)
		}
		if len(profiles) != 1 {
			t.Fatalf("len(profiles) = %d, want 1", len(profiles))
		}
	})

	t.Run("invalid explicit field returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		_, err := svc.SearchProfiles("apple", []string{"bogus"}, nil)
		if err == nil {
			t.Fatalf("SearchProfiles() error = nil, want error")
		}
	})
}

// --- SwitchProfile / SwitchCommandProfile / SwitchVariableProfile / SwitchSettingsProfile ---

func TestProfileServiceSwitchProfile(t *testing.T) {
	t.Run("sets all three active profile slots", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		target := mustCreateServiceProfileViaSvc(t, svc, "target", nil)

		state, err := svc.SwitchProfile("target")
		if err != nil {
			t.Fatalf("SwitchProfile() error = %v", err)
		}
		if state.ActiveCommandProfileID != target.ID || state.ActiveVariableProfileID != target.ID || state.ActiveSettingsProfileID != target.ID {
			t.Fatalf("SwitchProfile() state = %+v, want all slots = %d", *state, target.ID)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		_, err := svc.SwitchProfile("missing")
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("SwitchProfile() error = %v, want ErrProfileNotFound", err)
		}
	})
}

func TestProfileServiceSwitchCommandProfile(t *testing.T) {
	t.Run("sets only command profile slot", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		target := mustCreateServiceProfileViaSvc(t, svc, "cmd-target", nil)

		state, err := svc.SwitchCommandProfile("cmd-target")
		if err != nil {
			t.Fatalf("SwitchCommandProfile() error = %v", err)
		}
		if state.ActiveCommandProfileID != target.ID {
			t.Fatalf("ActiveCommandProfileID = %d, want %d", state.ActiveCommandProfileID, target.ID)
		}
		if state.ActiveVariableProfileID == target.ID || state.ActiveSettingsProfileID == target.ID {
			t.Fatalf("SwitchCommandProfile() unexpectedly changed other slots: %+v", *state)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		_, err := svc.SwitchCommandProfile("missing")
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("SwitchCommandProfile() error = %v, want ErrProfileNotFound", err)
		}
	})
}

func TestProfileServiceSwitchVariableProfile(t *testing.T) {
	t.Run("sets only variable profile slot", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		target := mustCreateServiceProfileViaSvc(t, svc, "var-target", nil)

		state, err := svc.SwitchVariableProfile("var-target")
		if err != nil {
			t.Fatalf("SwitchVariableProfile() error = %v", err)
		}
		if state.ActiveVariableProfileID != target.ID {
			t.Fatalf("ActiveVariableProfileID = %d, want %d", state.ActiveVariableProfileID, target.ID)
		}
		if state.ActiveCommandProfileID == target.ID || state.ActiveSettingsProfileID == target.ID {
			t.Fatalf("SwitchVariableProfile() unexpectedly changed other slots: %+v", *state)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		_, err := svc.SwitchVariableProfile("missing")
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("SwitchVariableProfile() error = %v, want ErrProfileNotFound", err)
		}
	})
}

func TestProfileServiceSwitchSettingsProfile(t *testing.T) {
	t.Run("sets only settings profile slot", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		target := mustCreateServiceProfileViaSvc(t, svc, "settings-target", nil)

		state, err := svc.SwitchSettingsProfile("settings-target")
		if err != nil {
			t.Fatalf("SwitchSettingsProfile() error = %v", err)
		}
		if state.ActiveSettingsProfileID != target.ID {
			t.Fatalf("ActiveSettingsProfileID = %d, want %d", state.ActiveSettingsProfileID, target.ID)
		}
		if state.ActiveCommandProfileID == target.ID || state.ActiveVariableProfileID == target.ID {
			t.Fatalf("SwitchSettingsProfile() unexpectedly changed other slots: %+v", *state)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		_, err := svc.SwitchSettingsProfile("missing")
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("SwitchSettingsProfile() error = %v, want ErrProfileNotFound", err)
		}
	})
}

// --- GetStatus ---

func TestProfileServiceGetStatus(t *testing.T) {
	t.Run("returns default profile in all slots before any switch", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)

		status, err := svc.GetStatus()
		if err != nil {
			t.Fatalf("GetStatus() error = %v", err)
		}
		if status.CommandProfile != repository.DefaultProfileName ||
			status.VariableProfile != repository.DefaultProfileName ||
			status.SettingsProfile != repository.DefaultProfileName {
			t.Fatalf("GetStatus() = %+v, want all slots = %q", status, repository.DefaultProfileName)
		}
		if !status.Linked() {
			t.Fatalf("GetStatus() expected Linked() = true")
		}
	})

	t.Run("reflects independent switches", func(t *testing.T) {
		svc, _, _ := setupProfileServiceTest(t)
		mustCreateServiceProfileViaSvc(t, svc, "cmd-only", nil)

		if _, err := svc.SwitchCommandProfile("cmd-only"); err != nil {
			t.Fatalf("SwitchCommandProfile() error = %v", err)
		}

		status, err := svc.GetStatus()
		if err != nil {
			t.Fatalf("GetStatus() error = %v", err)
		}
		if status.CommandProfile != "cmd-only" {
			t.Fatalf("CommandProfile = %q, want %q", status.CommandProfile, "cmd-only")
		}
		if status.VariableProfile != repository.DefaultProfileName || status.SettingsProfile != repository.DefaultProfileName {
			t.Fatalf("GetStatus() = %+v, want variable/settings still %q", status, repository.DefaultProfileName)
		}
		if status.Linked() {
			t.Fatalf("GetStatus() expected Linked() = false")
		}
	})
}
