package settings

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

// fakeProfileService is a minimal stand-in for services.ProfileService that only
// needs to support GetActiveSettingsProfile for the purposes of these tests.
type fakeProfileService struct {
	activeSettingsProfile *models.Profile
	activeSettingsErr     error
}

var errNotImplemented = errors.New("not implemented")

func (f *fakeProfileService) CreateProfile(string, *string) (*models.Profile, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) UpdateProfile(string, services.UpdateProfileConfig) (*models.Profile, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) DeleteProfile(string, bool) error { return errNotImplemented }
func (f *fakeProfileService) GetProfile(string) (*models.Profile, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) ListProfiles(string, *int) ([]models.Profile, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) SearchProfiles(string, []string, *int) ([]models.Profile, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) GetActiveCommandProfile() (*models.Profile, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) GetActiveVariableProfile() (*models.Profile, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) GetActiveSettingsProfile() (*models.Profile, error) {
	return f.activeSettingsProfile, f.activeSettingsErr
}
func (f *fakeProfileService) SwitchProfile(string) (*models.ProfileState, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) SwitchCommandProfile(string) (*models.ProfileState, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) SwitchVariableProfile(string) (*models.ProfileState, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) SwitchSettingsProfile(string) (*models.ProfileState, error) {
	return nil, errNotImplemented
}
func (f *fakeProfileService) GetStatus() (services.ProfileStatus, error) {
	return services.ProfileStatus{}, errNotImplemented
}

var _ services.ProfileService = (*fakeProfileService)(nil)

func newTestService(t *testing.T, path string, defaults *Settings, profileSvc services.ProfileService) *Service {
	t.Helper()
	repo := NewSettingsRepository(path)
	svc, err := NewService(repo, defaults, profileSvc)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return svc
}

// --- NewService / Get ---

func TestNewService(t *testing.T) {
	t.Run("uses package defaults when none provided", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		if got := svc.Get(); !reflect.DeepEqual(got, DefaultSettings()) {
			t.Fatalf("Get() = %+v, want %+v", got, DefaultSettings())
		}
	})

	t.Run("uses provided defaults", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		custom := DefaultSettings()
		custom.History.LimitPerCommand = 999

		svc := newTestService(t, path, &custom, nil)

		if got := svc.Get(); got.History.LimitPerCommand != 999 {
			t.Fatalf("Get().History.LimitPerCommand = %d, want 999", got.History.LimitPerCommand)
		}
	})

	t.Run("loads existing settings file over defaults", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		if err := writeFileForTest(t, path, "[history]\nlimit_per_command = 13\n"); err != nil {
			t.Fatalf("writing fixture: %v", err)
		}

		svc := newTestService(t, path, nil, nil)

		got := svc.Get()
		if got.History.LimitPerCommand != 13 {
			t.Fatalf("Get().History.LimitPerCommand = %d, want 13", got.History.LimitPerCommand)
		}
		if got.UI.PagerMode != "auto" {
			t.Fatalf("Get().UI.PagerMode = %q, want default %q", got.UI.PagerMode, "auto")
		}
	})

	t.Run("returns error for malformed settings file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		if err := writeFileForTest(t, path, "not valid toml [[["); err != nil {
			t.Fatalf("writing fixture: %v", err)
		}

		repo := NewSettingsRepository(path)
		_, err := NewService(repo, nil, nil)
		if err == nil {
			t.Fatalf("NewService() error = nil, want error for malformed settings file")
		}
	})
}

// --- Reload ---

func TestServiceReload(t *testing.T) {
	t.Run("picks up changes written to the file externally", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		if err := writeFileForTest(t, path, "[history]\nlimit_per_command = 3\n"); err != nil {
			t.Fatalf("writing fixture: %v", err)
		}

		got, err := svc.Reload()
		if err != nil {
			t.Fatalf("Reload() error = %v", err)
		}
		if got.History.LimitPerCommand != 3 {
			t.Fatalf("Reload().History.LimitPerCommand = %d, want 3", got.History.LimitPerCommand)
		}
		if svc.Get().History.LimitPerCommand != 3 {
			t.Fatalf("Get() after Reload() = %d, want 3", svc.Get().History.LimitPerCommand)
		}
	})

	t.Run("returns error when the file becomes malformed", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		if err := writeFileForTest(t, path, "not valid toml [[["); err != nil {
			t.Fatalf("writing fixture: %v", err)
		}

		if _, err := svc.Reload(); err == nil {
			t.Fatalf("Reload() error = nil, want error")
		}
	})
}

// --- Update ---

func TestServiceUpdate(t *testing.T) {
	t.Run("applies the mutation and persists it", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		got, err := svc.Update(func(s *Settings) {
			s.History.LimitPerCommand = 77
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if got.History.LimitPerCommand != 77 {
			t.Fatalf("Update() returned %+v, want LimitPerCommand=77", got)
		}
		if svc.Get().History.LimitPerCommand != 77 {
			t.Fatalf("Get() after Update() = %d, want 77", svc.Get().History.LimitPerCommand)
		}

		// Confirm it was actually persisted to disk, not just held in memory.
		fresh := newTestService(t, path, nil, nil)
		if fresh.Get().History.LimitPerCommand != 77 {
			t.Fatalf("reloaded service History.LimitPerCommand = %d, want 77", fresh.Get().History.LimitPerCommand)
		}
	})

	t.Run("mutation is applied on top of current settings, not defaults", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		if _, err := svc.Update(func(s *Settings) { s.History.LimitPerCommand = 5 }); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		got, err := svc.Update(func(s *Settings) { s.UI.UseColor = false })
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if got.History.LimitPerCommand != 5 {
			t.Fatalf("Update() lost prior change: LimitPerCommand = %d, want 5", got.History.LimitPerCommand)
		}
		if got.UI.UseColor != false {
			t.Fatalf("Update() UseColor = %v, want false", got.UI.UseColor)
		}
	})

	t.Run("returns error when save fails", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		// Turn the (still-nonexistent) settings path into a directory so the
		// atomic rename-over-destination in Save fails.
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatalf("creating blocking directory: %v", err)
		}

		_, err := svc.Update(func(s *Settings) { s.History.LimitPerCommand = 1 })
		if err == nil {
			t.Fatalf("Update() error = nil, want error when destination is a directory")
		}
	})
}

