package domain

import (
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"gorm.io/gorm"
)

type RunRepositoryInterface interface {
	GetRunByID(id string) (*models.Run, error)
	ListRuns() ([]models.Run, error)
	GetRunJobs(id string) ([]models.Outbox, error)
}

type RunRepository struct {
	DB *gorm.DB
	Logger *logger.Logger
}

func (r RunRepository) GetRunByID(id string) (*models.Run, error) {
	var run models.Run
	r.Logger.Info("[RUNS] - Fetching run with ID: %s", id)
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
	r.Logger.Info("[RUNS] - Fetching all runs")
	if err := r.DB.Find(&runs).Error; err != nil {
		return nil, err
	}
	return runs, nil
}

func (r RunRepository) GetRunJobs(id string) ([]models.Outbox, error) {
	var jobs []models.Outbox
	r.Logger.Info("[RUNS] - Fetching jobs for run with ID: %s", id)
	if err := r.DB.Where("run_id = ?", id).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}