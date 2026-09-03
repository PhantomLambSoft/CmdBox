package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
)

const DefaultProfileName = "default"

var profileOrderableColumns = map[string]bool{
	"id":          true,
	"name":        true,
	"description": true,
	"created_at":  true,
	"updated_at":  true,
	"last_used":   true,
}

var defaultProfileSearchFields = []string{"name", "description"}

type ProfileUpdateConfig struct {
	Name        *string
	Description *string
}

type SetProfileConfig struct {
	CommandProfile  *models.Profile
	VariableProfile *models.Profile
	SettingsProfile *models.Profile
}

type ProfileRepository interface {
	WithTx(tx *gorm.DB) ProfileRepository

	Create(name string, description *string) (*models.Profile, error)
	GetByName(name string) (*models.Profile, error)
	GetByID(id uint) (*models.Profile, error)
	ListAll(orderBy string, limit int) ([]models.Profile, error)
	Search(query string, fields []string, limit int) ([]models.Profile, error)
	Update(profile *models.Profile, input ProfileUpdateConfig) (*models.Profile, error)
	Delete(profile *models.Profile, force bool) error
	RecordUse(profile *models.Profile) error

	GetState() (*models.ProfileState, error)
	GetStateWithProfiles() (*models.ProfileState, error)
	SetActiveProfile(config SetProfileConfig) (*models.ProfileState, error)

	GetActiveCommandProfile() (*models.Profile, error)
	GetActiveVariableProfile() (*models.Profile, error)
	GetActiveSettingsProfile() (*models.Profile, error)
}

type profileRepository struct {
	db        *gorm.DB
	validator *validate.ProfileValidator
}

func NewProfileRepository(db *gorm.DB, validator *validate.ProfileValidator) ProfileRepository {
	if validator == nil {
		validator = validate.NewProfileValidator(nil)
	}
	return &profileRepository{db: db, validator: validator}
}

// WithTx creates a new instance of profileRepository using the provided database transaction.
func (r *profileRepository) WithTx(tx *gorm.DB) ProfileRepository {
	return &profileRepository{db: tx, validator: r.validator}
}

func (r *profileRepository) Create(name string, description *string) (*models.Profile, error) {
	if err := r.validator.ValidateCreate(name); err != nil {
		return nil, err
	}

	_, err := r.GetByName(name)
	if err == nil {
		return nil, ErrProfileNameExists
	}
	if !errors.Is(err, ErrProfileNotFound) {
		return nil, err
	}

	profile := &models.Profile{
		Name:        name,
		Description: description,
	}
	if err := r.db.Create(profile).Error; err != nil {
		return nil, fmt.Errorf("creating profile %q: %w", name, err)
	}
	return profile, nil
}

func (r *profileRepository) GetByName(name string) (*models.Profile, error) {
	var profile models.Profile
	err := r.db.Where("name = ?", name).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("looking up profile %q: %w", name, err)
	}
	return &profile, nil
}

func (r *profileRepository) GetByID(id uint) (*models.Profile, error) {
	var profile models.Profile
	err := r.db.First(&profile, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("looking up profile id %d: %w", id, err)
	}
	return &profile, nil
}

func (r *profileRepository) ListAll(orderBy string, limit int) ([]models.Profile, error) {
	if orderBy == "" {
		orderBy = "name"
	}
	if limit <= 0 {
		limit = 25
	}
	orderClauses, err := resolveOrdering(orderBy, profileOrderableColumns)
	if err != nil {
		return nil, err
	}

	var profiles []models.Profile
	q := r.db.Model(&models.Profile{})
	for _, clause := range orderClauses {
		q = q.Order(clause)
	}
	if err := q.Limit(limit).Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("listing profiles: %w", err)
	}
	return profiles, nil
}

func (r *profileRepository) Search(query string, fields []string, limit int) ([]models.Profile, error) {
	if len(fields) == 0 {
		fields = defaultProfileSearchFields
	}
	if limit <= 0 {
		limit = 25
	}

	var profiles []models.Profile
	searchInput := searchWithRelevanceInput{
		DB:                   r.db,
		Table:                "profiles",
		Query:                query,
		Fields:               fields,
		AllowedColumns:       profileOrderableColumns,
		SecondaryOrderColumn: "name",
		Limit:                limit,
		Dest:                 &profiles,
	}
	if err := searchWithRelevance(searchInput); err != nil {
		return nil, err
	}

	return profiles, nil
}

