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

func setupHistoryServiceTest(t *testing.T) (HistoryService, repository.HistoryRepository, repository.ProfileRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:history-service-%d?mode=memory&cache=shared", time.Now().UnixNano())
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
	historyRepo := repository.NewHistoryRepository(db, profileRepo)
	svc := NewHistoryService(historyRepo, profileRepo)
	return svc, historyRepo, profileRepo, db
}

func mustCreateHistoryProfile(t *testing.T, profileRepo repository.ProfileRepository, name string) *models.Profile {
	t.Helper()
	p, err := profileRepo.Create(name, nil)
	if err != nil {
		t.Fatalf("Create profile(%q) error = %v", name, err)
	}
	return p
}

func mustRecordHistory(t *testing.T, historyRepo repository.HistoryRepository, entry repository.HistoryEntry) *models.CommandHistory {
	t.Helper()
	h, err := historyRepo.Record(entry)
	if err != nil {
		t.Fatalf("Record(%+v) error = %v", entry, err)
	}
	return h
}

// --- GetRecent ---

func TestHistoryServiceGetRecent(t *testing.T) {
	t.Run("returns entries most recent first", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		first := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "build", Template: "t", Resolved: "r1"})
		time.Sleep(time.Millisecond)
		second := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "build", Template: "t", Resolved: "r2"})

		entries, err := svc.GetRecent(nil, nil, nil)
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

	t.Run("filters by alias", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "b", Template: "t", Resolved: "r"})

		alias := "a"
		entries, err := svc.GetRecent(&alias, nil, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 1 || entries[0].Alias != "a" {
			t.Fatalf("entries = %+v, want only alias a", entries)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		for i := 0; i < 3; i++ {
			mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "many", Template: "t", Resolved: "r"})
		}

		limit := 2
		entries, err := svc.GetRecent(nil, &limit, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("len(entries) = %d, want 2", len(entries))
		}
	})

	t.Run("nil limit defaults to repository default", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		for i := 0; i < 30; i++ {
			mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "many", Template: "t", Resolved: "r"})
		}

		entries, err := svc.GetRecent(nil, nil, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 25 {
			t.Fatalf("len(entries) = %d, want 25 (repository default)", len(entries))
		}
	})

	t.Run("scopes to named profile", func(t *testing.T) {
		svc, historyRepo, profileRepo, _ := setupHistoryServiceTest(t)
		other := mustCreateHistoryProfile(t, profileRepo, "other")
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "default-r"})
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "other-r", ProfileID: &other.ID})

		otherName := "other"
		entries, err := svc.GetRecent(nil, nil, &otherName)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(entries) != 1 || entries[0].Resolved != "other-r" {
			t.Fatalf("entries = %+v, want only other-r", entries)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _ := setupHistoryServiceTest(t)
		missing := "missing"
		_, err := svc.GetRecent(nil, nil, &missing)
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetRecent() error = %v, want ErrProfileNotFound", err)
		}
	})
}

// --- GetByRef ---

