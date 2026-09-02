package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"gorm.io/gorm"
)

type CreateCommandConfig struct {
	Alias       string
	Template    string
	Description *string
	Tags        []string
	Cwd         *string
	Shell       *string
	Env         map[string]string
	Timeout     *int
	ProfileName *string
}

type UpdateCommandConfig struct {
	NewAlias    *string
	Template    *string
	Description *string

	Cwd      *string
	ClearCwd bool

	Shell      *string
	ClearShell bool

	Env      map[string]string
	ClearEnv bool

	Timeout      *int
	ClearTimeout bool
}

type CommandService interface {
	CreateCommand(input CreateCommandConfig) (*models.Command, error)
	UpdateCommand(alias string, profileName *string, input UpdateCommandConfig) (*models.Command, error)
	DeleteCommand(alias string, profileName *string) error
	AddTags(alias string, tagNames []string, profileName *string) (repository.TagAttachResult, error)
	RemoveTags(alias string, tagNames []string, profileName *string) (repository.TagDetachResult, error)
	GetCommand(alias string, profileName *string) (*models.Command, error)
	GetCommandOrNil(alias string, profileName *string) (*models.Command, error)
	GetCommandByID(id uint, profileName *string) (*models.Command, error)
	ListCommands(orderBy string, tagNames []string, limit *int, profileName *string) ([]models.Command, error)
	SearchCommands(query string, fields []string, limit *int, profileName *string) ([]models.Command, error)
	MoveCommand(alias string, targetProfileName, profileName *string) (*models.Command, error)
	CopyCommand(alias string, targetProfileName, newAlias, profileName *string) (*models.Command, error)
}

type commandService struct {
	db          *gorm.DB
	commandRepo repository.CommandRepository
	tagRepo     repository.TagRepository
	profileRepo repository.ProfileRepository
}

func NewCommandService(
	db *gorm.DB,
	commandRepo repository.CommandRepository,
	tagRepo repository.TagRepository,
	profileRepo repository.ProfileRepository,
) CommandService {
	return &commandService{
		db:          db,
		commandRepo: commandRepo,
		tagRepo:     tagRepo,
		profileRepo: profileRepo,
	}
}

func (s *commandService) resolveProfile(profileName *string) (*models.Profile, error) {
	if profileName == nil {
		profile, err := s.profileRepo.GetActiveCommandProfile()
		if err != nil {
			return nil, err
		}
		return profile, nil
	}
	return s.profileRepo.GetByName(*profileName)
}

