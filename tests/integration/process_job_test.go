package integration_test

import (
	"context"
	"testing"
	"time"

	"sentineldb/internal/job/models"
	"sentineldb/internal/services"
	"sentineldb/internal/worker"
)

func TestIntegration_ProcessJobCompletesJobLifecycleAndRun(t *testing.T) {
	db := newIntegrationDB(t)
	log := newTestLogger()

	asset := seedAsset(t, db, "asset-process", "ip", "9.9.9.9")
	run := seedRun(t, db, "run-process")
	seedOutbox(t, db, models.Outbox{
		ID:          "job-process",
		RunID:       run.ID,
		AssetID:     asset.ID,
		JobType:     "internetdb_scan",
		Status:      "pending",
		Attempts:    0,
		MaxAttempts: 3,
		ScheduledAt: time.Now().Add(-1 * time.Minute),
		UpdatedAt:   time.Now(),
	})

	restoreLookup := worker.SetInternetDBLookupForTests(func(ctx context.Context, ip string) (*services.InternetDBResponse, error) {
		return &services.InternetDBResponse{
			IP:        ip,
			Ports:     []int{443},
			Hostnames: []string{"dns.google"},
			Tags:      []string{"resolver"},
			Vulns:     []string{},
			CPEs:      []string{},
		}, nil
	})
	t.Cleanup(restoreLookup)

	job, err := worker.Dequeue(context.Background(), db, log)
	if err != nil {
		t.Fatalf("Dequeue returned error: %v", err)
	}
	if job == nil {
		t.Fatal("expected a dequeued job, got nil")
	}

	var processingJob models.Outbox
	if err := db.First(&processingJob, "id = ?", job.ID).Error; err != nil {
		t.Fatalf("failed to reload processing job: %v", err)
	}

	if processingJob.Status != "processing" {
		t.Fatalf("expected processing status after dequeue, got %s", processingJob.Status)
	}

	worker.ProcessJob(context.Background(), *job, db, log)

	var completedJob models.Outbox
	if err := db.Preload("Asset").First(&completedJob, "id = ?", job.ID).Error; err != nil {
		t.Fatalf("failed to reload completed job: %v", err)
	}

	if completedJob.Status != "completed" {
		t.Fatalf("expected completed job, got %s", completedJob.Status)
	}

	if completedJob.FinishedAt == nil {
		t.Fatal("expected completed job to have finished_at set")
	}

	var completedRun models.Run
	if err := db.First(&completedRun, "id = ?", run.ID).Error; err != nil {
		t.Fatalf("failed to reload run: %v", err)
	}

	if completedRun.Status != "completed" {
		t.Fatalf("expected completed run, got %s", completedRun.Status)
	}

	if completedRun.FinishedAt == nil {
		t.Fatal("expected completed run to have finished_at set")
	}

	var snapshotCount int64
	if err := db.Model(&models.AssetSnapshot{}).Where("run_id = ?", run.ID).Count(&snapshotCount).Error; err != nil {
		t.Fatalf("failed to count snapshots: %v", err)
	}

	if snapshotCount != 1 {
		t.Fatalf("expected 1 snapshot after successful processing, got %d", snapshotCount)
	}
}