func TestHistoryServiceGetByRef(t *testing.T) {
	t.Run("resolves by id ref", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		entry := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "build", Template: "t", Resolved: "r"})

		got, err := svc.GetByRef(entry.ID, nil, nil)
		if err != nil {
			t.Fatalf("GetByRef() error = %v", err)
		}
		if got.ID != entry.ID {
			t.Fatalf("GetByRef() id = %q, want %q", got.ID, entry.ID)
		}
	})

	t.Run("unknown id ref returns error", func(t *testing.T) {
		svc, _, _, _ := setupHistoryServiceTest(t)
		_, err := svc.GetByRef("does-not-exist", nil, nil)
		if !errors.Is(err, repository.ErrUnknownCommandHistory) {
			t.Fatalf("GetByRef() error = %v, want ErrUnknownCommandHistory", err)
		}
	})

	t.Run("resolves 1-based index against most recent first", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		first := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r1"})
		time.Sleep(time.Millisecond)
		second := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r2"})

		got, err := svc.GetByRef("1", nil, nil)
		if err != nil {
			t.Fatalf("GetByRef(\"1\") error = %v", err)
		}
		if got.ID != second.ID {
			t.Fatalf("GetByRef(\"1\").ID = %q, want most recent %q", got.ID, second.ID)
		}

		got2, err := svc.GetByRef("2", nil, nil)
		if err != nil {
			t.Fatalf("GetByRef(\"2\") error = %v", err)
		}
		if got2.ID != first.ID {
			t.Fatalf("GetByRef(\"2\").ID = %q, want %q", got2.ID, first.ID)
		}
	})

	t.Run("index zero returns ErrUnknownHistoryIndex", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})

		_, err := svc.GetByRef("0", nil, nil)
		if !errors.Is(err, ErrUnknownHistoryIndex) {
			t.Fatalf("GetByRef(\"0\") error = %v, want ErrUnknownHistoryIndex", err)
		}
	})

	t.Run("negative index returns ErrUnknownHistoryIndex", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})

		_, err := svc.GetByRef("-1", nil, nil)
		if !errors.Is(err, ErrUnknownHistoryIndex) {
			t.Fatalf("GetByRef(\"-1\") error = %v, want ErrUnknownHistoryIndex", err)
		}
	})

	t.Run("out of range index returns ErrUnknownHistoryIndex", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})

		_, err := svc.GetByRef("5", nil, nil)
		if !errors.Is(err, ErrUnknownHistoryIndex) {
			t.Fatalf("GetByRef(\"5\") error = %v, want ErrUnknownHistoryIndex", err)
		}
	})

	t.Run("index filters by alias before resolving position", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "a1"})
		time.Sleep(time.Millisecond)
		wantSecond := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "b", Template: "t", Resolved: "b1"})
		time.Sleep(time.Millisecond)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "a2"})

		alias := "b"
		got, err := svc.GetByRef("1", &alias, nil)
		if err != nil {
			t.Fatalf("GetByRef(\"1\") error = %v", err)
		}
		if got.ID != wantSecond.ID {
			t.Fatalf("GetByRef(\"1\") with alias filter = %+v, want %+v", got, wantSecond)
		}
	})

	t.Run("unknown profile name returns error for id ref", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		entry := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})

		missing := "missing"
		_, err := svc.GetByRef(entry.ID, nil, &missing)
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetByRef() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown profile name returns error for index ref", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})

		missing := "missing"
		_, err := svc.GetByRef("1", nil, &missing)
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("GetByRef() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("scopes id ref lookup to named profile", func(t *testing.T) {
		svc, historyRepo, profileRepo, _ := setupHistoryServiceTest(t)
		other := mustCreateHistoryProfile(t, profileRepo, "other")
		entry := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r", ProfileID: &other.ID})

		_, err := svc.GetByRef(entry.ID, nil, nil)
		if !errors.Is(err, repository.ErrUnknownCommandHistory) {
			t.Fatalf("GetByRef() under default profile error = %v, want ErrUnknownCommandHistory", err)
		}

		otherName := "other"
		got, err := svc.GetByRef(entry.ID, nil, &otherName)
		if err != nil {
			t.Fatalf("GetByRef() under explicit profile error = %v", err)
		}
		if got.ID != entry.ID {
			t.Fatalf("GetByRef() id = %q, want %q", got.ID, entry.ID)
		}
	})
}

// --- GetVariables ---

func TestHistoryServiceGetVariables(t *testing.T) {
	t.Run("nil VariablesUsed returns nil, nil", func(t *testing.T) {
		svc, _, _, _ := setupHistoryServiceTest(t)
		entry := &models.CommandHistory{}

		vars, err := svc.GetVariables(entry)
		if err != nil {
			t.Fatalf("GetVariables() error = %v", err)
		}
		if vars != nil {
			t.Fatalf("GetVariables() = %v, want nil", vars)
		}
	})

	t.Run("unmarshals stored JSON variables", func(t *testing.T) {
		svc, _, _, _ := setupHistoryServiceTest(t)
		raw := `{"name":"value","other":"thing"}`
		entry := &models.CommandHistory{VariablesUsed: &raw}

		vars, err := svc.GetVariables(entry)
		if err != nil {
			t.Fatalf("GetVariables() error = %v", err)
		}
		if len(vars) != 2 || vars["name"] != "value" || vars["other"] != "thing" {
			t.Fatalf("GetVariables() = %v, want map[name:value other:thing]", vars)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		svc, _, _, _ := setupHistoryServiceTest(t)
		raw := `not-json`
		entry := &models.CommandHistory{VariablesUsed: &raw}

		_, err := svc.GetVariables(entry)
		if err == nil {
			t.Fatalf("GetVariables() error = nil, want error")
		}
	})

	t.Run("round-trips values recorded through the repository", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		raw := `{"pkg":"./..."}`
		recorded := mustRecordHistory(t, historyRepo, repository.HistoryEntry{
			Alias:         "build",
			Template:      "go build {{.pkg}}",
			Resolved:      "go build ./...",
			VariablesUsed: &raw,
		})

		vars, err := svc.GetVariables(recorded)
		if err != nil {
			t.Fatalf("GetVariables() error = %v", err)
		}
		if len(vars) != 1 || vars["pkg"] != "./..." {
			t.Fatalf("GetVariables() = %v, want map[pkg:./...]", vars)
		}
	})
}

// --- DeleteByRef ---

