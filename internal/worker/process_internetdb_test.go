package worker

import (
	"encoding/json"
	"testing"

	"sentineldb/internal/job/models"
	"sentineldb/internal/services"
)

func TestCompareSnapshotDiffDetectsAddedAndRemovedPortsAndVulns(t *testing.T) {
	previous := testSnapshot(t, services.InternetDBResponse{
		Ports: []int{80, 443},
		Vulns: []string{"CVE-2024-0001", "CVE-2024-0002"},
	})

	current := testSnapshot(t, services.InternetDBResponse{
		Ports: []int{443, 8080},
		Vulns: []string{"CVE-2024-0002", "CVE-2024-9999"},
	})

	findings, err := compareSnapshotDiff(previous, current)
	if err != nil {
		t.Fatalf("compareSnapshotDiff returned error: %v", err)
	}

	assertFinding(t, findings, "Ports added", []int{8080}, "high")
	assertFinding(t, findings, "Ports removed", []int{80}, "high")
	assertFinding(t, findings, "Vulns added", []string{"CVE-2024-9999"}, "critical")
	assertFinding(t, findings, "Vulns removed", []string{"CVE-2024-0001"}, "critical")

	if len(findings) != 4 {
		t.Fatalf("expected 4 findings, got %d", len(findings))
	}
}

func TestCompareSnapshotDiffIgnoresUnchangedFields(t *testing.T) {
	previous := testSnapshot(t, services.InternetDBResponse{
		Ports: []int{80},
		Vulns: []string{"CVE-2024-0001"},
	})

	current := testSnapshot(t, services.InternetDBResponse{
		Ports: []int{80},
		Vulns: []string{"CVE-2024-0001"},
	})

	findings, err := compareSnapshotDiff(previous, current)
	if err != nil {
		t.Fatalf("compareSnapshotDiff returned error: %v", err)
	}

	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(findings))
	}
}

func testSnapshot(t *testing.T, response services.InternetDBResponse) *models.AssetSnapshot {
	t.Helper()

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal test response: %v", err)
	}

	return &models.AssetSnapshot{
		AssetID: "asset-1",
		RunID:   "run-1",
		Data:    data,
	}
}

func assertFinding[T comparable](t *testing.T, findings []models.Finding, title string, expected []T, severity string) {
	t.Helper()

	for _, finding := range findings {
		if finding.Title != title {
			continue
		}

		if finding.Severity != severity {
			t.Fatalf("expected severity %q for %q, got %q", severity, title, finding.Severity)
		}

		var values []T
		if err := json.Unmarshal(finding.Detail, &values); err != nil {
			t.Fatalf("failed to unmarshal detail for %q: %v", title, err)
		}

		if len(values) != len(expected) {
			t.Fatalf("expected %d values for %q, got %d", len(expected), title, len(values))
		}

		for index, value := range expected {
			if values[index] != value {
				t.Fatalf("expected value %v at index %d for %q, got %v", value, index, title, values[index])
			}
		}

		return
	}

	t.Fatalf("finding %q not found", title)
}