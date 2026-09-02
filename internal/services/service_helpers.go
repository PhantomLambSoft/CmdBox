package services

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
)

// getTags retrieves a list of tags by their names using the supplied tag repository and
// returns them or an error if any occurs.
func getTags(tagNames []string, tagRepo repository.TagRepository) ([]models.Tag, error) {
	if len(tagNames) <= 0 {
		return nil, nil
	}
	var tags []models.Tag
	for _, name := range tagNames {
		tag, err := tagRepo.GetByName(name)
		if err != nil {
			return nil, fmt.Errorf("getting tag by name: %w", err)
		}
		tags = append(tags, *tag)
	}
	return tags, nil
}
