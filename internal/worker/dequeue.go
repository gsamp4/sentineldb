package worker

import (
	"context"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"gorm.io/gorm"
)

func Dequeue(ctx context.Context, db *gorm.DB, log *logger.Logger) (*models.Outbox, error) {
	// Asynchronous job processing loop that listens for shutdown signals via the context

	var job *models.Outbox

	err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.Raw(`
            SELECT * FROM public.outbox
            WHERE status = 'pending'
            FOR UPDATE SKIP LOCKED
            LIMIT 1
        `).Scan(job)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return nil
		}

		result = tx.Model(&models.Outbox{}).Where("id = ?", job.ID).Update("status", "processing")
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		log.Error("Error dequeuing job: %v", err)
		return nil, err
	}

	return job, err
}