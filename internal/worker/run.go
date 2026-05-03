package worker

import (
	"context"
	"sentineldb/pkg/logger"

	"gorm.io/gorm"
)

type NewWorkerType struct {
	DB *gorm.DB
	Log *logger.Logger
}

func NewWorker(db *gorm.DB, log *logger.Logger) *NewWorkerType {
	return &NewWorkerType{
		DB: db,
		Log: log,
	}
}

func (w *NewWorkerType) Run(ctx context.Context,) {
	// Asynchronous job processing loop that listens for shutdown signals via the context
	w.Log.Info("looking for new jobs...")
	for {
		select {
			case <-ctx.Done():
				// Em caso de cancelamento do contexto, loga a mensagem e sai do loop
				w.Log.Info("worker received shutdown signal, exiting...")
				return

			default:
				job, err := Dequeue(ctx, w.DB, w.Log)
				if err != nil {
					w.Log.Error("Error dequeuing job: %v", err)
					continue
				}
				if job != nil {
					w.Log.Info("Processing job ID: %s, Type: %s", job.ID, job.JobType)
					ProcessJob(ctx, *job, w.DB, w.Log)
				}
		}
	}
}