package domain

import (
	"errors"
	"strings"
	"time"

	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

var (
	ErrTriggerAssetNotFound     = errors.New("asset not found")
	ErrTriggerUnsupportedAsset  = errors.New("asset type not supported")
	ErrTriggerNoSupportedAssets = errors.New("no supported active assets found")
)

type TriggerResult struct {
	RunID    string `json:"run_id"`
	JobCount int    `json:"job_count"`
}

type TriggerRepositoryInterface interface {
	TriggerAllActiveAssets() (*TriggerResult, error)
	TriggerAssetByID(id string) (*TriggerResult, error)
}

type TriggerRepository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func (r TriggerRepository) TriggerAllActiveAssets() (*TriggerResult, error) {
	var assets []models.Asset
	if err := r.DB.Where("active = ?", true).Find(&assets).Error; err != nil {
		return nil, err
	}

	supportedAssets := filterTriggerableAssets(assets)
	if len(supportedAssets) == 0 {
		return nil, ErrTriggerNoSupportedAssets
	}

	return r.createRunWithJobs(supportedAssets)
}

func (r TriggerRepository) TriggerAssetByID(id string) (*TriggerResult, error) {
	var asset models.Asset
	if err := r.DB.First(&asset, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTriggerAssetNotFound
		}
		return nil, err
	}

	if !asset.Active {
		return nil, ErrTriggerAssetNotFound
	}

	if !isTriggerableAsset(asset) {
		return nil, ErrTriggerUnsupportedAsset
	}

	return r.createRunWithJobs([]models.Asset{asset})
}

func (r TriggerRepository) createRunWithJobs(assets []models.Asset) (*TriggerResult, error) {
	runID := ulid.Make().String()
	now := time.Now()

	jobs := make([]models.Outbox, 0, len(assets))
	for _, asset := range assets {
		jobType, ok := triggerJobTypeForAsset(asset)
		if !ok {
			continue
		}

		jobs = append(jobs, models.Outbox{
			ID:          ulid.Make().String(),
			RunID:       runID,
			AssetID:     asset.ID,
			JobType:     jobType,
			Status:      "pending",
			Attempts:    0,
			MaxAttempts: 3,
			ScheduledAt: now,
			UpdatedAt:   now,
		})
	}

	if len(jobs) == 0 {
		return nil, ErrTriggerNoSupportedAssets
	}

	run := models.Run{
		ID:        runID,
		CreatedAt: now,
		Status:    "pending",
	}

	if err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&run).Error; err != nil {
			return err
		}

		if err := tx.Create(&jobs).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	r.Logger.Infof("[TRIGGER] - created run %s with %d jobs", runID, len(jobs))

	return &TriggerResult{
		RunID:    runID,
		JobCount: len(jobs),
	}, nil
}

func filterTriggerableAssets(assets []models.Asset) []models.Asset {
	filtered := make([]models.Asset, 0, len(assets))
	for _, asset := range assets {
		if isTriggerableAsset(asset) {
			filtered = append(filtered, asset)
		}
	}

	return filtered
}

func isTriggerableAsset(asset models.Asset) bool {
	_, ok := triggerJobTypeForAsset(asset)
	return ok
}

func triggerJobTypeForAsset(asset models.Asset) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(asset.Type)) {
	case "ip":
		return "internetdb_scan", true
	default:
		return "", false
	}
}
