package worker

import (
	"context"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"gorm.io/gorm"
)

func Dequeue(ctx context.Context, db *gorm.DB, log *logger.Logger) (*models.Outbox, error) {
	// Asynchronous job processing loop that listens for shutdown signals via the context
	var job models.Outbox

	err := db.WithContext(ctx).
		Raw(`
			SELECT id, run_id, asset_id, job_type, status, attempts, max_attempts, scheduled_at, updated_at, finished_at
			FROM outboxes
			WHERE status = 'pending'
			AND scheduled_at <= NOW()
			ORDER BY scheduled_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		`).Scan(&job).Error

		if err != nil {
			log.Error("Error fetching job from outboxes: ", err)
			return nil, err
		}

		if job.ID == "" {
			log.Info("No pending jobs found, worker is idle")
			return nil, nil
		}

		db.Model(&job).Updates(map[string]interface{}{
			"status": "processing",
			"updated_at": gorm.Expr("NOW()"),
		})
		return &job, nil
}