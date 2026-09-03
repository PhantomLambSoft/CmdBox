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

func setupTagServiceTest(t *testing.T) (TagService, repository.TagRepository, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:tag-service-%d?mode=memory&cache=shared", time.Now().UnixNano())
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

	if err := db.AutoMigrate(&models.Tag{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	tagRepo := repository.NewTagRepository(db, validate.NewTagValidator(nil))
	svc := NewTagService(tagRepo)
	return svc, tagRepo, db
}

func mustCreateTag(t *testing.T, svc TagService, name string, description *string) *models.Tag {
	t.Helper()
	tag, err := svc.CreateTag(name, description)
	if err != nil {
		t.Fatalf("CreateTag(%q) error = %v", name, err)
	}
	return tag
}

// --- CreateTag ---

func TestTagServiceCreateTag(t *testing.T) {
	t.Run("creates tag", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		tag, err := svc.CreateTag("urgent", nil)
		if err != nil {
			t.Fatalf("CreateTag() error = %v", err)
		}
		if tag.ID == 0 {
			t.Fatalf("CreateTag() returned zero ID")
		}
		if tag.Name != "urgent" {
			t.Fatalf("Name = %q, want %q", tag.Name, "urgent")
		}
	})

	t.Run("creates tag with description", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		description := "high priority items"
		tag, err := svc.CreateTag("urgent", &description)
		if err != nil {
			t.Fatalf("CreateTag() error = %v", err)
		}
		if tag.Description == nil || *tag.Description != description {
			t.Fatalf("Description = %v, want %q", tag.Description, description)
		}
	})

	t.Run("duplicate name returns error", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		mustCreateTag(t, svc, "dup", nil)

		_, err := svc.CreateTag("dup", nil)
		if !errors.Is(err, repository.ErrTagNameConflict) {
			t.Fatalf("CreateTag() error = %v, want ErrTagNameConflict", err)
		}
	})

	t.Run("invalid name returns validation error", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		_, err := svc.CreateTag("", nil)
		if !errors.Is(err, validate.ErrValidation) {
			t.Fatalf("CreateTag() error = %v, want ErrValidation", err)
		}
	})
}

// --- UpdateTag ---

func TestTagServiceUpdateTag(t *testing.T) {
	t.Run("updates name", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		mustCreateTag(t, svc, "rename-me", nil)

		newName := "renamed"
		updated, err := svc.UpdateTag("rename-me", &UpdateTagConfig{NewName: &newName})
		if err != nil {
			t.Fatalf("UpdateTag() error = %v", err)
		}
		if updated.Name != "renamed" {
			t.Fatalf("Name = %q, want %q", updated.Name, "renamed")
		}
	})

	t.Run("updates description", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		mustCreateTag(t, svc, "described", nil)

		newDescription := "a new description"
		updated, err := svc.UpdateTag("described", &UpdateTagConfig{Description: &newDescription})
		if err != nil {
			t.Fatalf("UpdateTag() error = %v", err)
		}
		if updated.Description == nil || *updated.Description != newDescription {
			t.Fatalf("Description = %v, want %q", updated.Description, newDescription)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		_, err := svc.UpdateTag("missing", &UpdateTagConfig{})
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("UpdateTag() error = %v, want ErrUnknownTagName", err)
		}
	})

	t.Run("name conflict returns error", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		mustCreateTag(t, svc, "taken", nil)
		mustCreateTag(t, svc, "renaming", nil)

		newName := "taken"
		_, err := svc.UpdateTag("renaming", &UpdateTagConfig{NewName: &newName})
		if !errors.Is(err, repository.ErrTagNameConflict) {
			t.Fatalf("UpdateTag() error = %v, want ErrTagNameConflict", err)
		}
	})
}

// --- DeleteTag ---

func TestTagServiceDeleteTag(t *testing.T) {
	t.Run("deletes existing tag", func(t *testing.T) {
		svc, tagRepo, _ := setupTagServiceTest(t)
		mustCreateTag(t, svc, "delete-me", nil)

		if err := svc.DeleteTag("delete-me"); err != nil {
			t.Fatalf("DeleteTag() error = %v", err)
		}
		if _, err := tagRepo.GetByName("delete-me"); !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("tag still exists after delete: err = %v", err)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		err := svc.DeleteTag("missing")
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("DeleteTag() error = %v, want ErrUnknownTagName", err)
		}
	})
}

// --- GetTag / GetTagOrNil / GetTagById ---

