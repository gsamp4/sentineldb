package domain

import (
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type TriggerRepositoryInterface interface {
	RunTrigger() bool
    GetTrigger(id string) (*models.Run, error)
}

type TriggerRepository struct {
	DB *gorm.DB
	Logger *logger.Logger
}

func (r TriggerRepository) RunTrigger() bool {
    // passo 1 — busca assets ativos
    var assets []models.Asset
    if err := r.DB.Where("active = ?", true).Find(&assets).Error; err != nil {
        r.Logger.Error("Error fetching assets", err)
        return false
    }

    // passo 2 — tudo numa transação só
    err := r.DB.Transaction(func(tx *gorm.DB) error {

        // cria o run
        run := models.Run{
            ID:        ulid.Make().String(),
            Status:    "pending",
            CreatedAt: time.Now(),
        }
        if err := tx.Create(&run).Error; err != nil {
            return err  // rollback
        }

        // cria um job no outbox para cada asset
        for _, asset := range assets {
            job := models.Outbox{
                ID:          ulid.Make().String(),
                RunID:       run.ID,
                AssetID:     asset.ID,
                JobType:     "shodan_scan",
                Status:      "pending",
                ScheduledAt: time.Now(),
            }
            if err := tx.Create(&job).Error; err != nil {
                return err  // rollback
            }
        }

        return nil  // commit
    })

    if err != nil {
        r.Logger.Error("Error creating run and outbox jobs", err)
        return false
    }

    return true
}

func (r TriggerRepository) GetTrigger(id string) (*models.Run, error) {
    var run models.Run
    if err := r.DB.Where("id = ?", id).First(&run).Error; err != nil {
        return nil, err
    }
    return &run, nil
}