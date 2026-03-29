package domain

import (
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"gorm.io/gorm"
)

type RunRepositoryInterface interface {
	GetRunByID(id string) (*models.Run, error)
	ListRuns() ([]models.Run, error)
}

type RunRepository struct {
	DB *gorm.DB
	Logger *logger.Logger
}

func (r RunRepository) GetRunByID(id string) (*models.Run, error) {
	var run models.Run
	if err := r.DB.First(&run, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &run, nil
}

func (r RunRepository) ListRuns() ([]models.Run, error) {
	var runs []models.Run
	if err := r.DB.Find(&runs).Error; err != nil {
		return nil, err
	}
	return runs, nil
}