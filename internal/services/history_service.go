package services

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
)

type HistoryService interface {
	GetRecent(alias *string, limit *int, profileName *string) ([]models.CommandHistory, error)
	GetByRef(ref string, alias *string, profileName *string) (*models.CommandHistory, error)
	GetVariables(entry *models.CommandHistory) (map[string]string, error)
	DeleteByRef(ref string, alias *string, profileName *string) error
	Clear(alias *string, profileName *string) (int, error)
}

type historyService struct {
	historyRepo repository.HistoryRepository
	profileRepo repository.ProfileRepository
}

func NewHistoryService(historyRepo repository.HistoryRepository, profileRepo repository.ProfileRepository) HistoryService {
	return &historyService{
		historyRepo: historyRepo,
		profileRepo: profileRepo,
	}
}

// getProfileID resolves the profile ID by name or retrieves the active profile's ID if no name is provided.
func (s *historyService) getProfileID(profileName *string) (*uint, error) {
	var profile *models.Profile
	var err error
	if profileName != nil {
		profile, err = s.profileRepo.GetByName(*profileName)
		if err != nil {
			return nil, fmt.Errorf("getting active profile: %w", err)
		}
	} else {
		profile, err = s.profileRepo.GetActiveCommandProfile()
		if err != nil {
			return nil, fmt.Errorf("getting active profile: %w", err)
		}
	}
	return &profile.ID, nil
}

func (s *historyService) GetRecent(alias *string, limit *int, profileName *string) ([]models.CommandHistory, error) {
	profileID, err := s.getProfileID(profileName)
	if err != nil {
		return nil, fmt.Errorf("getting profile ID: %w", err)
	}
	resolvedLimit := 0
	if limit != nil {
		resolvedLimit = *limit
	}
	return s.historyRepo.GetRecent(alias, resolvedLimit, profileID)
}

func (s *historyService) GetByRef(ref string, alias *string, profileName *string) (*models.CommandHistory, error) {
	if index, err := strconv.Atoi(ref); err == nil {
		return s.getByIndex(index, alias, profileName)
	}

	profileID, err := s.getProfileID(profileName)
	if err != nil {
		return nil, fmt.Errorf("getting profile ID: %w", err)
	}

	return s.historyRepo.GetByID(ref, profileID)
}

func (s *historyService) GetVariables(entry *models.CommandHistory) (map[string]string, error) {
	if entry.VariablesUsed == nil {
		return nil, nil
	}

	var variables map[string]string
	if err := json.Unmarshal([]byte(*entry.VariablesUsed), &variables); err != nil {
		return nil, fmt.Errorf("unmarshaling variables: %w", err)
	}
	return variables, nil
}

func (s *historyService) DeleteByRef(ref string, alias *string, profileName *string) error {
	entry, err := s.GetByRef(ref, alias, profileName)
	if err != nil {
		return fmt.Errorf("getting history entry: %w", err)
	}

	profileID, err := s.getProfileID(profileName)
	if err != nil {
		return fmt.Errorf("getting profile ID: %w", err)
	}

	return s.historyRepo.DeleteByID(entry.ID, profileID)
}

func (s *historyService) Clear(alias *string, profileName *string) (int, error) {
	profileID, err := s.getProfileID(profileName)
	if err != nil {
		return 0, fmt.Errorf("getting profile ID: %w", err)
	}

	return s.historyRepo.Clear(alias, profileID)
}

// GetByIndex retrieves the nth most recent command history entry, based on the given index, alias, and profile name.
// Returns ErrUnknownHistoryIndex if the index is invalid or out of bounds.
func (s *historyService) getByIndex(index int, alias *string, profileName *string) (*models.CommandHistory, error) {
	if index < 1 {
		return nil, ErrUnknownHistoryIndex
	}
	entries, err := s.GetRecent(alias, &index, profileName)
	if err != nil {
		return nil, fmt.Errorf("getting recent history: %w", err)
	}
	if index > len(entries) {
		return nil, ErrUnknownHistoryIndex
	}
	return &entries[index-1], nil
}
