package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
	"gorm.io/gorm"
)

var variableOrderableColumns = map[string]bool{
	"id":         true,
	"name":       true,
	"value":      true,
	"created_at": true,
	"updated_at": true,
	"profile_id": true,
}

var defaultVariableSearchFields = []string{"name", "value"}

type VariableCreateConfig struct {
	Name      string
	Value     string
	ProfileID *uint
}

type VariableUpdateConfig struct {
	Name  *string
	Value *string
}

type VariableRepository interface {
	Create(input VariableCreateConfig) (*models.Variable, error)
	GetByName(name string, profileID *uint) (*models.Variable, error)
	GetByID(id uint, profileID *uint) (*models.Variable, error)
	Update(variable *models.Variable, input VariableUpdateConfig) (*models.Variable, error)
	Delete(variable *models.Variable) error

	ListAll(orderBy string, limit int, profileID *uint) ([]models.Variable, error)
	ListByTags(tag []models.Tag, orderBy string, limit int, profileID *uint) ([]models.Variable, error)
	Search(query string, fields []string, limit int, profileID *uint) ([]models.Variable, error)

	AddTags(variable *models.Variable, tags []models.Tag) (TagAttachResult, error)
	RemoveTags(variable *models.Variable, tags []models.Tag) (TagDetachResult, error)
}

type variableRepository struct {
	db          *gorm.DB
	profileRepo ProfileRepository
	validator   *validate.VariableValidator
}

func NewVariableRepository(db *gorm.DB, profileRepo ProfileRepository, validator *validate.VariableValidator) VariableRepository {
	if validator == nil {
		validator = validate.NewVariableValidator(nil)
	}
	return &variableRepository{db: db, profileRepo: profileRepo, validator: validator}
}

func (r *variableRepository) Create(input VariableCreateConfig) (*models.Variable, error) {
	name := strings.TrimSpace(input.Name)
	if err := r.validator.ValidateCreate(name, input.Value); err != nil {
		return nil, err
	}

	profileID, err := resolveVariableProfileID(input.ProfileID, r.profileRepo)
	if err != nil {
		return nil, err
	}

	variable := &models.Variable{
		Name:      name,
		Value:     input.Value,
		ProfileID: profileID,
	}

	if err := r.db.Create(variable).Error; err != nil {
		if isUniqueConstraintViolation(err, "variables", "name") {
			return nil, fmt.Errorf("%w: %q", ErrNameConflict, name)
		}
		return nil, fmt.Errorf("creating variable %q: %w", name, err)
	}
	return variable, nil
}

func (r *variableRepository) GetByName(name string, profileID *uint) (*models.Variable, error) {
	pid, err := resolveVariableProfileID(profileID, r.profileRepo)
	if err != nil {
		return nil, err
	}

	var variable models.Variable
	err = r.db.Where("name = ? AND profile_id = ?", name, pid).First(&variable).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %q", ErrUnKnownName, err)
	}
	if err != nil {
		return nil, fmt.Errorf("looking up variable name: %q: %w", name, err)
	}
	return &variable, nil
}

func (r *variableRepository) GetByID(id uint, profileID *uint) (*models.Variable, error) {
	pid, err := resolveVariableProfileID(profileID, r.profileRepo)
	if err != nil {
		return nil, err
	}

	var variable models.Variable
	err = r.db.Where("id = ? AND profile_id = ?", id, pid).First(&variable).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: id %q", ErrUnknownVariable, err)
	}
	if err != nil {
		return nil, fmt.Errorf("looking up variable id: %d: %w", id, err)
	}
	return &variable, nil
}

func (r *variableRepository) Update(variable *models.Variable, input VariableUpdateConfig) (*models.Variable, error) {
	if variable == nil {
		return nil, ErrNoUpdateTarget
	}

	mergedName := variable.Name
	if input.Name != nil {
		mergedName = strings.TrimSpace(*input.Name)
	}
	mergedValue := variable.Value
	if input.Value != nil {
		mergedValue = strings.TrimSpace(*input.Value)
	}

	if err := r.validator.ValidateUpdate(&mergedName, &mergedValue); err != nil {
		return nil, err
	}

	if input.Name != nil {
		variable.Name = mergedName
	}
	if input.Value != nil {
		variable.Value = mergedValue
	}

	if err := r.db.Save(variable).Error; err != nil {
		if isUniqueConstraintViolation(err, "variables", "name") {
			return nil, fmt.Errorf("%w: %q", ErrNameConflict, mergedName)
		}
		return nil, fmt.Errorf("updating variable id: %d: %w", variable.ID, err)
	}
	return variable, nil
}

