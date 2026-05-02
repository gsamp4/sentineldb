package worker

import (
	"context"
	"fmt"
	"sentineldb/internal/job/models"
	"sentineldb/internal/services"
	"sentineldb/pkg/logger"
	"time"

	"gorm.io/gorm"
)

func Process(ctx context.Context, db *gorm.DB, log *logger.Logger, job *models.Outbox) {
    jobLog := log.With(fmt.Sprintf("[job_id=%s job_type=%s asset_id=%s]", job.ID, job.JobType, job.AssetID))
    var err error

    jobLog.Info("starting job processing")

    switch job.JobType {
    case "shodan_scan":
        err = services.ProcessInternetDB(ctx, db, jobLog, job)
    default:
        jobLog.Error("unknown job type", job.JobType)
        err = fmt.Errorf("unknown job type: %s", job.JobType)
    }

    if err != nil {
        nextAttempts := job.Attempts + 1
        updates := map[string]interface{}{
            "attempts":   nextAttempts,
            "updated_at": time.Now(),
        }

        if nextAttempts >= job.MaxAttempts {
            updates["status"] = "failed"
            updates["finished_at"] = time.Now()
            jobLog.Error("job failed permanently", err)
        } else {
            updates["status"] = "pending"
            updates["scheduled_at"] = time.Now().Add(30 * time.Second)
            jobLog.Warn("job failed, rescheduling", err)
        }

        if updateErr := db.Model(job).Updates(updates).Error; updateErr != nil {
            jobLog.Error("failed to update job status after processing error", updateErr)
        }
        return
    }

    if updateErr := db.Model(job).Updates(map[string]interface{}{
        "status":      "completed",
        "finished_at": time.Now(),
    }).Error; updateErr != nil {
        jobLog.Error("failed to mark job as completed", updateErr)
        return
    }

    jobLog.Info("job completed")
}