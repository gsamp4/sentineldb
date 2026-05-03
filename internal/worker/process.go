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

var internetDBLookup = services.CallInternetDB

func SetInternetDBLookupForTests(fn func(context.Context, string) (*services.InternetDBResponse, error)) func() {
	previous := internetDBLookup
	internetDBLookup = fn

	return func() {
		internetDBLookup = previous
	}
}

func ProcessJob(ctx context.Context, job models.Outbox, db *gorm.DB, log *logger.Logger) {
	log.Infof("job start id=%s type=%s try=%d", job.ID, job.JobType, job.Attempts+1)
	updates := make(map[string]interface{})

	var err error
	switch job.JobType {
	case "internetdb_scan":
		var result *services.InternetDBResponse
		result, err = internetDBLookup(ctx, job.Asset.Value)
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
		updates = validateFailedJobAttempts(job, updates)
	} else {
		updates["status"] = "completed"
		updates["finished_at"] = time.Now()
		log.Infof("job ok id=%s", job.ID)
	}

	if err := db.Model(&job).Updates(updates).Error; err != nil {
		log.Errorf("error updating job id=%s err=%v", job.ID, err)
		return
	}

	var count int64
	countResult := db.Model(&models.Outbox{}).
		Where("run_id = ? AND status NOT IN ?", job.RunID, []string{"completed", "failed"}).
		Count(&count)

	if countResult.Error != nil {
		log.Errorf("error selecting run jobs id=%s err=%v", job.ID, countResult.Error)
		return
	}

	if count == 0 {
		queryResult := db.Model(&models.Run{}).
			Where("id = ?", job.RunID).
			Updates(map[string]interface{}{
				"status":      "completed",
				"finished_at": time.Now(),
			})

		if queryResult.Error != nil {
			log.Errorf("error updating run id=%s err=%v", job.ID, queryResult.Error)
			return
		}
		log.Infof("run completed run_id=%s", job.RunID)
	}
}

func validateFailedJobAttempts(
	job models.Outbox,
	jobUpdates map[string]interface{},
) map[string]interface{} {
	nextAttempt := job.Attempts + 1

	if nextAttempt >= job.MaxAttempts {
		jobUpdates["status"] = "failed"
		jobUpdates["finished_at"] = time.Now()
		jobUpdates["attempts"] = nextAttempt
	} else {
		jobUpdates["status"] = "pending"
		jobUpdates["attempts"] = nextAttempt
		jobUpdates["scheduled_at"] = time.Now().Add(time.Duration(30*nextAttempt) * time.Second)
	}
	return jobUpdates
}
