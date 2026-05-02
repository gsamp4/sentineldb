package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

var (
    internetDBEndpoint   = "https://internetdb.shodan.io"
    internetDBHTTPClient = &http.Client{Timeout: 15 * time.Second}
)

type internetDBResponse struct {
    IP        string   `json:"ip"`
    Ports     []int    `json:"ports"`
    Hostnames []string `json:"hostnames"`
}

type InternetDBScanResponse struct {
    Matches []struct {
        IP        string   `json:"ip_str"`
        Port      int      `json:"port"`
        Org       string   `json:"org"`
        Hostnames []string `json:"hostnames"`
    } `json:"matches"`
}

func ProcessInternetDB(ctx context.Context, db *gorm.DB, log *logger.Logger, job *models.Outbox) error {
    var asset models.Asset
    if err := db.First(&asset, "id = ?", job.AssetID).Error; err != nil {
        return fmt.Errorf("failed to load asset: %w", err)
    }

    log.Info("loaded asset for InternetDB scan", fmt.Sprintf("asset_value=%s asset_type=%s run_id=%s", asset.Value, asset.Type, job.RunID))

    if asset.Type != "ip" {
        return fmt.Errorf("internetdb scan supports only ip assets, got type %q", asset.Type)
    }

    log.Info("calling InternetDB API", fmt.Sprintf("ip=%s", asset.Value))

    result, err := CallInternetDB(ctx, asset.Value)
    if err != nil {
        return err
    }

    log.Info("InternetDB API call completed", fmt.Sprintf("ip=%s matches=%d", asset.Value, len(result.Matches)))

    var previous models.AssetSnapshot
    db.Where("asset_id = ? AND source = ?", asset.ID, "shodan").
        Order("snapshot_at DESC").
        First(&previous)

	rawData, err := json.Marshal(result)
    if err != nil {
        return fmt.Errorf("failed to marshal InternetDB result: %w", err)
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

// diffSnapshots compares a new InternetDB response against the previous raw snapshot.
// It returns findings only for ports that are new (not present in the previous snapshot).
func diffSnapshots(previousData json.RawMessage, current InternetDBScanResponse, assetID, runID string) []models.Finding {
    previousPorts := make(map[string]bool)

    if len(previousData) > 0 {
        var prev InternetDBScanResponse
        if err := json.Unmarshal(previousData, &prev); err == nil {
            for _, m := range prev.Matches {
                key := fmt.Sprintf("%s:%d", m.IP, m.Port)
                previousPorts[key] = true
            }
        }
    }

    var findings []models.Finding
    for _, m := range current.Matches {
        key := fmt.Sprintf("%s:%d", m.IP, m.Port)
        if previousPorts[key] {
            continue
        }

        severity := "info"
        switch {
        case m.Port == 22 || m.Port == 3389:
            severity = "high"
        case m.Port == 23 || m.Port == 445 || m.Port == 5432 || m.Port == 3306:
            severity = "critical"
        case m.Port == 80 || m.Port == 443:
            severity = "low"
        default:
            severity = "medium"
        }

        detail, _ := json.Marshal(map[string]interface{}{
            "ip":        m.IP,
            "port":      m.Port,
            "org":       m.Org,
            "hostnames": m.Hostnames,
        })

        findings = append(findings, models.Finding{
            ID:       ulid.Make().String(),
            AssetID:  assetID,
            RunID:    runID,
            Source:   "shodan",
            Severity: severity,
            Title:    fmt.Sprintf("New open port %d on %s", m.Port, m.IP),
            Detail:   json.RawMessage(detail),
        })
    }

    return findings
}

func CallInternetDB(ctx context.Context, assetValue string) (InternetDBScanResponse, error) {
    ip, err := netip.ParseAddr(assetValue)
    if err != nil {
        return InternetDBScanResponse{}, fmt.Errorf("internetdb requires a valid IP address: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, internetDBEndpoint+"/"+ip.String(), nil)
    if err != nil {
        return InternetDBScanResponse{}, fmt.Errorf("failed to create request: %w", err)
    }

    resp, err := internetDBHTTPClient.Do(req)
    if err != nil {
        return InternetDBScanResponse{}, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return InternetDBScanResponse{}, fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return InternetDBScanResponse{}, fmt.Errorf("InternetDB API error %d: %s", resp.StatusCode, string(body))
    }

    var raw internetDBResponse
    if err := json.Unmarshal(body, &raw); err != nil {
        return InternetDBScanResponse{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
    }

    result := InternetDBScanResponse{
        Matches: make([]struct {
            IP        string   `json:"ip_str"`
            Port      int      `json:"port"`
            Org       string   `json:"org"`
            Hostnames []string `json:"hostnames"`
        }, 0, len(raw.Ports)),
    }

    for _, port := range raw.Ports {
        result.Matches = append(result.Matches, struct {
            IP        string   `json:"ip_str"`
            Port      int      `json:"port"`
            Org       string   `json:"org"`
            Hostnames []string `json:"hostnames"`
        }{
            IP:        raw.IP,
            Port:      port,
            Hostnames: raw.Hostnames,
        })
    }

    return result, nil
}