// --- Edit ---

func TestServiceEdit(t *testing.T) {
	t.Run("no-op edit returns current settings without saving", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		got, err := svc.Edit(func(current string) (string, error) {
			return current, nil
		})
		if err != nil {
			t.Fatalf("Edit() error = %v", err)
		}
		if !reflect.DeepEqual(got, svc.Get()) {
			t.Fatalf("Edit() = %+v, want unchanged current settings %+v", got, svc.Get())
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("Edit() with no-op change should not create a settings file, stat err = %v", err)
		}
	})

	t.Run("valid edit is parsed and persisted", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		got, err := svc.Edit(func(current string) (string, error) {
			if !strings.Contains(current, "limit_per_command = 100") {
				t.Fatalf("current TOML = %q, want it to contain the default limit_per_command", current)
			}
			return strings.Replace(current, "limit_per_command = 100", "limit_per_command = 88", 1), nil
		})
		if err != nil {
			t.Fatalf("Edit() error = %v", err)
		}
		if got.History.LimitPerCommand != 88 {
			t.Fatalf("Edit() = %+v, want LimitPerCommand=88", got)
		}

		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected settings file to be written: %v", err)
		}
	})

	t.Run("invalid TOML returns error and does not persist", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		_, err := svc.Edit(func(current string) (string, error) {
			return "not valid toml [[[", nil
		})
		if err == nil {
			t.Fatalf("Edit() error = nil, want error for invalid TOML")
		}
		if !strings.Contains(err.Error(), "invalid TOML") {
			t.Fatalf("Edit() error = %v, want message containing %q", err, "invalid TOML")
		}
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("Edit() with invalid TOML should not create a settings file")
		}
	})

	t.Run("editFn error is propagated without saving", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		editErr := errors.New("editor failed")
		_, err := svc.Edit(func(current string) (string, error) {
			return "", editErr
		})
		if !errors.Is(err, editErr) {
			t.Fatalf("Edit() error = %v, want %v", err, editErr)
		}
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("Edit() with editFn error should not create a settings file")
		}
	})
}

// --- ResolveSettingsPath ---

func TestServiceResolveSettingsPath(t *testing.T) {
	t.Run("explicit default name resolves to config.toml", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		name := "default"
		got, err := svc.ResolveSettingsPath("/appdata", &name)
		if err != nil {
			t.Fatalf("ResolveSettingsPath() error = %v", err)
		}
		want := filepath.Join("/appdata", "config.toml")
		if got != want {
			t.Fatalf("ResolveSettingsPath() = %q, want %q", got, want)
		}
	})

	t.Run("explicit non-default name resolves to a named config file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		svc := newTestService(t, path, nil, nil)

		name := "work"
		got, err := svc.ResolveSettingsPath("/appdata", &name)
		if err != nil {
			t.Fatalf("ResolveSettingsPath() error = %v", err)
		}
		want := filepath.Join("/appdata", "work_config.toml")
		if got != want {
			t.Fatalf("ResolveSettingsPath() = %q, want %q", got, want)
		}
	})

	t.Run("explicit name does not consult the profile service", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		fake := &fakeProfileService{activeSettingsErr: errors.New("should not be called")}
		svc := newTestService(t, path, nil, fake)

		name := "work"
		if _, err := svc.ResolveSettingsPath("/appdata", &name); err != nil {
			t.Fatalf("ResolveSettingsPath() error = %v, want nil (profile service should not be consulted)", err)
		}
	})

	t.Run("nil name resolves via active settings profile - default profile", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		fake := &fakeProfileService{activeSettingsProfile: &models.Profile{Name: "default"}}
		svc := newTestService(t, path, nil, fake)

		got, err := svc.ResolveSettingsPath("/appdata", nil)
		if err != nil {
			t.Fatalf("ResolveSettingsPath() error = %v", err)
		}
		want := filepath.Join("/appdata", "config.toml")
		if got != want {
			t.Fatalf("ResolveSettingsPath() = %q, want %q", got, want)
		}
	})

	t.Run("nil name resolves via active settings profile - named profile", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		fake := &fakeProfileService{activeSettingsProfile: &models.Profile{Name: "personal"}}
		svc := newTestService(t, path, nil, fake)

		got, err := svc.ResolveSettingsPath("/appdata", nil)
		if err != nil {
			t.Fatalf("ResolveSettingsPath() error = %v", err)
		}
		want := filepath.Join("/appdata", "personal_config.toml")
		if got != want {
			t.Fatalf("ResolveSettingsPath() = %q, want %q", got, want)
		}
	})

	t.Run("nil name propagates profile service error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		underlying := errors.New("profile lookup failed")
		fake := &fakeProfileService{activeSettingsErr: underlying}
		svc := newTestService(t, path, nil, fake)

		_, err := svc.ResolveSettingsPath("/appdata", nil)
		if !errors.Is(err, underlying) {
			t.Fatalf("ResolveSettingsPath() error = %v, want wrapped %v", err, underlying)
		}
	})
}
