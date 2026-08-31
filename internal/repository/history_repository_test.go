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

func setupHistoryRepositoryTest(t *testing.T) (HistoryRepository, ProfileRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:history-repo-%d?mode=memory&cache=shared", time.Now().UnixNano())
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
		&models.CommandHistory{},
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
	historyRepo := NewHistoryRepository(db, profileRepo)
	return historyRepo, profileRepo, db
}

// mustCreateProfile creates an additional (non-default) profile for cross-profile tests.
func mustCreateProfile(t *testing.T, profileRepo ProfileRepository, name string) *models.Profile {
	t.Helper()
	p, err := profileRepo.Create(name, nil)
	if err != nil {
		t.Fatalf("Create profile(%q) error = %v", name, err)
	}
	return p
}

func mustRecord(t *testing.T, repo HistoryRepository, entry HistoryEntry) *models.CommandHistory {
	t.Helper()
	h, err := repo.Record(entry)
	if err != nil {
		t.Fatalf("Record(%+v) error = %v", entry, err)
	}
	return h
}

// --- Record ---

func TestHistoryRepositoryRecord(t *testing.T) {
	repo, _, db := setupHistoryRepositoryTest(t)

	t.Run("persists a new entry", func(t *testing.T) {
		entry := mustRecord(t, repo, HistoryEntry{
			Alias:    "build",
			Template: "go build {{.pkg}}",
			Resolved: "go build ./...",
		})

		if entry.ID == "" {
			t.Fatalf("Record() returned empty ID")
		}
		if entry.Alias != "build" {
			t.Fatalf("Alias = %q, want %q", entry.Alias, "build")
		}
		if entry.Template != "go build {{.pkg}}" {
			t.Fatalf("Template = %q, want %q", entry.Template, "go build {{.pkg}}")
		}
		if entry.Resolved != "go build ./..." {
			t.Fatalf("Resolved = %q, want %q", entry.Resolved, "go build ./...")
		}
		if entry.RanAt.IsZero() {
			t.Fatalf("RanAt is zero, want a timestamp")
		}

		var count int64
		if err := db.Model(&models.CommandHistory{}).Count(&count).Error; err != nil {
			t.Fatalf("count error = %v", err)
		}
		if count != 1 {
			t.Fatalf("rows in DB = %d, want 1", count)
		}

		fromDB, err := repo.GetByID(entry.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() after Record error = %v", err)
		}
		if fromDB.Resolved != "go build ./..." {
			t.Fatalf("persisted Resolved = %q, want %q", fromDB.Resolved, "go build ./...")
		}
	})

	t.Run("assigns unique hex IDs across entries", func(t *testing.T) {
		a := mustRecord(t, repo, HistoryEntry{Alias: "x", Template: "t", Resolved: "r"})
		b := mustRecord(t, repo, HistoryEntry{Alias: "x", Template: "t", Resolved: "r"})
		if a.ID == b.ID {
			t.Fatalf("Record() produced duplicate IDs: %q", a.ID)
		}
		if len(a.ID) != 32 {
			t.Fatalf("len(ID) = %d, want 32 (hex-encoded UUID)", len(a.ID))
		}
	})

	t.Run("stores optional fields when provided", func(t *testing.T) {
		vars := `{"name":"value"}`
		exit := 0
		entry := mustRecord(t, repo, HistoryEntry{
			Alias:         "with-opts",
			Template:      "t",
			Resolved:      "r",
			VariablesUsed: &vars,
			ExitCode:      &exit,
		})
		if entry.VariablesUsed == nil || *entry.VariablesUsed != vars {
			t.Fatalf("VariablesUsed = %v, want %q", entry.VariablesUsed, vars)
		}
		if entry.ExitCode == nil || *entry.ExitCode != 0 {
			t.Fatalf("ExitCode = %v, want 0", entry.ExitCode)
		}
	})

	t.Run("leaves optional fields nil when omitted", func(t *testing.T) {
		entry := mustRecord(t, repo, HistoryEntry{Alias: "no-opts", Template: "t", Resolved: "r"})
		if entry.VariablesUsed != nil {
			t.Fatalf("VariablesUsed = %v, want nil", *entry.VariablesUsed)
		}
		if entry.ExitCode != nil {
			t.Fatalf("ExitCode = %v, want nil", *entry.ExitCode)
		}
	})

	t.Run("defaults to active command profile when ProfileID is nil", func(t *testing.T) {
		entry := mustRecord(t, repo, HistoryEntry{Alias: "default-profile", Template: "t", Resolved: "r"})

		defaultProfile, gErr := repo.GetByID(entry.ID, nil)
		if gErr != nil {
			t.Fatalf("GetByID() error = %v", gErr)
		}
		if defaultProfile.ProfileID == 0 {
			t.Fatalf("ProfileID = 0, want the resolved active profile id")
		}
	})
}

