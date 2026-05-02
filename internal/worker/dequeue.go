package worker

import (
	"context"
	"fmt"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Dequeue(ctx context.Context, db *gorm.DB, log *logger.Logger) (*models.Outbox, error) {
	// Asynchronous job processing loop that listens for shutdown signals via the context
	var job models.Outbox

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", "pending").
			Where("scheduled_at <= NOW()").
			Order("scheduled_at ASC").
			Take(&job)

		if query.Error != nil {
			if query.Error == gorm.ErrRecordNotFound {
				return nil
			}

			return query.Error
		}

		return tx.Model(&job).Updates(map[string]interface{}{
			"status":     "processing",
			"updated_at": gorm.Expr("NOW()"),
		}).Error
	})

	if err != nil {
		log.Error("Error fetching job from outboxes: ", err)
		return nil, err
	}

	if job.ID == "" {
		return nil, nil
	}

	log.Info(
		"job claimed from queue",
		fmt.Sprintf("job_id=%s job_type=%s asset_id=%s run_id=%s attempts=%d", job.ID, job.JobType, job.AssetID, job.RunID, job.Attempts),
	)

	return &job, nil
}