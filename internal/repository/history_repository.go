package repository

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HistoryEntry struct {
	Alias          string
	Template       string
	Resolved       string
	VariablesUsed  *string
	ExitCode       *int
	RetentionLimit *int
	ProfileID      *uint
}

type HistoryRepository interface {
	Record(entry HistoryEntry) (*models.CommandHistory, error)
	GetByID(id string, profileID *uint) (*models.CommandHistory, error)
	GetRecent(alias *string, limit int, profileID *uint) ([]models.CommandHistory, error)
	DeleteByID(id string, profileID *uint) error
	Clear(alias *string, profileID *uint) (int, error)
}

type historyRepository struct {
	db          *gorm.DB
	profileRepo ProfileRepository
}

func NewHistoryRepository(db *gorm.DB, profileRepo ProfileRepository) HistoryRepository {
	return &historyRepository{db: db, profileRepo: profileRepo}
}

func (r *historyRepository) Record(entry HistoryEntry) (*models.CommandHistory, error) {
	pid, err := resolveCommandProfileID(entry.ProfileID, r.profileRepo)
	if err != nil {
		return nil, err
	}

	rawID := uuid.New()
	hexID := hex.EncodeToString(rawID[:])

	ranAt := time.Now()

	commandHistory := &models.CommandHistory{
		ID:            hexID,
		Alias:         entry.Alias,
		Template:      entry.Template,
		Resolved:      entry.Resolved,
		VariablesUsed: entry.VariablesUsed,
		ExitCode:      entry.ExitCode,
		ProfileID:     pid,
		RanAt:         ranAt,
	}
	if err = r.db.Create(commandHistory).Error; err != nil {
		return nil, fmt.Errorf("creating command history entry: %w", err)
	}

	err = r.applyRetention(entry.Alias, entry.RetentionLimit, pid)
	if err != nil {
		return nil, err
	}
	return commandHistory, nil
}

// applyRetention applies retention logic by limiting the number of recent command history entries for the specified
// alias. Deletes older entries exceeding the given limit.
func (r *historyRepository) applyRetention(alias string, limit *int, profileID uint) error {
	if limit == nil || *limit <= 0 {
		return nil
	}

	var keepIDs []string
	result := r.db.Model(&models.CommandHistory{}).
		Where("alias = ? AND profile_id = ?", alias, profileID).
		Order("ran_at DESC").
		Limit(*limit).
		Pluck("id", &keepIDs)
	if result.Error != nil {
		return fmt.Errorf("gathering keep IDs: %w", result.Error)
	}

	if len(keepIDs) >= *limit {
		err := r.db.Where("alias = ? AND profile_id = ? AND id NOT IN (?)", alias, profileID, keepIDs).
			Delete(&models.CommandHistory{}).Error
		if err != nil {
			return fmt.Errorf("deleting old command history: %w", err)
		}
	}
	return nil
}

func (r *historyRepository) GetByID(id string, profileID *uint) (*models.CommandHistory, error) {
	pid, err := resolveCommandProfileID(profileID, r.profileRepo)
	if err != nil {
		return nil, err
	}

	var entry models.CommandHistory
	err = r.db.Where("id = ? AND profile_id = ?", id, pid).First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: id %q", ErrUnknownCommandHistory, id)
	}
	if err != nil {
		return nil, fmt.Errorf("fetching command history entry: %w", err)
	}
	return &entry, nil
}

func (r *historyRepository) GetRecent(alias *string, limit int, profileID *uint) ([]models.CommandHistory, error) {
	pid, err := resolveCommandProfileID(profileID, r.profileRepo)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 25
	}

	var entries []models.CommandHistory
	q := r.db.Where("profile_id = ?", pid)
	if alias != nil {
		q = q.Where("alias = ?", *alias)
	}
	result := q.Order("ran_at DESC").
		Limit(limit).
		Find(&entries)
	if result.Error != nil {
		return nil, fmt.Errorf("fetching recent command history: %w", result.Error)
	}
	return entries, nil
}

func (r *historyRepository) DeleteByID(id string, profileID *uint) error {
	entry, err := r.GetByID(id, profileID)
	if err != nil {
		return fmt.Errorf("fetching command history entry: %w", err)
	}
	if err = r.db.Delete(&entry).Error; err != nil {
		return fmt.Errorf("deleting command history entry: %w", err)
	}
	return nil
}

func (r *historyRepository) Clear(alias *string, profileID *uint) (int, error) {
	pid, err := resolveCommandProfileID(profileID, r.profileRepo)
	if err != nil {
		return 0, err
	}
	q := r.db.Where("profile_id = ?", pid)
	if alias != nil {
		q = q.Where("alias = ?", *alias)
	}
	result := q.Delete(&models.CommandHistory{})
	if result.Error != nil {
		return 0, fmt.Errorf("clearing command history: %w", result.Error)
	}
	return int(result.RowsAffected), nil
}
