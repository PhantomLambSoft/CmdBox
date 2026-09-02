package repository

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
)

const DefaultProfileName = "default"

type ProfileRepository interface {
	WithTx(tx *gorm.DB) ProfileRepository

	Create(name string, description *string) (*models.Profile, error)
	GetByName(name string) (*models.Profile, error)
	GetByID(id uint) (*models.Profile, error)
	ListAll() ([]models.Profile, error)
	Update(id uint, name *string, description *string) (*models.Profile, error)
	Delete(id uint, force bool) error
	RecordUse(id uint) error

	GetState() (*models.ProfileState, error)
	SetActiveProfile(commandProfileID, variableProfileID, settingsProfileID *uint) (*models.ProfileState, error)

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

func (r *profileRepository) ListAll() ([]models.Profile, error) {
	var profiles []models.Profile
	if err := r.db.Order("name").Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("listing profiles: %w", err)
	}
	return profiles, nil
}

func (r *profileRepository) Update(id uint, name *string, description *string) (*models.Profile, error) {
	profile, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := r.validator.ValidateUpdate(name); err != nil {
		return nil, err
	}

	if name != nil && *name != profile.Name {
		if profile.Name == DefaultProfileName {
			return nil, ErrDefaultProfileName
		}
		if _, err := r.GetByName(*name); err == nil {
			return nil, ErrProfileNameExists
		} else if !errors.Is(err, ErrProfileNotFound) {
			return nil, err
		}
		profile.Name = *name
	}

	if description != nil {
		profile.Description = description
	}

	if err := r.db.Save(profile).Error; err != nil {
		return nil, fmt.Errorf("updating profile id %d: %w", id, err)
	}

	return profile, nil
}

func (r *profileRepository) Delete(id uint, force bool) error {
	profile, err := r.GetByID(id)
	if err != nil {
		return err
	}
	if profile.Name == DefaultProfileName {
		return ErrDefaultProfileDelete
	}

	state, err := r.GetState()
	if err != nil {
		return err
	}
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

func (r *profileRepository) RecordUse(id uint) error {
	result := r.db.Model(&models.Profile{}).Where("id = ?", id).Update("last_used", time.Now())
	if result.Error != nil {
		return fmt.Errorf("touching last_used for profile id %d: %w", id, result.Error)
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

func (r *profileRepository) SetActiveProfile(commandProfileID, variableProfileID, settingsProfileID *uint) (*models.ProfileState, error) {
	state, err := r.GetState()
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if commandProfileID != nil {
		updates["active_command_profile_id"] = *commandProfileID
	}
	if variableProfileID != nil {
		updates["active_variable_profile_id"] = *variableProfileID
	}
	if settingsProfileID != nil {
		updates["active_settings_profile_id"] = *settingsProfileID
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