func (r *variableRepository) Delete(variable *models.Variable) error {
	if variable == nil {
		return nil
	}
	if err := r.db.Delete(variable).Error; err != nil {
		return fmt.Errorf("deleting variable id: %d: %w", variable.ID, err)
	}
	return nil
}

func (r *variableRepository) AddTags(variable *models.Variable, tags []models.Tag) (TagAttachResult, error) {
	if len(tags) == 0 {
		return TagAttachResult{}, nil
	}

	var added, existing []string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, tag := range tags {
			var link models.VariableTag
			err := tx.Where("variable_id = ? AND tag_id = ?", variable.ID, tag.ID).First(&link).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				newLink := models.VariableTag{VariableID: variable.ID, TagID: tag.ID}
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

func (r *variableRepository) RemoveTags(variable *models.Variable, tags []models.Tag) (TagDetachResult, error) {
	if len(tags) == 0 {
		return TagDetachResult{}, nil
	}

	var removed, notAttached []string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, tag := range tags {
			result := tx.Where("variable_id = ? AND tag_id = ?", variable.ID, tag.ID).Delete(&models.VariableTag{})
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

func (r *variableRepository) ListAll(orderBy string, limit int, profileID *uint) ([]models.Variable, error) {
	if orderBy == "" {
		orderBy = "name"
	}
	if limit <= 0 {
		limit = 25
	}
	orderClauses, err := resolveOrdering(orderBy, variableOrderableColumns)
	if err != nil {
		return nil, err
	}
	pid, err := resolveVariableProfileID(profileID, r.profileRepo)
	if err != nil {
		return nil, err
	}

	var variables []models.Variable
	q := r.db.Where("profile_id = ?", pid)
	for _, clause := range orderClauses {
		q = q.Order(clause)
	}
	if err := q.Limit(limit).Find(&variables).Error; err != nil {
		return nil, fmt.Errorf("listing variables: %w", err)
	}
	return variables, nil
}

func (r *variableRepository) ListByTags(tags []models.Tag, orderBy string, limit int, profileID *uint) ([]models.Variable, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	if orderBy == "" {
		orderBy = "name"
	}
	if limit <= 0 {
		limit = 25
	}

	orderClauses, err := resolveOrdering(orderBy, variableOrderableColumns)
	if err != nil {
		return nil, err
	}
	pid, err := resolveVariableProfileID(profileID, r.profileRepo)
	if err != nil {
		return nil, err
	}

	tagIDs := make([]uint, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}

	var variables []models.Variable
	q := r.db.Distinct("variables.*").
		Joins("JOIN variable_tags ON variable_tags.variable_id = variables.id").
		Where("variable_tags.tag_id IN ? AND variables.profile_id = ?", tagIDs, pid)
	for _, clause := range orderClauses {
		q = q.Order(clause)
	}

	if err := q.Limit(limit).Find(&variables).Error; err != nil {
		return nil, fmt.Errorf("listing variables by tag: %w", err)
	}
	return variables, nil
}

func (r *variableRepository) Search(query string, fields []string, limit int, profileID *uint) ([]models.Variable, error) {
	if len(fields) == 0 {
		fields = defaultVariableSearchFields
	}
	if limit <= 0 {
		limit = 25
	}
	pid, err := resolveVariableProfileID(profileID, r.profileRepo)
	if err != nil {
		return nil, err
	}

	var variables []models.Variable
	searchInput := searchWithRelevanceInput{
		DB:                   r.db,
		Table:                "variables",
		Query:                query,
		Fields:               fields,
		AllowedColumns:       variableOrderableColumns,
		SecondaryOrderColumn: "name",
		Limit:                limit,
		ExtraWhere:           "profile_id = ?",
		ExtraArgs:            []interface{}{pid},
		Dest:                 &variables,
	}
	err = searchWithRelevance(searchInput)
	if err != nil {
		return nil, fmt.Errorf("searching variables: %w", err)
	}
	return variables, nil
}
