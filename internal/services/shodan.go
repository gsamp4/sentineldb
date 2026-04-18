package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type ShodanResponse struct {
	Matches []struct {
		IP        string `json:"ip_str"`
		Port      int    `json:"port"`
		Org       string `json:"org"`
		Hostnames []string `json:"hostnames"`
	} `json:"matches"`
}

func processShodan(ctx context.Context, db *gorm.DB, log *logger.Logger, job *models.Outbox) error {
    var asset models.Asset
    db.First(&asset, "id = ?", job.AssetID)

    result, err := CallShodan(asset.Value)
    if err != nil {
        return err
    }

    var previous models.AssetSnapshot
    db.Where("asset_id = ? AND source = ?", asset.ID, "shodan").
        Order("snapshot_at DESC").
        First(&previous)

	rawData, err := json.Marshal(result)
    if err != nil {
        return fmt.Errorf("failed to marshal Shodan result: %w", err)
    }

    snapshot := models.AssetSnapshot{
        ID:      ulid.Make().String(),
        AssetID: asset.ID,
        RunID:   job.RunID,
        Source:  "shodan",
        Data:    json.RawMessage(rawData),
    }
    db.Create(&snapshot)

    findings := diffSnapshots(previous.Data, result, asset.ID, job.RunID)
    for _, f := range findings {
        db.Create(&f)
    }

    return nil
}

func CallShodan(assetValue string) (ShodanResponse, error) {
    apiKey := os.Getenv("SHODAN_API_KEY")
    if apiKey == "" {
        return ShodanResponse{}, fmt.Errorf("SHODAN_API_KEY environment variable is not set")
    }

    query := fmt.Sprintf(`hostname:"%s"`, assetValue)
    endpoint := "https://api.shodan.io/shodan/host/search"

    params := url.Values{}
    params.Add("key", apiKey)
    params.Add("query", query)

    resp, err := http.Get(endpoint + "?" + params.Encode())
    if err != nil {
        return ShodanResponse{}, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return ShodanResponse{}, fmt.Errorf("Shodan API error %d: %s", resp.StatusCode, string(body))
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return ShodanResponse{}, fmt.Errorf("failed to read response: %w", err)
    }

    var result ShodanResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return ShodanResponse{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
    }

    return result, nil
}