package services

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
)

type ProfileStatus struct {
	CommandProfile  string
	VariableProfile string
	SettingsProfile string
}

func (p *ProfileStatus) Linked() bool {
	return p.CommandProfile == p.VariableProfile && p.VariableProfile == p.SettingsProfile
}

type UpdateProfileConfig struct {
	NewName     *string
	Description *string
}

type ProfileService interface {
	CreateProfile(name string, description *string) (*models.Profile, error)
	UpdateProfile(name string, input UpdateProfileConfig) (*models.Profile, error)
	DeleteProfile(name string, force bool) error
	GetProfile(name string) (*models.Profile, error)
	ListProfiles(orderBy string, limit *int) ([]models.Profile, error)
	SearchProfiles(query string, fields []string, limit *int) ([]models.Profile, error)

	SwitchProfile(name string) (*models.ProfileState, error)
	SwitchCommandProfile(name string) (*models.ProfileState, error)
	SwitchVariableProfile(name string) (*models.ProfileState, error)
	SwitchSettingsProfile(name string) (*models.ProfileState, error)
	GetStatus() (ProfileStatus, error)
}

type profileService struct {
	profileRepo repository.ProfileRepository
}

func NewProfileService(repo repository.ProfileRepository) ProfileService {
	return &profileService{profileRepo: repo}
}

func (s *profileService) CreateProfile(name string, description *string) (*models.Profile, error) {
	profile, err := s.profileRepo.Create(name, description)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *profileService) UpdateProfile(name string, input UpdateProfileConfig) (*models.Profile, error) {
	profile, err := s.GetProfile(name)
	if err != nil {
		return nil, fmt.Errorf("getting profile: %w", err)
	}

	config := repository.ProfileUpdateConfig{
		Name:        input.NewName,
		Description: input.Description,
	}

	updatedProfile, err := s.profileRepo.Update(profile, config)
	if err != nil {
		return nil, fmt.Errorf("updating profile: %w", err)
	}

	return updatedProfile, nil
}

func (s *profileService) DeleteProfile(name string, force bool) error {
	profile, err := s.GetProfile(name)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	return s.profileRepo.Delete(profile, force)
}

func (s *profileService) GetProfile(name string) (*models.Profile, error) {
	return s.profileRepo.GetByName(name)
}

func (s *profileService) ListProfiles(orderBy string, limit *int) ([]models.Profile, error) {
	resolvedLimit := 0
	if limit != nil {
		resolvedLimit = *limit
	}
	return s.profileRepo.ListAll(orderBy, resolvedLimit)
}

func (s *profileService) SearchProfiles(query string, fields []string, limit *int) ([]models.Profile, error) {
	if len(fields) <= 0 {
		fields = []string{"name", "description"}
	}

	resolvedLimit := 0
	if limit != nil {
		resolvedLimit = *limit
	}

	return s.profileRepo.Search(query, fields, resolvedLimit)
}

func (s *profileService) SwitchProfile(name string) (*models.ProfileState, error) {
	profile, err := s.GetProfile(name)
	if err != nil {
		return nil, fmt.Errorf("getting profile: %w", err)
	}

	config := repository.SetProfileConfig{
		CommandProfile:  profile,
		VariableProfile: profile,
		SettingsProfile: profile,
	}

	state, err := s.profileRepo.SetActiveProfile(config)
	if err != nil {
		return nil, fmt.Errorf("setting active profile: %w", err)
	}

	return state, nil
}

func (s *profileService) SwitchCommandProfile(name string) (*models.ProfileState, error) {
	profile, err := s.GetProfile(name)
	if err != nil {
		return nil, fmt.Errorf("getting profile: %w", err)
	}

	config := repository.SetProfileConfig{
		CommandProfile: profile,
	}

	state, err := s.profileRepo.SetActiveProfile(config)
	if err != nil {
		return nil, fmt.Errorf("setting active command profile: %w", err)
	}

	return state, nil
}

func (s *profileService) SwitchVariableProfile(name string) (*models.ProfileState, error) {
	profile, err := s.GetProfile(name)
	if err != nil {
		return nil, fmt.Errorf("getting profile: %w", err)
	}

	config := repository.SetProfileConfig{
		VariableProfile: profile,
	}

	state, err := s.profileRepo.SetActiveProfile(config)
	if err != nil {
		return nil, fmt.Errorf("setting active variable profile: %w", err)
	}

	return state, nil
}

func (s *profileService) SwitchSettingsProfile(name string) (*models.ProfileState, error) {
	profile, err := s.GetProfile(name)
	if err != nil {
		return nil, fmt.Errorf("getting profile: %w", err)
	}

	config := repository.SetProfileConfig{
		SettingsProfile: profile,
	}

	state, err := s.profileRepo.SetActiveProfile(config)
	if err != nil {
		return nil, fmt.Errorf("setting active settings profile: %w", err)
	}

	return state, nil
}

func (s *profileService) GetStatus() (ProfileStatus, error) {
	state, err := s.profileRepo.GetStateWithProfiles()
	if err != nil {
		return ProfileStatus{}, err
	}
	status := ProfileStatus{
		CommandProfile:  state.ActiveCommandProfile.Name,
		VariableProfile: state.ActiveVariableProfile.Name,
		SettingsProfile: state.ActiveSettingsProfile.Name,
	}
	return status, nil
}
