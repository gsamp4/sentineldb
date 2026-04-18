package worker

import (
	"context"
	"sentineldb/pkg/logger"
	"time"

	"gorm.io/gorm"
)

func Run(ctx context.Context, db *gorm.DB, log *logger.Logger) {
	// Asynchronous job processing loop that listens for shutdown signals via the context
	for {
		select {
			case <-ctx.Done():
				log.Info("Worker received shutdown signal")
				return
			default:
				log.Info("Worker is processing jobs...")
				job, err := Dequeue(ctx, db, log)
				if err != nil || job == nil {
					time.Sleep(2 * time.Second)
					continue
				}
				Process(ctx, db, log, job)
		}
	}
}