func (r *profileRepository) Update(profile *models.Profile, input ProfileUpdateConfig) (*models.Profile, error) {
	if profile == nil {
		return nil, ErrNoUpdateTarget
	}

	mergedName := profile.Name
	if input.Name != nil {
		mergedName = strings.TrimSpace(*input.Name)
	}

	if err := r.validator.ValidateUpdate(&mergedName); err != nil {
		return nil, err
	}

	if input.Name != nil && mergedName != profile.Name {
		if profile.Name == DefaultProfileName {
			return nil, ErrDefaultProfileName
		}
		if _, err := r.GetByName(mergedName); err == nil {
			return nil, ErrProfileNameExists
		} else if !errors.Is(err, ErrProfileNotFound) {
			return nil, err
		}
		profile.Name = mergedName
	}

	if input.Description != nil {
		profile.Description = input.Description
	}

	if err := r.db.Save(profile).Error; err != nil {
		return nil, fmt.Errorf("updating profile %s: %w", profile.Name, err)
	}

	return profile, nil
}

func (r *profileRepository) Delete(profile *models.Profile, force bool) error {
	if profile == nil {
		return nil
	}
	if profile.Name == DefaultProfileName {
		return ErrDefaultProfileDelete
	}

	state, err := r.GetState()
	if err != nil {
		return err
	}

	id := profile.ID

	if id == state.ActiveCommandProfileID || id == state.ActiveVariableProfileID || id == state.ActiveSettingsProfileID {
		return ErrProfileInUse
	}

	if !force {
		var commandCount, variableCount int64
		if err := r.db.Model(&models.Command{}).Where("profile_id = ?", id).Count(&commandCount).Error; err != nil {
			return fmt.Errorf("counting commands for profile id %d: %w", id, err)
		}
		if err := r.db.Model(&models.Variable{}).Where("profile_id = ?", id).Count(&variableCount).Error; err != nil {
			return fmt.Errorf("counting variables for profile id %d: %w", id, err)
		}
		if commandCount > 0 || variableCount > 0 {
			return fmt.Errorf("%w: %d command(s), %d variable(s)", ErrProfileHasContent, commandCount, variableCount)
		}
	}

	if err := r.db.Delete(&models.Profile{}, id).Error; err != nil {
		return fmt.Errorf("deleting profile id %d: %w", id, err)
	}
	return nil
}

func (r *profileRepository) RecordUse(profile *models.Profile) error {
	result := r.db.Model(&models.Profile{}).Where("id = ?", profile.ID).Update("last_used", time.Now())
	if result.Error != nil {
		return fmt.Errorf("touching last_used for profile %s: %w", profile.Name, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrProfileNotFound
	}
	return nil
}

func (r *profileRepository) GetState() (*models.ProfileState, error) {
	var state models.ProfileState
	// ProfileState is a singleton and never has a second row
	if err := r.db.First(&state, 1).Error; err != nil {
		return nil, fmt.Errorf("loading profile state: %w", err)
	}
	return &state, nil
}

func (r *profileRepository) GetStateWithProfiles() (*models.ProfileState, error) {
	var state models.ProfileState
	if err := r.db.
		Preload("ActiveCommandProfile").
		Preload("ActiveVariableProfile").
		Preload("ActiveSettingsProfile").
		First(&state, 1).Error; err != nil {
		return nil, fmt.Errorf("loading profile state with profiles: %w", err)
	}
	return &state, nil
}

func (r *profileRepository) SetActiveProfile(config SetProfileConfig) (*models.ProfileState, error) {
	state, err := r.GetState()
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if config.CommandProfile != nil {
		updates["active_command_profile_id"] = config.CommandProfile.ID
	}
	if config.VariableProfile != nil {
		updates["active_variable_profile_id"] = config.VariableProfile.ID
	}
	if config.SettingsProfile != nil {
		updates["active_settings_profile_id"] = config.SettingsProfile.ID
	}
	if len(updates) == 0 {
		return state, nil
	}

	if err := r.db.Model(&models.ProfileState{}).Where("id = ?", state.ID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("updating active profiles: %w", err)
	}

	return r.GetState()
}

func (r *profileRepository) GetActiveCommandProfile() (*models.Profile, error) {
	var state models.ProfileState
	if err := r.db.Preload("ActiveCommandProfile").First(&state, 1).Error; err != nil {
		return nil, fmt.Errorf("loading profile state: %w", err)
	}
	return &state.ActiveCommandProfile, nil
}

func (r *profileRepository) GetActiveVariableProfile() (*models.Profile, error) {
	var state models.ProfileState
	if err := r.db.Preload("ActiveVariableProfile").First(&state, 1).Error; err != nil {
		return nil, fmt.Errorf("loading profile state: %w", err)
	}
	return &state.ActiveVariableProfile, nil
}

func (r *profileRepository) GetActiveSettingsProfile() (*models.Profile, error) {
	var state models.ProfileState
	if err := r.db.Preload("ActiveSettingsProfile").First(&state, 1).Error; err != nil {
		return nil, fmt.Errorf("loading profile state: %w", err)
	}
	return &state.ActiveSettingsProfile, nil
}
