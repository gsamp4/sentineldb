package domain

import (
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"gorm.io/gorm"
)

type FindingRepositoryInterface interface {
	GetFindingByID(id string) (*models.Finding, error)
	ListFindings() ([]models.Finding, error)
}

type FindingRepository struct {
	DB *gorm.DB
	Logger *logger.Logger
}

func (f FindingRepository) ListFindings() ([]models.Finding, error) {
	var findings []models.Finding
	f.Logger.Info("[FINDINGS] Fetching all findings")
	if err := f.DB.Find(&findings).Error; err != nil {
		return nil, err
	}
	return findings, nil
}

func (f FindingRepository) GetFindingByID(id string) (*models.Finding, error) {
	var finding models.Finding
	f.Logger.Info("[FINDINGS] Fetching finding with ID: %s", id)
	if err := f.DB.First(&finding, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &finding, nil
}

func UpdateFindingStatus(db *gorm.DB, findingID string, status string) error {
	return db.Model(&models.Finding{}).Where("id = ?", findingID).Update("status", status).Error
}