func TestTagServiceGetTag(t *testing.T) {
	t.Run("returns tag when found", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		created := mustCreateTag(t, svc, "urgent", nil)

		got, err := svc.GetTag("urgent")
		if err != nil {
			t.Fatalf("GetTag() error = %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("GetTag() id = %d, want %d", got.ID, created.ID)
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		_, err := svc.GetTag("missing")
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("GetTag() error = %v, want ErrUnknownTagName", err)
		}
	})
}

func TestTagServiceGetTagOrNil(t *testing.T) {
	t.Run("returns nil, nil when name missing", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		tag, err := svc.GetTagOrNil("missing")
		if err != nil {
			t.Fatalf("GetTagOrNil() error = %v, want nil", err)
		}
		if tag != nil {
			t.Fatalf("GetTagOrNil() = %v, want nil", tag)
		}
	})

	t.Run("returns tag when found", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		created := mustCreateTag(t, svc, "urgent", nil)

		tag, err := svc.GetTagOrNil("urgent")
		if err != nil {
			t.Fatalf("GetTagOrNil() error = %v", err)
		}
		if tag == nil || tag.ID != created.ID {
			t.Fatalf("GetTagOrNil() = %v, want id %d", tag, created.ID)
		}
	})
}

func TestTagServiceGetTagById(t *testing.T) {
	t.Run("returns tag when found", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		created := mustCreateTag(t, svc, "urgent", nil)

		got, err := svc.GetTagById(created.ID)
		if err != nil {
			t.Fatalf("GetTagById() error = %v", err)
		}
		if got.Name != "urgent" {
			t.Fatalf("Name = %q, want %q", got.Name, "urgent")
		}
	})

	t.Run("unknown id returns error", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		_, err := svc.GetTagById(999999)
		if !errors.Is(err, repository.ErrUnknownTag) {
			t.Fatalf("GetTagById() error = %v, want ErrUnknownTag", err)
		}
	})
}

// --- ListTags ---

func TestTagServiceListTags(t *testing.T) {
	t.Run("lists all tags with default ordering", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		mustCreateTag(t, svc, "zeta", nil)
		mustCreateTag(t, svc, "alpha", nil)

		tags, err := svc.ListTags("", nil)
		if err != nil {
			t.Fatalf("ListTags() error = %v", err)
		}
		if len(tags) != 2 {
			t.Fatalf("len(tags) = %d, want 2", len(tags))
		}
		if tags[0].Name != "alpha" {
			t.Fatalf("tags[0].Name = %q, want %q", tags[0].Name, "alpha")
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		mustCreateTag(t, svc, "a", nil)
		mustCreateTag(t, svc, "b", nil)

		limit := 1
		tags, err := svc.ListTags("", &limit)
		if err != nil {
			t.Fatalf("ListTags() error = %v", err)
		}
		if len(tags) != 1 {
			t.Fatalf("len(tags) = %d, want 1", len(tags))
		}
	})

	t.Run("invalid order field returns error", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		_, err := svc.ListTags("bogus", nil)
		if err == nil {
			t.Fatalf("ListTags() error = nil, want error")
		}
	})
}

// --- Search ---

func TestTagServiceSearch(t *testing.T) {
	t.Run("uses default fields", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		mustCreateTag(t, svc, "apple", nil)
		mustCreateTag(t, svc, "banana", nil)

		tags, err := svc.Search("apple", nil, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(tags) != 1 || tags[0].Name != "apple" {
			t.Fatalf("tags = %+v, want only [apple]", tags)
		}
	})

	t.Run("respects explicit fields", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		fruitDescription := "a fruit"
		mustCreateTag(t, svc, "apple", &fruitDescription)
		mustCreateTag(t, svc, "unrelated", &fruitDescription)

		tags, err := svc.Search("apple", []string{"name"}, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(tags) != 1 || tags[0].Name != "apple" {
			t.Fatalf("tags = %+v, want only [apple]", tags)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		mustCreateTag(t, svc, "apple", nil)
		mustCreateTag(t, svc, "apple-two", nil)

		limit := 1
		tags, err := svc.Search("apple", nil, &limit)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(tags) != 1 {
			t.Fatalf("len(tags) = %d, want 1", len(tags))
		}
	})

	t.Run("invalid explicit field returns error", func(t *testing.T) {
		svc, _, _ := setupTagServiceTest(t)
		_, err := svc.Search("apple", []string{"bogus"}, nil)
		if err == nil {
			t.Fatalf("Search() error = nil, want error")
		}
	})
}
