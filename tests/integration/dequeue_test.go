package integration_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"sentineldb/internal/job/models"
	"sentineldb/internal/worker"
)

func TestIntegration_DequeueOnlyReturnsJobToOneConcurrentWorker(t *testing.T) {
	db := newIntegrationDB(t)
	log := newTestLogger()

	asset := seedAsset(t, db, "asset-dequeue", "ip", "8.8.4.4")
	run := seedRun(t, db, "run-dequeue")
	seedOutbox(t, db, models.Outbox{
		ID:          "job-dequeue",
		RunID:       run.ID,
		AssetID:     asset.ID,
		JobType:     "internetdb_scan",
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		ScheduledAt: time.Now().Add(-1 * time.Minute),
		UpdatedAt:   time.Now(),
	})

	type result struct {
		job *models.Outbox
		err error
	}

	start := make(chan struct{})
	results := make(chan result, 2)
	var ready sync.WaitGroup
	ready.Add(2)

	for range 2 {
		go func() {
			ready.Done()
			<-start

			job, err := worker.Dequeue(context.Background(), db, log)
			results <- result{job: job, err: err}
		}()
	}

	ready.Wait()
	close(start)

	var receivedJobs int
	var nilResults int
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("Dequeue returned error: %v", result.err)
		}

		if result.job == nil {
			nilResults++
			continue
		}

		receivedJobs++
		if result.job.ID != "job-dequeue" {
			t.Fatalf("expected dequeued job ID job-dequeue, got %s", result.job.ID)
		}
	}

	if receivedJobs != 1 {
		t.Fatalf("expected exactly one worker to receive the job, got %d", receivedJobs)
	}

	if nilResults != 1 {
		t.Fatalf("expected exactly one worker to receive nil, got %d", nilResults)
	}

	var updated models.Outbox
	if err := db.First(&updated, "id = ?", "job-dequeue").Error; err != nil {
		t.Fatalf("failed to reload outbox job: %v", err)
	}

	if updated.Status != "processing" {
		t.Fatalf("expected job status processing after dequeue, got %s", updated.Status)
	}
}
