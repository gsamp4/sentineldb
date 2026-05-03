package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sentineldb/internal/job/models"
	"sentineldb/internal/services"
	"sentineldb/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func processInternetDB(
	ctx context.Context,
	db *gorm.DB,
	log *logger.Logger,
	job models.Outbox,
	result *services.InternetDBResponse,
) error {

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Error("Error serializing InternetDBResponse", "error", err)
		return fmt.Errorf("error serializing InternetDBResponse: %v", err)
	}

	assetSnapshot := models.AssetSnapshot{
		ID:         uuid.NewString(),
		AssetID:    job.AssetID,
		RunID:      job.RunID,
		Source:     "shodan_internetdb",
		Data:       json.RawMessage(jsonData),
		SnapshotAt: time.Now(),
	}

	tx := db.WithContext(ctx).Create(&assetSnapshot)
	if tx.Error != nil {
		log.Error("Error saving AssetSnapshot", "error", tx.Error)
		return fmt.Errorf("error saving snapshot: %v", tx.Error)
	}

	if tx.RowsAffected == 0 {
		log.Error("No snapshot has been created.")
		return fmt.Errorf("no snapshot has been created.")
	}

	var previousSnapshot *models.AssetSnapshot
	queryResult := db.Raw("SELECT * FROM asset_snapshots WHERE asset_id = ? AND source = ? ORDER BY snapshot_at DESC LIMIT 1 OFFSET 1",
    job.AssetID, "shodan_internetdb").Scan(&previousSnapshot)

	if queryResult.Error != nil {
		log.Error("error getting previous snapshot", "error", queryResult.Error)
		return fmt.Errorf("error getting previous snapshot: %v", err)
	}

	if queryResult .RowsAffected == 0 {
		log.Error("No previous snapshot found.")
		return nil
	}

	newFindings, err := compareSnapshotDiff(previousSnapshot, &assetSnapshot)
	if err != nil {
		log.Error("error comparing snapshots", "error", err)
		return fmt.Errorf("error getting previous snapshot")
	}

	if len(newFindings) > 0 {
		tx = db.WithContext(ctx).Create(&newFindings)
		if tx.Error != nil {
			log.Error("Error saving NEW AssetSnapshot", "error", tx.Error)
			return fmt.Errorf("error saving NEW snapshot: %v", tx.Error)
		}

		telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

		for _, finding := range newFindings {
			if err := services.Notify(log, telegramToken, telegramChatID, finding); err != nil {
				log.Error("telegram notification failed", "finding_id", finding.ID, "error", err)
			}
		}
	}

	return nil
}

func compareSnapshotDiff(previous *models.AssetSnapshot, current *models.AssetSnapshot) ([]models.Finding, error) {
	var prevData, currData services.InternetDBResponse

	if err := json.Unmarshal(previous.Data, &prevData); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(current.Data, &currData); err != nil {
		return nil, err
	}

	var changes []snapshotChange

	changes = append(changes, collectSnapshotChanges(prevData.Ports, currData.Ports, "ports", true)...)
	changes = append(changes, collectSnapshotChanges(prevData.Vulns, currData.Vulns, "vulns", true)...)
	changes = append(changes, collectSnapshotChanges(prevData.Tags, currData.Tags, "tags", false)...)
	changes = append(changes, collectSnapshotChanges(prevData.Hostnames, currData.Hostnames, "hostnames", false)...)

	var findings []models.Finding

	for _, change := range changes {
		detailJSON, err := json.Marshal(change.values)
		if err != nil {
			return nil, err
		}

		newFinding := models.Finding{
			ID:       uuid.NewString(),
			AssetID:  current.AssetID,
			RunID:    current.RunID,
			Source:   "shodan_internetdb",
			Severity: snapshotFieldSeverity(change.field),
			Title:    snapshotChangeTitle(change.field, change.action),
			Detail:   detailJSON,
			SeenAt:   time.Now(),
		}
		findings = append(findings, newFinding)
	}
	return findings, nil
}

type snapshotChange struct {
	field  string
	action string
	values interface{}
}

func collectSnapshotChanges[T comparable](prev, curr []T, field string, detectRemovals bool) []snapshotChange {
	prevSet := make(map[T]struct{}, len(prev))
	currSet := make(map[T]struct{}, len(curr))

	for _, value := range prev {
		prevSet[value] = struct{}{}
	}

	for _, value := range curr {
		currSet[value] = struct{}{}
	}

	var changes []snapshotChange
	var added []T
	for _, value := range curr {
		if _, found := prevSet[value]; !found {
			added = append(added, value)
		}
	}

	if len(added) > 0 {
		changes = append(changes, snapshotChange{
			field:  field,
			action: "added",
			values: added,
		})
	}

	if !detectRemovals {
		return changes
	}

	var removed []T
	for _, value := range prev {
		if _, found := currSet[value]; !found {
			removed = append(removed, value)
		}
	}

	if len(removed) > 0 {
		changes = append(changes, snapshotChange{
			field:  field,
			action: "removed",
			values: removed,
		})
	}

	return changes
}

func snapshotFieldSeverity(field string) string {
	switch field {
	case "ports":
		return "high"
	case "vulns":
		return "critical"
	default:
		return "low"
	}
}

func snapshotChangeTitle(field, action string) string {
	switch {
	case field == "ports" && action == "added":
		return "Ports added"
	case field == "ports" && action == "removed":
		return "Ports removed"
	case field == "vulns" && action == "added":
		return "Vulns added"
	case field == "vulns" && action == "removed":
		return "Vulns removed"
	case field == "tags":
		return "Tags added"
	default:
		return "Hostnames added"
	}
}
