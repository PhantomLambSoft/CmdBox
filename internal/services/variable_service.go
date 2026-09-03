package services

import (
	"errors"
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"gorm.io/gorm"
)

type CreateVariableConfig struct {
	Name        string
	Value       string
	Tags        []string
	ProfileName *string
}

type UpdateVariableConfig struct {
	NewName *string
	Value   *string
}

type VariableService interface {
	CreateVariable(input CreateVariableConfig) (*models.Variable, error)
	UpdateVariable(name string, profileName *string, input UpdateVariableConfig) (*models.Variable, error)
	DeleteVariable(name string, profileName *string) error
	AddTags(name string, tagNames []string, profileName *string) (repository.TagAttachResult, error)
	RemoveTags(name string, tagNames []string, profileName *string) (repository.TagDetachResult, error)
	GetVariable(name string, profileName *string) (*models.Variable, error)
	GetVariableOrNil(name string, profileName *string) (*models.Variable, error)
	GetVariableByID(id uint, profileName *string) (*models.Variable, error)
	ListVariables(orderBy string, tagNames []string, limit *int, profileName *string) ([]models.Variable, error)
	SearchVariables(query string, fields []string, limit *int, profileName *string) ([]models.Variable, error)
	MoveVariable(name string, targetProfileName, profileName *string) (*models.Variable, error)
	CopyVariable(name string, targetProfileName, newName, profileName *string) (*models.Variable, error)
}

type variableService struct {
	db           *gorm.DB
	variableRepo repository.VariableRepository
	tagRepo      repository.TagRepository
	profileRepo  repository.ProfileRepository
}

func NewVariableService(
	db *gorm.DB,
	variableRepo repository.VariableRepository,
	tagRepo repository.TagRepository,
	profileRepo repository.ProfileRepository,
) VariableService {
	return &variableService{
		db:           db,
		variableRepo: variableRepo,
		tagRepo:      tagRepo,
		profileRepo:  profileRepo,
	}
}

func (s *variableService) resolveProfile(profileName *string) (*models.Profile, error) {
	if profileName == nil {
		profile, err := s.profileRepo.GetActiveVariableProfile()
		if err != nil {
			return nil, err
		}
		return profile, nil
	}
	return s.profileRepo.GetByName(*profileName)
}

func (s *variableService) getTags(tagNames []string) ([]models.Tag, error) {
	return getTags(tagNames, s.tagRepo)
}

