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

func setupTagRepositoryTest(t *testing.T) (TagRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:tag-repo-%d?mode=memory&cache=shared", time.Now().UnixNano())
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

	if err := db.AutoMigrate(&models.Tag{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	tagRepo := NewTagRepository(db, validate.NewTagValidator(nil))
	return tagRepo, db
}

func mustCreateTag(t *testing.T, repo TagRepository, input TagCreateConfig) *models.Tag {
	t.Helper()
	tag, err := repo.Create(input)
	if err != nil {
		t.Fatalf("Create(%+v) error = %v", input, err)
	}
	return tag
}

func descPtr(s string) *string {
	return &s
}

// --- Create ---

func TestTagRepositoryCreate(t *testing.T) {
	repo, _ := setupTagRepositoryTest(t)

	t.Run("creates tag", func(t *testing.T) {
		tag, err := repo.Create(TagCreateConfig{Name: "urgent"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if tag.ID == 0 {
			t.Fatalf("Create() returned zero ID")
		}
		if tag.Name != "urgent" {
			t.Fatalf("Name = %q, want %q", tag.Name, "urgent")
		}

		fromDB, err := repo.GetByID(tag.ID)
		if err != nil {
			t.Fatalf("GetByID() after Create error = %v", err)
		}
		if fromDB.Name != "urgent" {
			t.Fatalf("persisted name = %q, want %q", fromDB.Name, "urgent")
		}
	})

	t.Run("trims name whitespace", func(t *testing.T) {
		tag, err := repo.Create(TagCreateConfig{Name: "  spaced  "})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if tag.Name != "spaced" {
			t.Fatalf("Name = %q, want %q", tag.Name, "spaced")
		}
	})

	t.Run("stores optional description", func(t *testing.T) {
		tag, err := repo.Create(TagCreateConfig{Name: "described", Description: descPtr("a description")})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if tag.Description == nil || *tag.Description != "a description" {
			t.Fatalf("Description = %v, want %q", tag.Description, "a description")
		}
	})

	t.Run("nil description stays nil", func(t *testing.T) {
		tag, err := repo.Create(TagCreateConfig{Name: "no-desc"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if tag.Description != nil {
			t.Fatalf("Description = %v, want nil", *tag.Description)
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := repo.Create(TagCreateConfig{Name: ""})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects whitespace-only name", func(t *testing.T) {
		_, err := repo.Create(TagCreateConfig{Name: "   "})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects name with spaces", func(t *testing.T) {
		_, err := repo.Create(TagCreateConfig{Name: "two words"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects reserved name", func(t *testing.T) {
		_, err := repo.Create(TagCreateConfig{Name: "help"})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects name over max length", func(t *testing.T) {
		_, err := repo.Create(TagCreateConfig{Name: fmt.Sprintf("%0101d", 0)})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Create() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		if _, err := repo.Create(TagCreateConfig{Name: "dup"}); err != nil {
			t.Fatalf("Create() first error = %v", err)
		}
		_, err := repo.Create(TagCreateConfig{Name: "dup"})
		if !errors.Is(err, ErrTagNameConflict) {
			t.Fatalf("Create() second error = %v, want ErrTagNameConflict", err)
		}
	})
}

// --- GetByName ---

func TestTagRepositoryGetByName(t *testing.T) {
	repo, _ := setupTagRepositoryTest(t)
	tag := mustCreateTag(t, repo, TagCreateConfig{Name: "urgent"})

	t.Run("returns tag by name", func(t *testing.T) {
		got, err := repo.GetByName("urgent")
		if err != nil {
			t.Fatalf("GetByName() error = %v", err)
		}
		if got.ID != tag.ID {
			t.Fatalf("GetByName() id = %d, want %d", got.ID, tag.ID)
		}
	})

	t.Run("lookup is case-sensitive", func(t *testing.T) {
		_, err := repo.GetByName("URGENT")
		if !errors.Is(err, ErrUnknownTagName) {
			t.Fatalf("GetByName(\"URGENT\") error = %v, want ErrUnknownTagName", err)
		}
	})

	t.Run("returns ErrUnknownTagName when missing", func(t *testing.T) {
		_, err := repo.GetByName("missing")
		if !errors.Is(err, ErrUnknownTagName) {
			t.Fatalf("GetByName() error = %v, want ErrUnknownTagName", err)
		}
	})
}

// --- GetByID ---

func TestTagRepositoryGetByID(t *testing.T) {
	repo, _ := setupTagRepositoryTest(t)
	tag := mustCreateTag(t, repo, TagCreateConfig{Name: "urgent"})

	t.Run("returns tag by id", func(t *testing.T) {
		got, err := repo.GetByID(tag.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got.Name != "urgent" {
			t.Fatalf("Name = %q, want %q", got.Name, "urgent")
		}
	})

	t.Run("returns ErrUnknownTag when missing", func(t *testing.T) {
		_, err := repo.GetByID(tag.ID + 999999)
		if !errors.Is(err, ErrUnknownTag) {
			t.Fatalf("GetByID() error = %v, want ErrUnknownTag", err)
		}
	})
}

// --- Update ---

func TestTagRepositoryUpdateValidation(t *testing.T) {
	repo, _ := setupTagRepositoryTest(t)
	tag := mustCreateTag(t, repo, TagCreateConfig{Name: "urgent"})

	t.Run("nil tag returns ErrNoUpdateTarget", func(t *testing.T) {
		_, err := repo.Update(nil, TagUpdateConfig{})
		if !errors.Is(err, ErrNoUpdateTarget) {
			t.Fatalf("Update(nil) error = %v, want ErrNoUpdateTarget", err)
		}
	})

	t.Run("rejects reserved name on rename", func(t *testing.T) {
		fresh := *tag
		newName := "help"
		_, err := repo.Update(&fresh, TagUpdateConfig{Name: &newName})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Update() error = %v, want ErrValidation", err)
		}
	})

	t.Run("rejects rename to name with spaces", func(t *testing.T) {
		fresh := *tag
		newName := "two words"
		_, err := repo.Update(&fresh, TagUpdateConfig{Name: &newName})
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("Update() error = %v, want ErrValidation", err)
		}
	})
}

func TestTagRepositoryUpdateFields(t *testing.T) {
	repo, _ := setupTagRepositoryTest(t)

	t.Run("updates name", func(t *testing.T) {
		tag := mustCreateTag(t, repo, TagCreateConfig{Name: "rename-me"})
		newName := "renamed"
		updated, err := repo.Update(tag, TagUpdateConfig{Name: &newName})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Name != "renamed" {
			t.Fatalf("Name = %q, want %q", updated.Name, "renamed")
		}

		persisted, err := repo.GetByID(tag.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if persisted.Name != "renamed" {
			t.Fatalf("persisted name = %q, want %q", persisted.Name, "renamed")
		}
	})

	t.Run("trims renamed name whitespace", func(t *testing.T) {
		tag := mustCreateTag(t, repo, TagCreateConfig{Name: "trim-me"})
		newName := "  trimmed  "
		updated, err := repo.Update(tag, TagUpdateConfig{Name: &newName})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Name != "trimmed" {
			t.Fatalf("Name = %q, want %q", updated.Name, "trimmed")
		}
	})

	t.Run("updates description", func(t *testing.T) {
		tag := mustCreateTag(t, repo, TagCreateConfig{Name: "desc-me"})
		updated, err := repo.Update(tag, TagUpdateConfig{Description: descPtr("new description")})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Description == nil || *updated.Description != "new description" {
			t.Fatalf("Description = %v, want %q", updated.Description, "new description")
		}
	})

	t.Run("no-op update leaves fields unchanged", func(t *testing.T) {
		tag := mustCreateTag(t, repo, TagCreateConfig{Name: "noop-me", Description: descPtr("kept")})
		updated, err := repo.Update(tag, TagUpdateConfig{})
		if err != nil {
			t.Fatalf("Update(no-op) error = %v", err)
		}
		if updated.Name != "noop-me" {
			t.Fatalf("Update(no-op) changed name: %+v", updated)
		}
		if updated.Description == nil || *updated.Description != "kept" {
			t.Fatalf("Update(no-op) changed description: %v", updated.Description)
		}
	})

	t.Run("rejects rename to name already used", func(t *testing.T) {
		_ = mustCreateTag(t, repo, TagCreateConfig{Name: "taken"})
		tag := mustCreateTag(t, repo, TagCreateConfig{Name: "renaming"})

		newName := "taken"
		_, err := repo.Update(tag, TagUpdateConfig{Name: &newName})
		if !errors.Is(err, ErrTagNameConflict) {
			t.Fatalf("Update() error = %v, want ErrTagNameConflict", err)
		}
	})
}

// --- Delete ---

func TestTagRepositoryDelete(t *testing.T) {
	repo, _ := setupTagRepositoryTest(t)

	t.Run("nil tag is a no-op", func(t *testing.T) {
		if err := repo.Delete(nil); err != nil {
			t.Fatalf("Delete(nil) error = %v", err)
		}
	})

	t.Run("deletes existing tag", func(t *testing.T) {
		tag := mustCreateTag(t, repo, TagCreateConfig{Name: "delete-me"})
		if err := repo.Delete(tag); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		_, err := repo.GetByID(tag.ID)
		if !errors.Is(err, ErrUnknownTag) {
			t.Fatalf("GetByID() after delete error = %v, want ErrUnknownTag", err)
		}
	})
}

// --- ListAll ---

func TestTagRepositoryListAll(t *testing.T) {
	repo, _ := setupTagRepositoryTest(t)
	_ = mustCreateTag(t, repo, TagCreateConfig{Name: "zeta"})
	_ = mustCreateTag(t, repo, TagCreateConfig{Name: "alpha"})
	_ = mustCreateTag(t, repo, TagCreateConfig{Name: "mid"})

	t.Run("defaults to name order and limit 25", func(t *testing.T) {
		tags, err := repo.ListAll("", 0)
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(tags) != 3 {
			t.Fatalf("len(tags) = %d, want 3", len(tags))
		}
		wantOrder := []string{"alpha", "mid", "zeta"}
		for i, want := range wantOrder {
			if tags[i].Name != want {
				t.Fatalf("tags[%d].Name = %q, want %q", i, tags[i].Name, want)
			}
		}
	})

	t.Run("respects explicit order and limit", func(t *testing.T) {
		tags, err := repo.ListAll("-name", 2)
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if len(tags) != 2 {
			t.Fatalf("len(tags) = %d, want 2", len(tags))
		}
		if tags[0].Name != "zeta" || tags[1].Name != "mid" {
			t.Fatalf("tags = %+v, want [zeta, mid]", tags)
		}
	})

	t.Run("invalid order field returns error", func(t *testing.T) {
		_, err := repo.ListAll("bogus", 0)
		if err == nil {
			t.Fatalf("ListAll() error = nil, want error")
		}
	})
}

func TestTagRepositoryListAllEmptyTable(t *testing.T) {
	repo, _ := setupTagRepositoryTest(t)
	tags, err := repo.ListAll("", 0)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("tags = %+v, want empty", tags)
	}
}

// --- Search ---

func TestTagRepositorySearch(t *testing.T) {
	repo, _ := setupTagRepositoryTest(t)
	_ = mustCreateTag(t, repo, TagCreateConfig{Name: "apple", Description: descPtr("a fruit")})
	_ = mustCreateTag(t, repo, TagCreateConfig{Name: "green-apple", Description: descPtr("a green fruit")})
	_ = mustCreateTag(t, repo, TagCreateConfig{Name: "banana", Description: descPtr("a yellow fruit")})

	t.Run("uses default fields when none provided", func(t *testing.T) {
		tags, err := repo.Search("apple", nil, 0)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(tags) != 2 {
			t.Fatalf("len(tags) = %d, want 2; tags = %+v", len(tags), tags)
		}
		if tags[0].Name != "apple" {
			t.Fatalf("tags[0].Name = %q, want %q (closest match first)", tags[0].Name, "apple")
		}
	})

	t.Run("matches against description field by default", func(t *testing.T) {
		tags, err := repo.Search("yellow", nil, 0)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(tags) != 1 || tags[0].Name != "banana" {
			t.Fatalf("tags = %+v, want only [banana]", tags)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		tags, err := repo.Search("apple", nil, 1)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(tags) != 1 {
			t.Fatalf("len(tags) = %d, want 1", len(tags))
		}
	})

	t.Run("invalid explicit field returns error", func(t *testing.T) {
		_, err := repo.Search("apple", []string{"bogus"}, 0)
		if err == nil {
			t.Fatalf("Search() error = nil, want error")
		}
	})

	t.Run("no matches returns empty slice", func(t *testing.T) {
		tags, err := repo.Search("nonexistent", nil, 0)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(tags) != 0 {
			t.Fatalf("tags = %+v, want empty", tags)
		}
	})
}

// --- misc ---

func TestNewTagRepositoryDefaultValidator(t *testing.T) {
	_, db := setupTagRepositoryTest(t)
	repo := NewTagRepository(db, nil)

	_, err := repo.Create(TagCreateConfig{Name: "help"})
	if !errors.Is(err, validate.ErrValidation) {
		t.Fatalf("Create() with default validator error = %v, want ErrValidation (reserved name)", err)
	}
}
