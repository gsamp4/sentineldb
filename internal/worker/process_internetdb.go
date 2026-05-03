package worker

import (
	"sentineldb/internal/job/models"
	"github.com/google/uuid"
)

func processInternetDB(
	ctx context.Context,
	db *gorm.DB,
	log *logger.Logger,
	job models.Outbox,
	result *InternetDBResponse
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
		log.Errorf("Error saving AssetSnapshot", "error", tx.Error)
		return fmt.Errorf("error saving snapshot: %v", tx.Error)
	}

	if tx.RowsAffected == 0 {
		log.Error("No snapshot has been created.")
		return fmt.Errorf("no snapshot has been created.")
	}

	var previousSnapshot *models.AssetSnapshot
	queryResult , err := db.Raw("SELECT * FROM asset_snapshots WHERE asset_id = ? AND source = ? ORDER BY snapshot_at DESC LIMIT 1 OFFSET 1",
    job.AssetID, "shodan_internetdb").Scan(&previousSnapshot)

	if err != nil {
		log.Errorf("No snapshot has been created: %v", err)
		return fmt.Errorf("error getting previous snapshot: %v", err)
	}

	if queryResult .RowsAffected == 0 {
		log.Error("No previous snapshot found.")
		return nil
	}

	newFindings, err := compareSnapshotDiff(previousSnapshot, assetSnapshot)
	if err != nil {
		log.Errorf("error comparing snapshots: %v", err)
		return fmt.Errorf("error getting previous snapshot")
	}

	tx := db.WithContext(ctx).Create(&newFindings)
	if tx.Error != nil {
		log.Errorf("Error saving NEW AssetSnapshot", "error", tx.Error)
		return fmt.Errorf("error saving NEW snapshot: %v", tx.Error)
	}

	return nil
}

func compareSnapshotDiff(previous *models.AssetSnapshot, current *models.AssetSnapshot) ([]models.Finding, error) {
	var prevData, currData InternetDBResponse

	if err := json.Unmarshal(previous.Data, &prevData); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(current.Data, &currData); err != nil {
		return nil, err
	}

	diffs := make(map[string]interface{})

	getDiffs := func[T comparable](prev, curr []T, key string) {
		mapPrevious := make(map[T]struct{})
		for _, v := range prev {
			mapPrevious[v] = struct{}{}
		}

		var added []T
		for _, v := range curr {
			if _, found := mapPrevious[v]; !found {
				added = append(added, v)
			}
		}

		if len(added) > 0 {
			diffs[key] = added
		}
	}

	getDiffs(prevData.Ports, currData.Ports, "ports")
	getDiffs(prevData.Vulns, currData.Vulns, "vulns")
	getDiffs(prevData.Tags, currData.Tags, "tags")
	getDiffs(prevData.Hostnames, currData.Hostnames, "tags")

	newFindings := models.Finding{
		ID:         uuid.NewString(),
		AssetID:    uuid.NewString(),
		RunID:      uuid.NewString(),
		Source:     "shodan_internetdb",
		Data:       json.RawMessage(diffs),
		SnapshotAt: time.Now(),
	}
	return newFindings, nil
}