// --- Record with explicit profile ---

func TestHistoryRepositoryRecordExplicitProfile(t *testing.T) {
	repo, profileRepo, _ := setupHistoryRepositoryTest(t)
	other := mustCreateProfile(t, profileRepo, "other")

	entry := mustRecord(t, repo, HistoryEntry{
		Alias:     "scoped",
		Template:  "t",
		Resolved:  "r",
		ProfileID: &other.ID,
	})
	if entry.ProfileID != other.ID {
		t.Fatalf("ProfileID = %d, want %d", entry.ProfileID, other.ID)
	}

	// Not visible under the default (active) profile.
	_, err := repo.GetByID(entry.ID, nil)
	if !errors.Is(err, ErrUnknownCommandHistory) {
		t.Fatalf("GetByID() under default profile error = %v, want ErrUnknownCommandHistory", err)
	}

	// Visible when explicitly scoped to the other profile.
	fromDB, err := repo.GetByID(entry.ID, &other.ID)
	if err != nil {
		t.Fatalf("GetByID() under explicit profile error = %v", err)
	}
	if fromDB.ID != entry.ID {
		t.Fatalf("GetByID() id = %q, want %q", fromDB.ID, entry.ID)
	}
}

// --- Record retention ---

func TestHistoryRepositoryRecordRetention(t *testing.T) {
	t.Run("nil limit keeps all entries", func(t *testing.T) {
		repo, _, _ := setupHistoryRepositoryTest(t)
		for i := 0; i < 5; i++ {
			mustRecord(t, repo, HistoryEntry{Alias: "keep-all", Template: "t", Resolved: "r"})
		}
		entries, err := repo.GetRecent(strPtr("keep-all"), 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 5 {
			t.Fatalf("len(entries) = %d, want 5", len(entries))
		}
	})

	t.Run("zero or negative limit keeps all entries", func(t *testing.T) {
		repo, _, _ := setupHistoryRepositoryTest(t)
		for i := 0; i < 3; i++ {
			mustRecord(t, repo, HistoryEntry{Alias: "zero-limit", Template: "t", Resolved: "r", RetentionLimit: intPtr(0)})
		}
		for i := 0; i < 3; i++ {
			mustRecord(t, repo, HistoryEntry{Alias: "neg-limit", Template: "t", Resolved: "r", RetentionLimit: intPtr(-1)})
		}
		zeroEntries, err := repo.GetRecent(strPtr("zero-limit"), 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(zeroEntries) != 3 {
			t.Fatalf("len(zeroEntries) = %d, want 3", len(zeroEntries))
		}
		negEntries, err := repo.GetRecent(strPtr("neg-limit"), 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(negEntries) != 3 {
			t.Fatalf("len(negEntries) = %d, want 3", len(negEntries))
		}
	})

	t.Run("trims older entries beyond the limit", func(t *testing.T) {
		repo, _, _ := setupHistoryRepositoryTest(t)
		var ids []string
		for i := 0; i < 5; i++ {
			e := mustRecord(t, repo, HistoryEntry{Alias: "trim-me", Template: "t", Resolved: "r", RetentionLimit: intPtr(3)})
			ids = append(ids, e.ID)
			time.Sleep(time.Millisecond)
		}

		entries, err := repo.GetRecent(strPtr("trim-me"), 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 3 {
			t.Fatalf("len(entries) = %d, want 3", len(entries))
		}

		// The three most recently recorded entries (last three IDs) should survive.
		want := map[string]bool{ids[2]: true, ids[3]: true, ids[4]: true}
		for _, e := range entries {
			if !want[e.ID] {
				t.Fatalf("unexpected surviving entry id %q", e.ID)
			}
		}
	})

	t.Run("retention only affects the matching alias and profile", func(t *testing.T) {
		repo, profileRepo, _ := setupHistoryRepositoryTest(t)
		other := mustCreateProfile(t, profileRepo, "other-profile")

		for i := 0; i < 3; i++ {
			mustRecord(t, repo, HistoryEntry{Alias: "shared-alias", Template: "t", Resolved: "r", RetentionLimit: intPtr(1)})
		}
		for i := 0; i < 3; i++ {
			mustRecord(t, repo, HistoryEntry{Alias: "other-alias", Template: "t", Resolved: "r", RetentionLimit: intPtr(1)})
		}
		for i := 0; i < 3; i++ {
			mustRecord(t, repo, HistoryEntry{Alias: "shared-alias", Template: "t", Resolved: "r", ProfileID: &other.ID})
		}

		sameAliasOtherProfile, err := repo.GetRecent(strPtr("shared-alias"), 0, &other.ID)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(sameAliasOtherProfile) != 3 {
			t.Fatalf("len(sameAliasOtherProfile) = %d, want 3 (retention on default profile must not affect other profile)", len(sameAliasOtherProfile))
		}

		otherAliasEntries, err := repo.GetRecent(strPtr("other-alias"), 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(otherAliasEntries) != 1 {
			t.Fatalf("len(otherAliasEntries) = %d, want 1 (retention on shared-alias must not affect other-alias)", len(otherAliasEntries))
		}
	})
}

// --- GetByID ---

func TestHistoryRepositoryGetByID(t *testing.T) {
	repo, _, _ := setupHistoryRepositoryTest(t)
	entry := mustRecord(t, repo, HistoryEntry{Alias: "lookup", Template: "t", Resolved: "r"})

	t.Run("returns entry by id", func(t *testing.T) {
		got, err := repo.GetByID(entry.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got.ID != entry.ID {
			t.Fatalf("GetByID() id = %q, want %q", got.ID, entry.ID)
		}
	})

	t.Run("returns ErrUnknownCommandHistory when missing", func(t *testing.T) {
		_, err := repo.GetByID("does-not-exist", nil)
		if !errors.Is(err, ErrUnknownCommandHistory) {
			t.Fatalf("GetByID() error = %v, want ErrUnknownCommandHistory", err)
		}
	})

	t.Run("scopes lookup to the given profile", func(t *testing.T) {
		_, err := repo.GetByID(entry.ID, intPtrFromUint(999999))
		if !errors.Is(err, ErrUnknownCommandHistory) {
			t.Fatalf("GetByID() with wrong profile error = %v, want ErrUnknownCommandHistory", err)
		}
	})
}

func intPtrFromUint(v uint) *uint {
	return &v
}

// --- GetRecent ---

func TestHistoryRepositoryGetRecent(t *testing.T) {
	repo, profileRepo, _ := setupHistoryRepositoryTest(t)

	first := mustRecord(t, repo, HistoryEntry{Alias: "a", Template: "t", Resolved: "r1"})
	time.Sleep(time.Millisecond)
	second := mustRecord(t, repo, HistoryEntry{Alias: "a", Template: "t", Resolved: "r2"})
	time.Sleep(time.Millisecond)
	third := mustRecord(t, repo, HistoryEntry{Alias: "b", Template: "t", Resolved: "r3"})

	t.Run("filters by alias and orders most recent first", func(t *testing.T) {
		entries, err := repo.GetRecent(strPtr("a"), 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("len(entries) = %d, want 2", len(entries))
		}
		if entries[0].ID != second.ID || entries[1].ID != first.ID {
			t.Fatalf("entries = %+v, want [second, first]", entries)
		}
	})

	t.Run("nil alias returns entries across all aliases", func(t *testing.T) {
		entries, err := repo.GetRecent(nil, 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 3 {
			t.Fatalf("len(entries) = %d, want 3", len(entries))
		}
		if entries[0].ID != third.ID {
			t.Fatalf("entries[0].ID = %q, want %q (most recent first)", entries[0].ID, third.ID)
		}
	})

	t.Run("defaults limit to 25 when non-positive", func(t *testing.T) {
		for i := 0; i < 30; i++ {
			mustRecord(t, repo, HistoryEntry{Alias: "many", Template: "t", Resolved: "r"})
		}
		entries, err := repo.GetRecent(strPtr("many"), 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 25 {
			t.Fatalf("len(entries) = %d, want 25", len(entries))
		}

		negEntries, err := repo.GetRecent(strPtr("many"), -5, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(negEntries) != 25 {
			t.Fatalf("len(negEntries) = %d, want 25", len(negEntries))
		}
	})

	t.Run("respects explicit limit", func(t *testing.T) {
		entries, err := repo.GetRecent(nil, 1, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("len(entries) = %d, want 1", len(entries))
		}
	})

	t.Run("scopes to profile", func(t *testing.T) {
		other := mustCreateProfile(t, profileRepo, "scoped-profile")
		mustRecord(t, repo, HistoryEntry{Alias: "a", Template: "t", Resolved: "other-profile-r", ProfileID: &other.ID})

		defaultEntries, err := repo.GetRecent(strPtr("a"), 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		for _, e := range defaultEntries {
			if e.Resolved == "other-profile-r" {
				t.Fatalf("GetRecent() under default profile leaked entry from other profile: %+v", e)
			}
		}

		otherEntries, err := repo.GetRecent(strPtr("a"), 0, &other.ID)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(otherEntries) != 1 || otherEntries[0].Resolved != "other-profile-r" {
			t.Fatalf("otherEntries = %+v, want a single other-profile-r entry", otherEntries)
		}
	})

	t.Run("no matches returns empty slice", func(t *testing.T) {
		entries, err := repo.GetRecent(strPtr("nonexistent"), 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("entries = %+v, want empty", entries)
		}
	})
}

// --- DeleteByID ---

func TestHistoryRepositoryDeleteByID(t *testing.T) {
	repo, profileRepo, db := setupHistoryRepositoryTest(t)

	t.Run("deletes an existing entry", func(t *testing.T) {
		entry := mustRecord(t, repo, HistoryEntry{Alias: "del-me", Template: "t", Resolved: "r"})

		if err := repo.DeleteByID(entry.ID, nil); err != nil {
			t.Fatalf("DeleteByID() error = %v", err)
		}

		_, err := repo.GetByID(entry.ID, nil)
		if !errors.Is(err, ErrUnknownCommandHistory) {
			t.Fatalf("GetByID() after delete error = %v, want ErrUnknownCommandHistory", err)
		}

		var count int64
		db.Model(&models.CommandHistory{}).Where("id = ?", entry.ID).Count(&count)
		if count != 0 {
			t.Fatalf("rows remaining in DB = %d, want 0", count)
		}
	})

	t.Run("returns error when entry does not exist", func(t *testing.T) {
		err := repo.DeleteByID("does-not-exist", nil)
		if !errors.Is(err, ErrUnknownCommandHistory) {
			t.Fatalf("DeleteByID() error = %v, want ErrUnknownCommandHistory", err)
		}
	})

	t.Run("does not delete entry belonging to a different profile", func(t *testing.T) {
		other := mustCreateProfile(t, profileRepo, "protect-me")
		entry := mustRecord(t, repo, HistoryEntry{Alias: "protected", Template: "t", Resolved: "r", ProfileID: &other.ID})

		err := repo.DeleteByID(entry.ID, nil)
		if !errors.Is(err, ErrUnknownCommandHistory) {
			t.Fatalf("DeleteByID() under wrong profile error = %v, want ErrUnknownCommandHistory", err)
		}

		fromDB, gErr := repo.GetByID(entry.ID, &other.ID)
		if gErr != nil {
			t.Fatalf("GetByID() after failed cross-profile delete error = %v", gErr)
		}
		if fromDB.ID != entry.ID {
			t.Fatalf("entry was unexpectedly removed")
		}
	})
}

// --- Clear ---

func TestHistoryRepositoryClear(t *testing.T) {
	t.Run("clears all entries for the active profile when alias is nil", func(t *testing.T) {
		repo, _, db := setupHistoryRepositoryTest(t)
		mustRecord(t, repo, HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})
		mustRecord(t, repo, HistoryEntry{Alias: "b", Template: "t", Resolved: "r"})
		mustRecord(t, repo, HistoryEntry{Alias: "c", Template: "t", Resolved: "r"})

		n, err := repo.Clear(nil, nil)
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}
		if n != 3 {
			t.Fatalf("Clear() returned %d, want 3", n)
		}

		var count int64
		db.Model(&models.CommandHistory{}).Count(&count)
		if count != 0 {
			t.Fatalf("rows remaining in DB = %d, want 0", count)
		}
	})

	t.Run("clears only the given alias when provided", func(t *testing.T) {
		repo, _, _ := setupHistoryRepositoryTest(t)
		mustRecord(t, repo, HistoryEntry{Alias: "keep", Template: "t", Resolved: "r"})
		mustRecord(t, repo, HistoryEntry{Alias: "clear-me", Template: "t", Resolved: "r"})
		mustRecord(t, repo, HistoryEntry{Alias: "clear-me", Template: "t", Resolved: "r"})

		n, err := repo.Clear(strPtr("clear-me"), nil)
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}
		if n != 2 {
			t.Fatalf("Clear() returned %d, want 2", n)
		}

		remaining, err := repo.GetRecent(nil, 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(remaining) != 1 || remaining[0].Alias != "keep" {
			t.Fatalf("remaining = %+v, want only [keep]", remaining)
		}
	})

	t.Run("does not affect other profiles", func(t *testing.T) {
		repo, profileRepo, _ := setupHistoryRepositoryTest(t)
		other := mustCreateProfile(t, profileRepo, "untouched")
		mustRecord(t, repo, HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})
		mustRecord(t, repo, HistoryEntry{Alias: "a", Template: "t", Resolved: "r", ProfileID: &other.ID})

		n, err := repo.Clear(nil, nil)
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}
		if n != 1 {
			t.Fatalf("Clear() returned %d, want 1", n)
		}

		otherEntries, err := repo.GetRecent(nil, 0, &other.ID)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(otherEntries) != 1 {
			t.Fatalf("len(otherEntries) = %d, want 1 (Clear must not touch other profiles)", len(otherEntries))
		}
	})

	t.Run("returns zero when nothing matches", func(t *testing.T) {
		repo, _, _ := setupHistoryRepositoryTest(t)
		n, err := repo.Clear(strPtr("nonexistent"), nil)
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}
		if n != 0 {
			t.Fatalf("Clear() returned %d, want 0", n)
		}
	})
}
