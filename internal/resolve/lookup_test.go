package resolve

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

func setupRepoLookupTest(t *testing.T) (repository.CommandRepository, repository.VariableRepository) {
	t.Helper()

	dsn := fmt.Sprintf("file:resolve-lookup-%d?mode=memory&cache=shared", time.Now().UnixNano())
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
	variableRepo := repository.NewVariableRepository(db, profileRepo, validate.NewVariableValidator(nil))
	return commandRepo, variableRepo
}

// --- RepoLookup ---

func TestRepoLookupGetCommandFound(t *testing.T) {
	cmdRepo, varRepo := setupRepoLookupTest(t)
	if _, err := cmdRepo.Create(repository.CommandCreateConfig{Alias: "build", Template: "go build"}); err != nil {
		t.Fatalf("create command: %v", err)
	}

	lookup := NewLookup(cmdRepo, varRepo)
	rec, err := lookup.GetCommand("build")
	if err != nil {
		t.Fatalf("GetCommand() error = %v", err)
	}
	want := CommandRecord{Alias: "build", Template: "go build"}
	if rec != want {
		t.Fatalf("GetCommand() = %+v, want %+v", rec, want)
	}
}

func TestRepoLookupGetCommandNotFound(t *testing.T) {
	cmdRepo, varRepo := setupRepoLookupTest(t)
	lookup := NewLookup(cmdRepo, varRepo)

	_, err := lookup.GetCommand("missing")
	if err == nil {
		t.Fatal("expected error for missing command")
	}
	if !errors.Is(err, repository.ErrUnknownAlias) {
		t.Fatalf("error = %v, want wrapped ErrUnknownAlias", err)
	}
}

func TestRepoLookupGetVariableFound(t *testing.T) {
	cmdRepo, varRepo := setupRepoLookupTest(t)
	if _, err := varRepo.Create(repository.VariableCreateConfig{Name: "greeting", Value: "hello"}); err != nil {
		t.Fatalf("create variable: %v", err)
	}

	lookup := NewLookup(cmdRepo, varRepo)
	rec, err := lookup.GetVariable("greeting")
	if err != nil {
		t.Fatalf("GetVariable() error = %v", err)
	}
	want := VariableRecord{Name: "greeting", Value: "hello"}
	if rec != want {
		t.Fatalf("GetVariable() = %+v, want %+v", rec, want)
	}
}

func TestRepoLookupGetVariableNotFound(t *testing.T) {
	cmdRepo, varRepo := setupRepoLookupTest(t)
	lookup := NewLookup(cmdRepo, varRepo)

	_, err := lookup.GetVariable("missing")
	if err == nil {
		t.Fatal("expected error for missing variable")
	}
	if !errors.Is(err, repository.ErrUnKnownName) {
		t.Fatalf("error = %v, want wrapped ErrUnKnownName", err)
	}
}

// --- MemoizedLookup ---

type countingLookup struct {
	commands     map[string]CommandRecord
	variables    map[string]VariableRecord
	cmdCallCount map[string]int
	varCallCount map[string]int
}

func newCountingLookup() *countingLookup {
	return &countingLookup{
		commands:     make(map[string]CommandRecord),
		variables:    make(map[string]VariableRecord),
		cmdCallCount: make(map[string]int),
		varCallCount: make(map[string]int),
	}
}

func (c *countingLookup) GetCommand(alias string) (CommandRecord, error) {
	c.cmdCallCount[alias]++
	rec, ok := c.commands[alias]
	if !ok {
		return CommandRecord{}, fmt.Errorf("missing command %q: %w", alias, repository.ErrUnknownAlias)
	}
	return rec, nil
}

func (c *countingLookup) GetVariable(name string) (VariableRecord, error) {
	c.varCallCount[name]++
	rec, ok := c.variables[name]
	if !ok {
		return VariableRecord{}, fmt.Errorf("missing variable %q: %w", name, repository.ErrUnKnownName)
	}
	return rec, nil
}

