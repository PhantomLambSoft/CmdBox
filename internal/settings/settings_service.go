package settings

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/services"
	"github.com/pelletier/go-toml/v2"
)

type Service struct {
	repo           *Repository
	defaults       Settings
	current        Settings
	profileService services.ProfileService
}

func NewService(repo *Repository, defaults *Settings, profileService services.ProfileService) (*Service, error) {
	d := DefaultSettings()
	if defaults != nil {
		d = *defaults
	}
	s := &Service{repo: repo, defaults: d, profileService: profileService}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Service) load() error {
	current := s.defaults
	if err := s.repo.Load(&current); err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}
	s.current = current
	return nil
}

func (s *Service) Get() Settings {
	return s.current
}

// Reload reloads the current settings by reading from the repository and applies the default settings if necessary.
func (s *Service) Reload() (Settings, error) {
	if err := s.load(); err != nil {
		return Settings{}, err
	}
	return s.current, nil
}

// Update applies the given mutation function to the current settings and saves the updated settings to the repository.
func (s *Service) Update(mutate func(*Settings)) (Settings, error) {
	updated := s.current
	mutate(&updated)
	if err := s.repo.Save(updated); err != nil {
		return Settings{}, err
	}
	return s.Reload()
}

// Edit applies a user-defined function to modify the current settings and saves the updated settings to the repository.
// The entire settings file is passed through the edit function (ie an external editor or the interactive terminal editor).
func (s *Service) Edit(editFn func(current string) (string, error)) (Settings, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).EnableMarshalerInterface().Encode(s.current); err != nil {
		return Settings{}, err
	}
	text := buf.String()

	edited, err := editFn(text)
	if err != nil {
		return Settings{}, err
	}
	if edited == text {
		return s.current, nil
	}

	candidate := s.defaults
	if err := toml.NewDecoder(strings.NewReader(edited)).EnableUnmarshalerInterface().Decode(&candidate); err != nil {
		return Settings{}, fmt.Errorf("invalid TOML: %w", err)
	}

	if err := s.repo.Save(candidate); err != nil {
		return Settings{}, err
	}
	return s.Reload()
}

// ResolveSettingsPath determines the file path to the settings configuration based on the active or
// provided profile name.
func (s *Service) ResolveSettingsPath(appDataDir string, name *string) (string, error) {
	if name == nil {
		profile, err := s.profileService.GetActiveSettingsProfile()
		if err != nil {
			return "", fmt.Errorf("getting active settings profile: %w", err)
		}
		name = &profile.Name
	}
	if *name == "default" {
		return filepath.Join(appDataDir, "config.toml"), nil
	}
	return filepath.Join(appDataDir, *name+"_config.toml"), nil
}
