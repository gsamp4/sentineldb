package integration_test

import (
	"testing"

	"sentineldb/internal/job/domain"
	"sentineldb/internal/job/models"
)

func TestIntegration_TriggerAllActiveAssetsPersistsRunAndOutboxAtomically(t *testing.T) {
	db := newIntegrationDB(t)
	log := newTestLogger()

	seedAsset(t, db, "asset-trigger-ok", "ip", "8.8.8.8")

	repo := domain.TriggerRepository{DB: db, Logger: log}

	result, err := repo.TriggerAllActiveAssets()
	if err != nil {
		t.Fatalf("TriggerAllActiveAssets returned error: %v", err)
	}

	if result.JobCount != 1 {
		t.Fatalf("expected 1 job, got %d", result.JobCount)
	}

	var run models.Run
	if err := db.First(&run, "id = ?", result.RunID).Error; err != nil {
		t.Fatalf("expected persisted run, got error: %v", err)
	}

	var jobs []models.Outbox
	if err := db.Where("run_id = ?", result.RunID).Find(&jobs).Error; err != nil {
		t.Fatalf("failed to query outboxes: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 outbox job, got %d", len(jobs))
	}

	if jobs[0].RunID != run.ID {
		t.Fatalf("expected outbox run_id %s, got %s", run.ID, jobs[0].RunID)
	}

	if jobs[0].Status != "pending" {
		t.Fatalf("expected pending job, got %s", jobs[0].Status)
	}
}

func TestIntegration_TriggerAllActiveAssetsRollsBackWhenOutboxInsertFails(t *testing.T) {
	db := newIntegrationDB(t)
	log := newTestLogger()

	seedAsset(t, db, "asset-trigger-fail", "ip", "1.1.1.1")

	mustExec(t, db, "DROP TRIGGER IF EXISTS fail_outbox_insert ON outboxes")
	mustExec(t, db, "DROP FUNCTION IF EXISTS fail_outbox_insert()")
	mustExec(t, db, `
		CREATE FUNCTION fail_outbox_insert() RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'forced outbox failure';
		END;
		$$ LANGUAGE plpgsql
	`)
	mustExec(t, db, `
		CREATE TRIGGER fail_outbox_insert
		BEFORE INSERT ON outboxes
		FOR EACH ROW
		EXECUTE FUNCTION fail_outbox_insert()
	`)

	t.Cleanup(func() {
		_ = db.Exec("DROP TRIGGER IF EXISTS fail_outbox_insert ON outboxes").Error
		_ = db.Exec("DROP FUNCTION IF EXISTS fail_outbox_insert()").Error
	})

	repo := domain.TriggerRepository{DB: db, Logger: log}

	if _, err := repo.TriggerAllActiveAssets(); err == nil {
		t.Fatal("expected TriggerAllActiveAssets to fail when outbox insert trigger raises an error")
	}

	var runCount int64
	if err := db.Model(&models.Run{}).Count(&runCount).Error; err != nil {
		t.Fatalf("failed to count runs: %v", err)
	}

	var outboxCount int64
	if err := db.Model(&models.Outbox{}).Count(&outboxCount).Error; err != nil {
		t.Fatalf("failed to count outboxes: %v", err)
	}

	if runCount != 0 {
		t.Fatalf("expected no runs after rollback, got %d", runCount)
	}

	if outboxCount != 0 {
		t.Fatalf("expected no outboxes after rollback, got %d", outboxCount)
	}
}
