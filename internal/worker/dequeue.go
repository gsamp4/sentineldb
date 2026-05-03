package worker

import (
	"context"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Dequeue(ctx context.Context, db *gorm.DB, log *logger.Logger) (*models.Outbox, error) {
	// Asynchronous job processing loop that listens for shutdown signals via the context

	job := &models.Outbox{}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Preload("Asset").
			Where("status = ? AND scheduled_at <= NOW()", "pending").
			Order("scheduled_at ASC").
			First(job).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				job = nil
				return nil
			}

			return err
		}

		if err := tx.Model(&models.Outbox{}).
			Where("id = ?", job.ID).
			Updates(map[string]interface{}{"status": "processing", "updated_at": gorm.Expr("NOW()")}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Error("Error dequeuing job: %v", err)
		return nil, err
	}

	return job, err
}
