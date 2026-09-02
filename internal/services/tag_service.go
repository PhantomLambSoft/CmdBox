package services

import (
	"errors"
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
)

type UpdateTagConfig struct {
	Name        *string
	Description *string
}

type TagService interface {
	CreateTag(name string, description *string) (*models.Tag, error)
	UpdateTag(name string, config *UpdateTagConfig) (*models.Tag, error)
	DeleteTag(name string) error
	GetTag(name string) (*models.Tag, error)
	GetTagOrNil(name string) (*models.Tag, error)
	GetTagById(id uint) (*models.Tag, error)
	ListTags(orderBy string, limit *int) ([]models.Tag, error)
	Search(query string, fields []string, limit *int) ([]models.Tag, error)
}

type tagService struct {
	tagRepo repository.TagRepository
}

func NewTagService(tagRepo repository.TagRepository) TagService {
	return &tagService{
		tagRepo: tagRepo,
	}
}

func (s *tagService) CreateTag(name string, description *string) (*models.Tag, error) {
	input := repository.TagCreateConfig{
		Name:        name,
		Description: description,
	}

	var tag *models.Tag
	tag, err := s.tagRepo.Create(input)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *tagService) UpdateTag(name string, config *UpdateTagConfig) (*models.Tag, error) {
	tag, err := s.GetTag(name)
	if err != nil {
		return nil, fmt.Errorf("getting tag: %w", err)
	}

	input := repository.TagUpdateConfig{
		Name:        config.Name,
		Description: config.Description,
	}

	updatedTag, err := s.tagRepo.Update(tag, input)
	if err != nil {
		return nil, fmt.Errorf("updating tag: %w", err)
	}

	return updatedTag, nil
}

func (s *tagService) DeleteTag(name string) error {
	tag, err := s.GetTag(name)
	if err != nil {
		return fmt.Errorf("getting tag: %w", err)
	}

	err = s.tagRepo.Delete(tag)
	if err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}

	return nil
}

func (s *tagService) GetTag(name string) (*models.Tag, error) {
	return s.tagRepo.GetByName(name)
}

func (s *tagService) GetTagOrNil(name string) (*models.Tag, error) {
	tag, err := s.tagRepo.GetByName(name)
	if err != nil {
		if errors.Is(err, repository.ErrUnknownTagName) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting tag: %w", err)
	}
	return tag, nil
}

func (s *tagService) GetTagById(id uint) (*models.Tag, error) {
	return s.tagRepo.GetByID(id)
}

func (s *tagService) ListTags(orderBy string, limit *int) ([]models.Tag, error) {
	resolvedLimit := 0
	if limit != nil {
		resolvedLimit = *limit
	}
	return s.tagRepo.ListAll(orderBy, resolvedLimit)
}

func (s *tagService) Search(query string, fields []string, limit *int) ([]models.Tag, error) {
	resolvedLimit := 0
	if limit != nil {
		resolvedLimit = *limit
	}
	return s.tagRepo.Search(query, fields, resolvedLimit)
}
