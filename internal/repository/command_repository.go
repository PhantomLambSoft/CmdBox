package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
	"gorm.io/gorm"
)

var commandOrderableColumns = map[string]bool{
	"id":          true,
	"alias":       true,
	"template":    true,
	"description": true,
	"cwd":         true,
	"shell":       true,
	"timeout":     true,
	"created_at":  true,
	"updated_at":  true,
	"used":        true,
	"last_used":   true,
	"profile_id":  true,
}

var defaultCommandSearchFields = []string{"alias", "template", "description"}

type CommandCreateConfig struct {
	Alias       string
	Template    string
	Description *string
	Cwd         *string
	Shell       *string
	Env         map[string]string
	Timeout     *int
	ProfileID   *uint
}

type CommandUpdateConfig struct {
	Alias       *string
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

type CommandRepository interface {
	Create(input CommandCreateConfig) (*models.Command, error)
	GetByAlias(alias string, profileID *uint) (*models.Command, error)
	GetByID(id uint, profileID *uint) (*models.Command, error)
	Update(command *models.Command, input CommandUpdateConfig) (*models.Command, error)
	Delete(command *models.Command) error

	// RecordUse automatically increments the used count of a command with the supplied id.
	RecordUse(commandID uint) error

	ListAll(orderBy string, limit int, profileID *uint) ([]models.Command, error)
	ListByTags(tags []models.Tag, orderBy string, limit int, profileID *uint) ([]models.Command, error)
	Search(query string, fields []string, limit int, profileID *uint) ([]models.Command, error)

	AddTags(command *models.Command, tags []models.Tag) (TagAttachResult, error)
	RemoveTags(command *models.Command, tags []models.Tag) (TagDetachResult, error)
}

type commandRepository struct {
	db          *gorm.DB
	profileRepo ProfileRepository
	validator   *validate.CommandValidator
}

func NewCommandRepository(db *gorm.DB, profileRepo ProfileRepository, validator *validate.CommandValidator) CommandRepository {
	if validator == nil {
		validator = validate.NewCommandValidator(nil)
	}
	return &commandRepository{db: db, profileRepo: profileRepo, validator: validator}
}

// resolveProfileID determines the profile ID to use, returning the provided ID if non-nil or the active profile ID otherwise.
// It queries the active profile state from the profile repository if the input ID is nil.
// Returns the resolved profile ID or an error if the profile state retrieval fails.
func (r *commandRepository) resolveProfileID(profileID *uint) (uint, error) {
	if profileID != nil {
		return *profileID, nil
	}
	state, err := r.profileRepo.GetState()
	if err != nil {
		return 0, err
	}
	return state.ActiveCommandProfileID, nil
}

// encodeEnv encodes a map of environment variables into a JSON string and returns a pointer to the encoded string or an error.
func encodeEnv(env map[string]string) (*string, error) {
	if len(env) == 0 {
		return nil, nil
	}
	data, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("failed to encode environment variables: %w", err)
	}
	s := string(data)
	return &s, nil
}

// validateNoClearAndSet checks for conflicts between clear flags and corresponding field values in CommandUpdateConfig.
// Returns an error if both a clear flag and a non-nil value are set for the same field.
func validateNoClearAndSet(input CommandUpdateConfig) error {
	if input.ClearCwd && input.Cwd != nil {
		return fmt.Errorf("%w: cwd", ErrConflictingClearAndSet)
	}
	if input.ClearShell && input.Shell != nil {
		return fmt.Errorf("%w: shell", ErrConflictingClearAndSet)
	}
	if input.ClearTimeout && input.Timeout != nil {
		return fmt.Errorf("%w: timeout", ErrConflictingClearAndSet)
	}
	if input.ClearEnv && input.Env != nil {
		return fmt.Errorf("%w: env", ErrConflictingClearAndSet)
	}
	return nil
}

