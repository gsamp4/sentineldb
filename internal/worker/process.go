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

func ProcessJob(ctx context.Context, job models.Outbox, db *gorm.DB, log *logger.Logger) {
	log.Infof("job start id=%s type=%s try=%d", job.ID, job.JobType, job.Attempts+1)
	updates := make(map[string]interface{})

	var err error
	switch job.JobType {
	case "internetdb_scan":
		var result *services.InternetDBResponse
		service := services.NewInternetDBService("", nil)
		result, err = service.LookupByIP(ctx, job.Asset.Value)
		if err == nil {
			log.Infof("internetdb ok id=%s asset=%s", job.ID, job.AssetID)
		}

		if err == nil {
			err = processInternetDB(ctx, db, log, job, result)
		}

	default:
		err = fmt.Errorf("job type not implemented: %s", job.JobType)
	}

	if err != nil {
		log.Warnf("job fail id=%s err=%v", job.ID, err)
		updates = validateFailAttempts(job, updates)
	} else {
		updates["status"] = "completed"
		updates["finished_at"] = time.Now()
		log.Infof("job ok id=%s", job.ID)
	}

	if err := db.Model(&job).Updates(updates).Error; err != nil {
		log.Errorf("error updating job id=%s err=%v", job.ID, err)
		return
	}
}

func validateFailAttempts(job models.Outbox, updates map[string]interface{}) map[string]interface{} {
	nextAttempt := job.Attempts + 1

	if nextAttempt >= job.MaxAttempts {
		updates["status"] = "failed"
		updates["attempts"] = nextAttempt
		updates["finished_at"] = time.Now()
	} else {
		updates["status"] = "pending"
		updates["attempts"] = nextAttempt
		updates["scheduled_at"] = time.Now().Add(time.Duration(30*nextAttempt) * time.Second)
	}
	return updates
}
