package worker

import (
	"context"
	"fmt"
	"os"
	"sentineldb/pkg/logger"
	"time"

	"gorm.io/gorm"
)

func Run(ctx context.Context, db *gorm.DB, log *logger.Logger) {
	// Asynchronous job processing loop that listens for shutdown signals via the context
	workerLog := log.With(fmt.Sprintf("[worker pid=%d]", os.Getpid()))
	workerLog.Info("looking for new jobs...")
	for {
		select {
			case <-ctx.Done():
				workerLog.Info("worker received shutdown signal")
				return
			default:
				job, err := Dequeue(ctx, db, workerLog)
				if err != nil || job == nil {
					time.Sleep(2 * time.Second)
					continue
				}
				Process(ctx, db, workerLog, job)
		}
	}
}