func (r *commandRepository) Create(input CommandCreateConfig) (*models.Command, error) {
	alias := strings.TrimSpace(input.Alias)
	if err := r.validator.ValidateCreate(alias, input.Template); err != nil {
		return nil, err
	}

	profileID, err := r.resolveProfileID(input.ProfileID)
	if err != nil {
		return nil, err
	}

	envJSON, err := encodeEnv(input.Env)
	if err != nil {
		return nil, err
	}

	command := &models.Command{
		Alias:       alias,
		Template:    input.Template,
		Description: input.Description,
		Cwd:         input.Cwd,
		Shell:       input.Shell,
		Timeout:     input.Timeout,
		Env:         envJSON,
		ProfileID:   profileID,
	}

	if err := r.db.Create(command).Error; err != nil {
		if isUniqueConstraintViolation(err, "commands", "alias") {
			return nil, fmt.Errorf("%w: %q", ErrAliasConflict, alias)
		}
		return nil, fmt.Errorf("creating command %q:%w", alias, err)
	}
	return command, nil
}

func (r *commandRepository) GetByAlias(alias string, profileID *uint) (*models.Command, error) {
	pid, err := r.resolveProfileID(profileID)
	if err != nil {
		return nil, err
	}

	var command models.Command
	err = r.db.Where("alias = ? AND profile_id = ?", alias, pid).First(&command).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %q", ErrUnknownAlias, alias)
	}
	if err != nil {
		return nil, fmt.Errorf("looking up alias %q:%w", alias, err)
	}
	return &command, err
}

func (r *commandRepository) GetByID(id uint, profileID *uint) (*models.Command, error) {
	pid, err := r.resolveProfileID(profileID)
	if err != nil {
		return nil, err
	}

	var command models.Command
	err = r.db.Where("id = ? AND profile_id = ?", id, pid).First(&command).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: id %d", ErrUnknownCommand, id)
	}
	if err != nil {
		return nil, fmt.Errorf("looking up command %d:%w", id, err)
	}
	return &command, nil
}

func (r *commandRepository) Update(command *models.Command, input CommandUpdateConfig) (*models.Command, error) {
	if command == nil {
		return nil, ErrNoUpdateTarget
	}
	if err := validateNoClearAndSet(input); err != nil {
		return nil, err
	}

	mergedAlias := command.Alias
	if input.Alias != nil {
		mergedAlias = strings.TrimSpace(*input.Alias)
	}
	mergedTemplate := command.Template
	if input.Template != nil {
		mergedTemplate = *input.Template
	}

	if err := r.validator.ValidateUpdate(&mergedAlias, &mergedTemplate); err != nil {
		return nil, err
	}

	if input.Alias != nil {
		command.Alias = mergedAlias
	}
	if input.Template != nil {
		command.Template = mergedTemplate
	}
	if input.Description != nil {
		command.Description = input.Description
	}

	switch {
	case input.ClearCwd:
		command.Cwd = nil
	case input.Cwd != nil:
		command.Cwd = input.Cwd
	}

	switch {
	case input.ClearShell:
		command.Shell = nil
	case input.Shell != nil:
		command.Shell = input.Shell
	}

	switch {
	case input.ClearTimeout:
		command.Timeout = nil
	case input.Timeout != nil:
		command.Timeout = input.Timeout
	}

	switch {
	case input.ClearEnv:
		command.Env = nil
	case input.Env != nil:
		envJSON, err := encodeEnv(input.Env)
		if err != nil {
			return nil, err
		}
		command.Env = envJSON
	}

	if err := r.db.Save(command).Error; err != nil {
		if isUniqueConstraintViolation(err, "commands", "alias") {
			return nil, fmt.Errorf("%w: %q", ErrAliasConflict, mergedAlias)
		}
		return nil, fmt.Errorf("updating command id %d: %w", command.ID, err)
	}
	return command, nil
}

func (r *commandRepository) Delete(command *models.Command) error {
	if command == nil {
		return nil
	}
	if err := r.db.Delete(command).Error; err != nil {
		return fmt.Errorf("deleting command id %d: %w", command.ID, err)
	}
	return nil
}