func (s *variableService) CreateVariable(input CreateVariableConfig) (*models.Variable, error) {
	profile, err := s.resolveProfile(input.ProfileName)
	if err != nil {
		return nil, fmt.Errorf("resolving profile: %w", err)
	}

	tags, err := s.getTags(input.Tags)
	if err != nil {
		return nil, fmt.Errorf("getting tags: %w", err)
	}

	var variable *models.Variable
	err = s.db.Transaction(func(tx *gorm.DB) error {
		repo := s.variableRepo.WithTx(tx)

		config := repository.VariableCreateConfig{
			Name:      input.Name,
			Value:     input.Value,
			ProfileID: &profile.ID,
		}
		variable, err = repo.Create(config)
		if err != nil {
			return fmt.Errorf("creating variable: %w", err)
		}

		if len(tags) > 0 {
			if _, err := repo.AddTags(variable, tags); err != nil {
				return fmt.Errorf("adding tags: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return variable, nil
}

func (s *variableService) UpdateVariable(
	name string,
	profileName *string,
	input UpdateVariableConfig,
) (*models.Variable, error) {
	variable, err := s.GetVariable(name, profileName)
	if err != nil {
		return nil, fmt.Errorf("getting variable: %w", err)
	}

	config := repository.VariableUpdateConfig{
		Name:  input.NewName,
		Value: input.Value,
	}

	updatedVariable, err := s.variableRepo.Update(variable, config)
	if err != nil {
		return nil, fmt.Errorf("updating variable: %w", err)
	}

	return updatedVariable, nil
}

func (s *variableService) DeleteVariable(name string, profileName *string) error {
	variable, err := s.GetVariable(name, profileName)
	if err != nil {
		return fmt.Errorf("getting variable: %w", err)
	}

	if err = s.variableRepo.Delete(variable); err != nil {
		return fmt.Errorf("deleting variable: %w", err)
	}
	return nil
}

func (s *variableService) AddTags(name string, tagNames []string, profileName *string) (repository.TagAttachResult, error) {
	variable, err := s.GetVariable(name, profileName)
	if err != nil {
		return repository.TagAttachResult{}, fmt.Errorf("getting variable: %w", err)
	}

	tags, err := s.getTags(tagNames)
	if err != nil {
		return repository.TagAttachResult{}, fmt.Errorf("getting tags: %w", err)
	}

	return s.variableRepo.AddTags(variable, tags)
}

func (s *variableService) RemoveTags(name string, tagNames []string, profileName *string) (repository.TagDetachResult, error) {
	variable, err := s.GetVariable(name, profileName)
	if err != nil {
		return repository.TagDetachResult{}, fmt.Errorf("getting variable: %w", err)
	}

	tags, err := s.getTags(tagNames)
	if err != nil {
		return repository.TagDetachResult{}, fmt.Errorf("getting tags: %w", err)
	}

	return s.variableRepo.RemoveTags(variable, tags)
}

func (s *variableService) GetVariable(name string, profileName *string) (*models.Variable, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving profile: %w", err)
	}

	return s.variableRepo.GetByName(name, &profile.ID)
}

func (s *variableService) GetVariableOrNil(name string, profileName *string) (*models.Variable, error) {
	variable, err := s.GetVariable(name, profileName)
	if err != nil {
		if errors.Is(err, repository.ErrUnKnownName) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting variable: %w", err)
	}
	return variable, nil
}

func (s *variableService) GetVariableByID(id uint, profileName *string) (*models.Variable, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving profile: %w", err)
	}
	return s.variableRepo.GetByID(id, &profile.ID)
}

func (s *variableService) ListVariables(orderBy string, tagNames []string, limit *int, profileName *string) ([]models.Variable, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving profile: %w", err)
	}

	resolvedLimit := 0
	if limit != nil {
		resolvedLimit = *limit
	}

	if len(tagNames) > 0 {
		tags, err := s.getTags(tagNames)
		if err != nil {
			return nil, fmt.Errorf("getting tags: %w", err)
		}
		return s.variableRepo.ListByTags(tags, orderBy, resolvedLimit, &profile.ID)
	}
	return s.variableRepo.ListAll(orderBy, resolvedLimit, &profile.ID)
}

func (s *variableService) SearchVariables(query string, fields []string, limit *int, profileName *string) ([]models.Variable, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving profile: %w", err)
	}

	if len(fields) <= 0 {
		fields = []string{"name", "value"}
	}

	resolvedLimit := 0
	if limit != nil {
		resolvedLimit = *limit
	}

	return s.variableRepo.Search(query, fields, resolvedLimit, &profile.ID)
}

func (s *variableService) MoveVariable(name string, targetProfileName, profileName *string) (*models.Variable, error) {
	if targetProfileName == nil {
		return nil, ErrNoTargetProfile
	}

	targetProfile, err := s.resolveProfile(targetProfileName)
	if err != nil {
		return nil, fmt.Errorf("resolving target profile: %w", err)
	}

	sourceProfile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving source profile: %w", err)
	}

	variable, err := s.variableRepo.GetByName(name, &sourceProfile.ID)
	if err != nil {
		return nil, fmt.Errorf("getting variable by alias: %w", err)
	}

	input := repository.VariableUpdateConfig{ProfileID: &targetProfile.ID}
	updatedVar, err := s.variableRepo.Update(variable, input)
	if err != nil {
		return nil, fmt.Errorf("updating variable: %w", err)
	}

	return updatedVar, nil
}

func (s *variableService) CopyVariable(name string, targetProfileName, newName, profileName *string) (*models.Variable, error) {
	if targetProfileName == nil {
		return nil, ErrNoTargetProfile
	}

	sourceProfile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving source profile: %w", err)
	}

	variable, err := s.variableRepo.GetByName(name, &sourceProfile.ID)
	if err != nil {
		return nil, fmt.Errorf("getting variable by alias: %w", err)
	}

	resolvedName := name
	if newName != nil {
		resolvedName = *newName
	}

	input := CreateVariableConfig{
		Name:        resolvedName,
		Value:       variable.Value,
		ProfileName: targetProfileName,
	}

	newVar, err := s.CreateVariable(input)
	if err != nil {
		return nil, fmt.Errorf("creating variable: %w", err)
	}

	return newVar, nil
}
