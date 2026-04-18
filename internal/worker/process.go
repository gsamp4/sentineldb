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
    case "scan_shodan":
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
       //handleFailure(db, job, err)
        return
    }

    db.Model(job).Updates(map[string]interface{}{
        "status":      "completed",
        "finished_at": time.Now(),
    })
}