func (r *commandRepository) RecordUse(commandID uint) error {
	result := r.db.Model(&models.Command{}).
		Where("id = ?", commandID).
		Updates(map[string]interface{}{
			"used":      gorm.Expr("used + 1"),
			"last_used": time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("recording use of command id %d: %w", commandID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: id %d", ErrUnknownCommand, commandID)
	}
	return nil
}

func (r *commandRepository) AddTags(command *models.Command, tags []models.Tag) (TagAttachResult, error) {
	if len(tags) == 0 {
		return TagAttachResult{}, nil
	}

	var added, existing []string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, tag := range tags {
			var link models.CommandTag
			err := tx.Where("command_id = ? AND tag_id = ?", command.ID, tag.ID).First(&link).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				newLink := models.CommandTag{CommandID: command.ID, TagID: tag.ID}
				if err := tx.Create(&newLink).Error; err != nil {
					return err
				}
				added = append(added, tag.Name)
			case err != nil:
				return err
			default:
				existing = append(existing, tag.Name)
			}
		}
		return nil
	})
	if err != nil {
		return TagAttachResult{}, fmt.Errorf("%w: %v", ErrTagAttachFailed, err)
	}
	return TagAttachResult{Added: added, Existing: existing}, nil
}

func (r *commandRepository) RemoveTags(command *models.Command, tags []models.Tag) (TagDetachResult, error) {
	if len(tags) == 0 {
		return TagDetachResult{}, nil
	}

	var removed, notAttached []string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, tag := range tags {
			result := tx.Where("command_id = ? AND tag_id = ?", command.ID, tag.ID).Delete(&models.CommandTag{})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				removed = append(removed, tag.Name)
			} else {
				notAttached = append(notAttached, tag.Name)
			}
		}
		return nil
	})
	if err != nil {
		return TagDetachResult{}, fmt.Errorf("%w: %v", ErrTagDetachFailed, err)
	}
	return TagDetachResult{Removed: removed, NotAttached: notAttached}, nil
}

func (r *commandRepository) ListAll(orderBy string, limit int, profileID *uint) ([]models.Command, error) {
	if orderBy == "" {
		orderBy = "alias"
	}
	if limit <= 0 {
		limit = 25
	}
	orderClauses, err := resolveOrdering(orderBy, commandOrderableColumns)
	if err != nil {
		return nil, err
	}
	pid, err := r.resolveProfileID(profileID)
	if err != nil {
		return nil, err
	}

	var commands []models.Command
	q := r.db.Where("profile_id = ?", pid)
	for _, clause := range orderClauses {
		q = q.Order(clause)
	}
	if err := q.Limit(limit).Find(&commands).Error; err != nil {
		return nil, fmt.Errorf("listing commands: %w", err)
	}
	return commands, nil
}

func (r *commandRepository) ListByTags(tags []models.Tag, orderBy string, limit int, profileID *uint) ([]models.Command, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	if orderBy == "" {
		orderBy = "alias"
	}
	if limit <= 0 {
		limit = 25
	}
	orderClauses, err := resolveOrdering(orderBy, commandOrderableColumns)
	if err != nil {
		return nil, err
	}
	pid, err := r.resolveProfileID(profileID)
	if err != nil {
		return nil, err
	}

	tagsIDs := make([]uint, len(tags))
	for i, t := range tags {
		tagsIDs[i] = t.ID
	}

	var commands []models.Command
	q := r.db.Distinct("commands.*").
		Joins("JOIN command_tags ON command_tags.command_id = commands.id").
		Where("command_tags.tag_id IN ? AND commands.profile_id = ?", tagsIDs, pid)
	for _, clause := range orderClauses {
		q = q.Order(clause)
	}
	if err := q.Limit(limit).Find(&commands).Error; err != nil {
		return nil, fmt.Errorf("listing commands by tag: %w", err)
	}
	return commands, nil
}

func (r *commandRepository) Search(query string, fields []string, limit int, profileID *uint) ([]models.Command, error) {
	if len(fields) == 0 {
		fields = defaultCommandSearchFields
	}
	if limit <= 0 {
		limit = 25
	}
	pid, err := r.resolveProfileID(profileID)
	if err != nil {
		return nil, err
	}

	var commands []models.Command
	searchInput := searchWithRelevanceInput{
		DB:                   r.db,
		Table:                "commands",
		Query:                query,
		Fields:               fields,
		AllowedColumns:       commandOrderableColumns,
		SecondaryOrderColumn: "alias",
		Limit:                limit,
		ExtraWhere:           "profile_id = ?",
		ExtraArgs:            []interface{}{pid},
		Dest:                 &commands,
	}
	err = searchWithRelevance(searchInput)
	if err != nil {
		return nil, fmt.Errorf("searching commands: %w", err)
	}
	return commands, nil
}