// CreateCommand creates a new command with the specified configuration, profile, and tags, and stores it in the database.
// Command is created and tags added within a transaction, so if any part fails, the entire transaction is rolled back.
func (s *commandService) CreateCommand(input CreateCommandConfig) (*models.Command, error) {
	profile, err := s.resolveProfile(input.ProfileName)
	if err != nil {
		return nil, fmt.Errorf("resolving profile: %w", err)
	}

	tags, err := s.getTags(input.Tags)
	if err != nil {
		return nil, fmt.Errorf("getting tags: %w", err)
	}

	var cmd *models.Command
	err = s.db.Transaction(func(tx *gorm.DB) error {
		repo := s.commandRepo.WithTx(tx)

		config := repository.CommandCreateConfig{
			Alias:       input.Alias,
			Template:    input.Template,
			Description: input.Description,
			Cwd:         input.Cwd,
			Shell:       input.Shell,
			Env:         input.Env,
			Timeout:     input.Timeout,
			ProfileID:   &profile.ID,
		}
		cmd, err = repo.Create(config)
		if err != nil {
			return fmt.Errorf("creating command: %w", err)
		}

		if len(tags) > 0 {
			if _, err := repo.AddTags(cmd, tags); err != nil {
				return fmt.Errorf("adding tags: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func (s *commandService) UpdateCommand(alias string, profileName *string, input UpdateCommandConfig) (*models.Command, error) {
	cmd, err := s.GetCommand(alias, profileName)
	if err != nil {
		return nil, fmt.Errorf("getting command: %w", err)
	}

	config := repository.CommandUpdateConfig{
		Alias:        input.NewAlias,
		Template:     input.Template,
		Description:  input.Description,
		Cwd:          input.Cwd,
		ClearCwd:     input.ClearCwd,
		Shell:        input.Shell,
		ClearShell:   input.ClearShell,
		Env:          input.Env,
		ClearEnv:     input.ClearEnv,
		Timeout:      input.Timeout,
		ClearTimeout: input.ClearTimeout,
	}

	updatedCmd, err := s.commandRepo.Update(cmd, config)
	if err != nil {
		return nil, fmt.Errorf("updating command: %w", err)
	}

	return updatedCmd, nil
}

func (s *commandService) DeleteCommand(alias string, profileName *string) error {
	cmd, err := s.GetCommand(alias, profileName)
	if err != nil {
		return fmt.Errorf("getting command: %w", err)
	}

	if err = s.commandRepo.Delete(cmd); err != nil {
		return fmt.Errorf("deleting command: %w", err)
	}
	return nil
}

func (s *commandService) AddTags(alias string, tagNames []string, profileName *string) (repository.TagAttachResult, error) {
	cmd, err := s.GetCommand(alias, profileName)
	if err != nil {
		return repository.TagAttachResult{}, fmt.Errorf("getting command: %w", err)
	}

	tags, err := s.getTags(tagNames)
	if err != nil {
		return repository.TagAttachResult{}, fmt.Errorf("getting tags: %w", err)
	}

	return s.commandRepo.AddTags(cmd, tags)
}

func (s *commandService) RemoveTags(alias string, tagNames []string, profileName *string) (repository.TagDetachResult, error) {
	cmd, err := s.GetCommand(alias, profileName)
	if err != nil {
		return repository.TagDetachResult{}, fmt.Errorf("getting command: %w", err)
	}

	tags, err := s.getTags(tagNames)
	if err != nil {
		return repository.TagDetachResult{}, fmt.Errorf("getting tags: %w", err)
	}

	return s.commandRepo.RemoveTags(cmd, tags)
}

func (s *commandService) GetCommand(alias string, profileName *string) (*models.Command, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving profile: %w", err)
	}

	return s.commandRepo.GetByAlias(alias, &profile.ID)
}

func (s *commandService) GetCommandOrNil(alias string, profileName *string) (*models.Command, error) {
	cmd, err := s.GetCommand(alias, profileName)
	if err != nil {
		if errors.Is(err, repository.ErrUnknownAlias) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting command: %w", err)
	}
	return cmd, nil
}

func (s *commandService) GetCommandByID(id uint, profileName *string) (*models.Command, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving profile: %w", err)
	}
	return s.commandRepo.GetByID(id, &profile.ID)
}

func (s *commandService) ListCommands(
	orderBy string,
	tagNames []string,
	limit *int,
	profileName *string,
) ([]models.Command, error) {
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
		return s.commandRepo.ListByTags(tags, orderBy, resolvedLimit, &profile.ID)
	}
	return s.commandRepo.ListAll(orderBy, resolvedLimit, &profile.ID)
}

func (s *commandService) SearchCommands(
	query string,
	fields []string,
	limit *int,
	profileName *string,
) ([]models.Command, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving profile: %w", err)
	}

	if len(fields) <= 0 {
		fields = []string{"alias", "template", "description"}
	}

	resolvedLimit := 0
	if limit != nil {
		resolvedLimit = *limit
	}

	return s.commandRepo.Search(query, fields, resolvedLimit, &profile.ID)
}

func (s *commandService) MoveCommand(alias string, targetProfileName, profileName *string) (*models.Command, error) {
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

	cmd, err := s.commandRepo.GetByAlias(alias, &sourceProfile.ID)
	if err != nil {
		return nil, fmt.Errorf("getting command by alias: %w", err)
	}

	input := repository.CommandUpdateConfig{ProfileID: &targetProfile.ID}
	updatedCmd, err := s.commandRepo.Update(cmd, input)
	if err != nil {
		return nil, fmt.Errorf("updating command: %w", err)
	}

	return updatedCmd, nil
}

func (s *commandService) CopyCommand(
	alias string,
	targetProfileName,
	newAlias,
	profileName *string,
) (*models.Command, error) {
	if targetProfileName == nil {
		return nil, ErrNoTargetProfile
	}

	sourceProfile, err := s.resolveProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("resolving source profile: %w", err)
	}

	cmd, err := s.commandRepo.GetByAlias(alias, &sourceProfile.ID)
	if err != nil {
		return nil, fmt.Errorf("getting command by alias: %w", err)
	}

	resolvedAlias := alias
	if newAlias != nil {
		resolvedAlias = *newAlias
	}

	envSource := ""
	if cmd.Env != nil {
		envSource = *cmd.Env
	}
	parsedEnv, err := parseEnv(envSource)
	if err != nil {
		return nil, fmt.Errorf("parsing env: %w", err)
	}

	input := CreateCommandConfig{
		Alias:       resolvedAlias,
		Template:    cmd.Template,
		Description: cmd.Description,
		Cwd:         cmd.Cwd,
		Shell:       cmd.Shell,
		Env:         parsedEnv,
		Timeout:     cmd.Timeout,
		ProfileName: targetProfileName,
	}

	newCmd, err := s.CreateCommand(input)
	if err != nil {
		return nil, fmt.Errorf("creating command: %w", err)
	}

	return newCmd, nil
}

// getTags retrieves a list of tags by their names using the tag repository and returns them or an error if any occurs.
func (s *commandService) getTags(tagNames []string) ([]models.Tag, error) {
	return getTags(tagNames, s.tagRepo)
}

// parseEnv parses a JSON-encoded string into a map of environment variables and returns it or an error if parsing fails.
func parseEnv(source string) (map[string]string, error) {
	if source == "" {
		return nil, nil
	}
	var env map[string]string
	if err := json.Unmarshal([]byte(source), &env); err != nil {
		return nil, fmt.Errorf("parsing env: %w", err)
	}
	return env, nil
}