func TestMemoizedLookupCachesCommandAfterFirstCall(t *testing.T) {
	inner := newCountingLookup()
	inner.commands["build"] = CommandRecord{Alias: "build", Template: "go build"}
	lookup := NewMemoizedLookup(inner)

	for i := 0; i < 3; i++ {
		rec, err := lookup.GetCommand("build")
		if err != nil {
			t.Fatalf("GetCommand() error = %v", err)
		}
		if rec != inner.commands["build"] {
			t.Fatalf("GetCommand() = %+v, want %+v", rec, inner.commands["build"])
		}
	}
	if inner.cmdCallCount["build"] != 1 {
		t.Fatalf("inner GetCommand called %d times, want 1", inner.cmdCallCount["build"])
	}
}

func TestMemoizedLookupCachesVariableAfterFirstCall(t *testing.T) {
	inner := newCountingLookup()
	inner.variables["greeting"] = VariableRecord{Name: "greeting", Value: "hello"}
	lookup := NewMemoizedLookup(inner)

	for i := 0; i < 3; i++ {
		rec, err := lookup.GetVariable("greeting")
		if err != nil {
			t.Fatalf("GetVariable() error = %v", err)
		}
		if rec != inner.variables["greeting"] {
			t.Fatalf("GetVariable() = %+v, want %+v", rec, inner.variables["greeting"])
		}
	}
	if inner.varCallCount["greeting"] != 1 {
		t.Fatalf("inner GetVariable called %d times, want 1", inner.varCallCount["greeting"])
	}
}

func TestMemoizedLookupDoesNotCacheErrors(t *testing.T) {
	inner := newCountingLookup()
	lookup := NewMemoizedLookup(inner)

	for i := 0; i < 2; i++ {
		if _, err := lookup.GetCommand("missing"); err == nil {
			t.Fatal("expected error for missing command")
		}
	}
	if inner.cmdCallCount["missing"] != 2 {
		t.Fatalf("inner GetCommand called %d times, want 2 (errors should not be cached)", inner.cmdCallCount["missing"])
	}
}

func TestMemoizedLookupClearForcesRefetch(t *testing.T) {
	inner := newCountingLookup()
	inner.commands["build"] = CommandRecord{Alias: "build", Template: "go build"}
	lookup := NewMemoizedLookup(inner).(*MemoizedLookup)

	if _, err := lookup.GetCommand("build"); err != nil {
		t.Fatalf("GetCommand() error = %v", err)
	}
	lookup.Clear()

	inner.commands["build"] = CommandRecord{Alias: "build", Template: "go build -v"}
	rec, err := lookup.GetCommand("build")
	if err != nil {
		t.Fatalf("GetCommand() error = %v", err)
	}
	if rec.Template != "go build -v" {
		t.Fatalf("GetCommand() = %+v, want updated template after Clear()", rec)
	}
	if inner.cmdCallCount["build"] != 2 {
		t.Fatalf("inner GetCommand called %d times, want 2 after Clear()", inner.cmdCallCount["build"])
	}
}

func TestMemoizedLookupIndependentCommandAndVariableCaches(t *testing.T) {
	inner := newCountingLookup()
	inner.commands["x"] = CommandRecord{Alias: "x", Template: "echo x"}
	inner.variables["x"] = VariableRecord{Name: "x", Value: "val"}
	lookup := NewMemoizedLookup(inner)

	if _, err := lookup.GetCommand("x"); err != nil {
		t.Fatalf("GetCommand() error = %v", err)
	}
	if _, err := lookup.GetVariable("x"); err != nil {
		t.Fatalf("GetVariable() error = %v", err)
	}
	if inner.cmdCallCount["x"] != 1 || inner.varCallCount["x"] != 1 {
		t.Fatalf("call counts = cmd:%d var:%d, want 1 and 1", inner.cmdCallCount["x"], inner.varCallCount["x"])
	}
}