func TestHistoryServiceDeleteByRef(t *testing.T) {
	t.Run("deletes by id ref", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		entry := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})

		if err := svc.DeleteByRef(entry.ID, nil, nil); err != nil {
			t.Fatalf("DeleteByRef() error = %v", err)
		}
		if _, err := historyRepo.GetByID(entry.ID, nil); !errors.Is(err, repository.ErrUnknownCommandHistory) {
			t.Fatalf("entry still exists after delete: err = %v", err)
		}
	})

	t.Run("deletes by index ref", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		first := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r1"})
		time.Sleep(time.Millisecond)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r2"})

		// index 2 is the older ("first") entry
		if err := svc.DeleteByRef("2", nil, nil); err != nil {
			t.Fatalf("DeleteByRef() error = %v", err)
		}
		if _, err := historyRepo.GetByID(first.ID, nil); !errors.Is(err, repository.ErrUnknownCommandHistory) {
			t.Fatalf("entry still exists after delete: err = %v", err)
		}
	})

	t.Run("unknown id ref returns error", func(t *testing.T) {
		svc, _, _, _ := setupHistoryServiceTest(t)
		err := svc.DeleteByRef("does-not-exist", nil, nil)
		if !errors.Is(err, repository.ErrUnknownCommandHistory) {
			t.Fatalf("DeleteByRef() error = %v, want ErrUnknownCommandHistory", err)
		}
	})

	t.Run("out of range index returns ErrUnknownHistoryIndex", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})

		err := svc.DeleteByRef("5", nil, nil)
		if !errors.Is(err, ErrUnknownHistoryIndex) {
			t.Fatalf("DeleteByRef() error = %v, want ErrUnknownHistoryIndex", err)
		}
	})

	t.Run("does not delete entry belonging to a different profile", func(t *testing.T) {
		svc, historyRepo, profileRepo, _ := setupHistoryServiceTest(t)
		other := mustCreateHistoryProfile(t, profileRepo, "other")
		entry := mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r", ProfileID: &other.ID})

		err := svc.DeleteByRef(entry.ID, nil, nil)
		if !errors.Is(err, repository.ErrUnknownCommandHistory) {
			t.Fatalf("DeleteByRef() under wrong profile error = %v, want ErrUnknownCommandHistory", err)
		}

		otherName := "other"
		got, gErr := svc.GetByRef(entry.ID, nil, &otherName)
		if gErr != nil {
			t.Fatalf("GetByRef() after failed cross-profile delete error = %v", gErr)
		}
		if got.ID != entry.ID {
			t.Fatalf("entry was unexpectedly removed")
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _ := setupHistoryServiceTest(t)
		missing := "missing"
		err := svc.DeleteByRef("1", nil, &missing)
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("DeleteByRef() error = %v, want ErrProfileNotFound", err)
		}
	})
}

// --- Clear ---

func TestHistoryServiceClear(t *testing.T) {
	t.Run("clears all entries for active profile when alias is nil", func(t *testing.T) {
		svc, historyRepo, _, db := setupHistoryServiceTest(t)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "b", Template: "t", Resolved: "r"})

		n, err := svc.Clear(nil, nil)
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}
		if n != 2 {
			t.Fatalf("Clear() returned %d, want 2", n)
		}

		var count int64
		db.Model(&models.CommandHistory{}).Count(&count)
		if count != 0 {
			t.Fatalf("rows remaining in DB = %d, want 0", count)
		}
	})

	t.Run("clears only the given alias", func(t *testing.T) {
		svc, historyRepo, _, _ := setupHistoryServiceTest(t)
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "keep", Template: "t", Resolved: "r"})
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "clear-me", Template: "t", Resolved: "r"})

		alias := "clear-me"
		n, err := svc.Clear(&alias, nil)
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}
		if n != 1 {
			t.Fatalf("Clear() returned %d, want 1", n)
		}

		remaining, err := svc.GetRecent(nil, nil, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(remaining) != 1 || remaining[0].Alias != "keep" {
			t.Fatalf("remaining = %+v, want only [keep]", remaining)
		}
	})

	t.Run("scopes to named profile", func(t *testing.T) {
		svc, historyRepo, profileRepo, _ := setupHistoryServiceTest(t)
		other := mustCreateHistoryProfile(t, profileRepo, "other")
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r"})
		mustRecordHistory(t, historyRepo, repository.HistoryEntry{Alias: "a", Template: "t", Resolved: "r", ProfileID: &other.ID})

		otherName := "other"
		n, err := svc.Clear(nil, &otherName)
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}
		if n != 1 {
			t.Fatalf("Clear() returned %d, want 1", n)
		}

		defaultEntries, err := svc.GetRecent(nil, nil, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(defaultEntries) != 1 {
			t.Fatalf("len(defaultEntries) = %d, want 1 (Clear on other profile must not affect default)", len(defaultEntries))
		}
	})

	t.Run("returns zero when nothing matches", func(t *testing.T) {
		svc, _, _, _ := setupHistoryServiceTest(t)
		alias := "nonexistent"
		n, err := svc.Clear(&alias, nil)
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}
		if n != 0 {
			t.Fatalf("Clear() returned %d, want 0", n)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		svc, _, _, _ := setupHistoryServiceTest(t)
		missing := "missing"
		_, err := svc.Clear(nil, &missing)
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("Clear() error = %v, want ErrProfileNotFound", err)
		}
	})
}
