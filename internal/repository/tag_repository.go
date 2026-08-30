package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
	"gorm.io/gorm"
)

var tagOrderableColumns = map[string]bool{
	"id":          true,
	"name":        true,
	"description": true,
	"created_at":  true,
	"updated_at":  true,
}

var defaultTagSearchFields = []string{"name", "description"}

type TagCreateConfig struct {
	Name        string
	Description *string
}

type TagUpdateConfig struct {
	Name        *string
	Description *string
}

type TagRepository interface {
	Create(input TagCreateConfig) (*models.Tag, error)
	GetByName(name string) (*models.Tag, error)
	GetByID(ID uint) (*models.Tag, error)
	Update(tag *models.Tag, input TagUpdateConfig) (*models.Tag, error)
	Delete(command *models.Tag) error

	ListAll(orderBy string, limit int) ([]models.Tag, error)
	Search(query string, fields []string, limit int) ([]models.Tag, error)
}

type tagRepository struct {
	db        *gorm.DB
	validator *validate.TagValidator
}

func NewTagRepository(db *gorm.DB, validator *validate.TagValidator) TagRepository {
	if validator == nil {
		validator = validate.NewTagValidator(nil)
	}
	return &tagRepository{db: db, validator: validator}
}

func (r *tagRepository) Create(input TagCreateConfig) (*models.Tag, error) {
	name := strings.TrimSpace(input.Name)
	if err := r.validator.ValidateCreate(name); err != nil {
		return nil, err
	}

	tag := &models.Tag{
		Name:        name,
		Description: input.Description,
	}

	if err := r.db.Create(tag).Error; err != nil {
		if isUniqueConstraintViolation(err, "tags", "name") {
			return nil, fmt.Errorf("%w: %q", ErrTagNameConflict, name)
		}
		return nil, fmt.Errorf("creating tag: %q: %w", name, err)
	}
	return tag, nil
}

func (r *tagRepository) GetByName(name string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %q", ErrUnknownTagName, name)
	}
	if err != nil {
		return nil, fmt.Errorf("getting tag by name: %q: %w", name, err)
	}
	return &tag, nil
}

func (r *tagRepository) GetByID(ID uint) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.Where("id = ?", ID).First(&tag).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %d", ErrUnknownTag, ID)
	}
	if err != nil {
		return nil, fmt.Errorf("getting tag by id: %d: %w", ID, err)
	}
	return &tag, nil
}

func (r *tagRepository) Update(tag *models.Tag, input TagUpdateConfig) (*models.Tag, error) {
	if tag == nil {
		return nil, ErrNoUpdateTarget
	}

	mergedName := tag.Name
	if input.Name != nil {
		mergedName = strings.TrimSpace(*input.Name)
	}

	if err := r.validator.ValidateUpdate(&mergedName); err != nil {
		return nil, err
	}

	if input.Name != nil {
		tag.Name = mergedName
	}
	if input.Description != nil {
		tag.Description = input.Description
	}

	if err := r.db.Save(tag).Error; err != nil {
		if isUniqueConstraintViolation(err, "tags", "name") {
			return nil, fmt.Errorf("%w: %q", ErrTagNameConflict, mergedName)
		}
		return nil, fmt.Errorf("updating tag: %d: %w", tag.ID, err)
	}
	return tag, nil
}

func (r *tagRepository) Delete(tag *models.Tag) error {
	if tag == nil {
		return nil
	}
	if err := r.db.Delete(tag).Error; err != nil {
		return fmt.Errorf("deleting tag: %d: %w", tag.ID, err)
	}
	return nil
}

func (r *tagRepository) ListAll(orderBy string, limit int) ([]models.Tag, error) {
	if orderBy == "" {
		orderBy = "name"
	}
	if limit <= 0 {
		limit = 25
	}
	orderClauses, err := resolveOrdering(orderBy, tagOrderableColumns)
	if err != nil {
		return nil, err
	}

	var tags []models.Tag
	q := r.db.Model(&models.Tag{})
	for _, clause := range orderClauses {
		q = q.Order(clause)
	}
	if err := q.Limit(limit).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}
	return tags, nil
}

func (r *tagRepository) Search(query string, fields []string, limit int) ([]models.Tag, error) {
	if len(fields) == 0 {
		fields = defaultTagSearchFields
	}
	if limit <= 0 {
		limit = 25
	}

	var tags []models.Tag
	searchInput := searchWithRelevanceInput{
		DB:                   r.db,
		Table:                "tags",
		Query:                query,
		Fields:               fields,
		AllowedColumns:       tagOrderableColumns,
		SecondaryOrderColumn: "name",
		Limit:                limit,
		ExtraWhere:           "",
		ExtraArgs:            nil,
		Dest:                 &tags,
	}
	err := searchWithRelevance(searchInput)
	if err != nil {
		return nil, fmt.Errorf("searching tags: %w", err)
	}
	return tags, nil
}
