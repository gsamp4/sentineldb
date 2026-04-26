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
    var err error

    switch job.JobType {
    case "shodan_scan":
        err = services.ProcessShodan(ctx, db, log, job)
    // case "scan_hibp":
    //     err = processHIBP(ctx, db, log, job)
    // case "correlate":
    //     err = processCorrelate(ctx, db, log, job)
    default:
        log.Error("unknown job type: ", job.JobType)
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
            log.Error("job failed permanently: ", err)
        } else {
            updates["status"] = "pending"
            updates["scheduled_at"] = time.Now().Add(30 * time.Second)
            log.Error("job failed, rescheduling: ", err)
        }

        db.Model(job).Updates(updates)
        return
    }

    db.Model(job).Updates(map[string]interface{}{
        "status":      "completed",
        "finished_at": time.Now(),
    })
}