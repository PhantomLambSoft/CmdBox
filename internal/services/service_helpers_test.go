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

func setupTagRepoForHelperTest(t *testing.T) repository.TagRepository {
	t.Helper()

	dsn := fmt.Sprintf("file:service-helpers-%d?mode=memory&cache=shared", time.Now().UnixNano())
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

	return repository.NewTagRepository(db, validate.NewTagValidator(nil))
}

func mustCreateHelperTag(t *testing.T, tagRepo repository.TagRepository, name string) *models.Tag {
	t.Helper()
	tag, err := tagRepo.Create(repository.TagCreateConfig{Name: name})
	if err != nil {
		t.Fatalf("create tag %q: %v", name, err)
	}
	return tag
}

// --- getTags ---

func TestGetTags(t *testing.T) {
	t.Run("nil names returns nil, nil", func(t *testing.T) {
		tagRepo := setupTagRepoForHelperTest(t)
		tags, err := getTags(nil, tagRepo)
		if err != nil {
			t.Fatalf("getTags(nil) error = %v", err)
		}
		if tags != nil {
			t.Fatalf("getTags(nil) = %v, want nil", tags)
		}
	})

	t.Run("empty slice returns nil, nil", func(t *testing.T) {
		tagRepo := setupTagRepoForHelperTest(t)
		tags, err := getTags([]string{}, tagRepo)
		if err != nil {
			t.Fatalf("getTags([]) error = %v", err)
		}
		if tags != nil {
			t.Fatalf("getTags([]) = %v, want nil", tags)
		}
	})

	t.Run("resolves a single tag by name", func(t *testing.T) {
		tagRepo := setupTagRepoForHelperTest(t)
		created := mustCreateHelperTag(t, tagRepo, "alpha")

		tags, err := getTags([]string{"alpha"}, tagRepo)
		if err != nil {
			t.Fatalf("getTags() error = %v", err)
		}
		if len(tags) != 1 {
			t.Fatalf("len(tags) = %d, want 1", len(tags))
		}
		if tags[0].ID != created.ID {
			t.Fatalf("tags[0].ID = %d, want %d", tags[0].ID, created.ID)
		}
	})

	t.Run("resolves multiple tags in the given order", func(t *testing.T) {
		tagRepo := setupTagRepoForHelperTest(t)
		mustCreateHelperTag(t, tagRepo, "alpha")
		mustCreateHelperTag(t, tagRepo, "beta")
		mustCreateHelperTag(t, tagRepo, "gamma")

		tags, err := getTags([]string{"gamma", "alpha", "beta"}, tagRepo)
		if err != nil {
			t.Fatalf("getTags() error = %v", err)
		}
		if len(tags) != 3 {
			t.Fatalf("len(tags) = %d, want 3", len(tags))
		}
		wantOrder := []string{"gamma", "alpha", "beta"}
		for i, want := range wantOrder {
			if tags[i].Name != want {
				t.Fatalf("tags[%d].Name = %q, want %q", i, tags[i].Name, want)
			}
		}
	})

	t.Run("unknown tag name returns error", func(t *testing.T) {
		tagRepo := setupTagRepoForHelperTest(t)
		_, err := getTags([]string{"missing"}, tagRepo)
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("getTags() error = %v, want ErrUnknownTagName", err)
		}
	})

	t.Run("stops at the first unknown tag without resolving the rest", func(t *testing.T) {
		tagRepo := setupTagRepoForHelperTest(t)
		mustCreateHelperTag(t, tagRepo, "alpha")
		mustCreateHelperTag(t, tagRepo, "gamma")

		_, err := getTags([]string{"alpha", "missing", "gamma"}, tagRepo)
		if !errors.Is(err, repository.ErrUnknownTagName) {
			t.Fatalf("getTags() error = %v, want ErrUnknownTagName", err)
		}
	})

	t.Run("duplicate names in input resolve to duplicate tags in output", func(t *testing.T) {
		tagRepo := setupTagRepoForHelperTest(t)
		created := mustCreateHelperTag(t, tagRepo, "alpha")

		tags, err := getTags([]string{"alpha", "alpha"}, tagRepo)
		if err != nil {
			t.Fatalf("getTags() error = %v", err)
		}
		if len(tags) != 2 {
			t.Fatalf("len(tags) = %d, want 2", len(tags))
		}
		if tags[0].ID != created.ID || tags[1].ID != created.ID {
			t.Fatalf("tags = %+v, want both id %d", tags, created.ID)
		}
	})
}
