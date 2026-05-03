package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sentineldb/internal/job/models"
	"sentineldb/internal/services"
	"sentineldb/internal/worker"
)

func TestIntegration_ProcessJobRetriesWithBackoffAfterFailure(t *testing.T) {
	db := newIntegrationDB(t)
	log := newTestLogger()

	asset := seedAsset(t, db, "asset-retry", "ip", "4.4.4.4")
	run := seedRun(t, db, "run-retry")
	initialSchedule := time.Now().Add(-1 * time.Minute)
	seedOutbox(t, db, models.Outbox{
		ID:          "job-retry",
		RunID:       run.ID,
		AssetID:     asset.ID,
		JobType:     "internetdb_scan",
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		ScheduledAt: initialSchedule,
		UpdatedAt:   time.Now(),
	})

	restoreLookup := worker.SetInternetDBLookupForTests(func(ctx context.Context, ip string) (*services.InternetDBResponse, error) {
		return nil, fmt.Errorf("forced internetdb failure for %s", ip)
	})
	t.Cleanup(restoreLookup)

	job, err := worker.Dequeue(context.Background(), db, log)
	if err != nil {
		t.Fatalf("Dequeue returned error: %v", err)
	}
	if job == nil {
		t.Fatal("expected a dequeued job, got nil")
	}

	processStartedAt := time.Now()
	worker.ProcessJob(context.Background(), *job, db, log)

	var retriedJob models.Outbox
	if err := db.First(&retriedJob, "id = ?", job.ID).Error; err != nil {
		t.Fatalf("failed to reload retried job: %v", err)
	}

	if retriedJob.Status != "pending" {
		t.Fatalf("expected failed job to return to pending, got %s", retriedJob.Status)
	}

	if retriedJob.Attempts != 1 {
		t.Fatalf("expected attempts to increment to 1, got %d", retriedJob.Attempts)
	}

	if retriedJob.FinishedAt != nil {
		t.Fatal("expected retried job to keep finished_at nil")
	}

	if !retriedJob.ScheduledAt.After(processStartedAt.Add(28 * time.Second)) {
		t.Fatalf("expected scheduled_at to move into the future, got %s", retriedJob.ScheduledAt)
	}

	var pendingRun models.Run
	if err := db.First(&pendingRun, "id = ?", run.ID).Error; err != nil {
		t.Fatalf("failed to reload run: %v", err)
	}

	if pendingRun.Status != "pending" {
		t.Fatalf("expected run to remain pending after retryable failure, got %s", pendingRun.Status)
	}

	if pendingRun.FinishedAt != nil {
		t.Fatal("expected pending run to keep finished_at nil")
	}
}
