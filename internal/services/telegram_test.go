package services

import (
	"encoding/json"
	"strings"
	"testing"

	"sentineldb/internal/job/models"
)

func TestBuildTelegramFindingMessageFormatsMetadataAndListDetails(t *testing.T) {
	detail, err := json.Marshal([]int{80, 443})
	if err != nil {
		t.Fatalf("failed to marshal detail: %v", err)
	}

	message := buildTelegramFindingMessage(models.Finding{
		Title:    "Ports removed",
		Severity: "high",
		Source:   "shodan_internetdb",
		AssetID:  "asset-1",
		RunID:    "run-1",
		Detail:   detail,
	})

	for _, expected := range []string{
		"SentinelDB detected a finding update.",
		"Title: Ports removed",
		"Severity: HIGH",
		"Source: shodan_internetdb",
		"Asset ID: asset-1",
		"Run ID: run-1",
		"Details:\n- 80\n- 443",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected message to contain %q, got %q", expected, message)
		}
	}
}

func TestBuildTelegramFindingMessageHandlesEmptyDetail(t *testing.T) {
	message := buildTelegramFindingMessage(models.Finding{})

	if !strings.Contains(message, "No additional detail provided.") {
		t.Fatalf("expected empty detail fallback, got %q", message)
	}

	if !strings.Contains(message, "Severity: UNKNOWN") {
		t.Fatalf("expected severity fallback, got %q", message)
	}
}

func TestMaskTelegramTarget(t *testing.T) {
	if masked := maskTelegramTarget("123456789"); masked != "*****6789" {
		t.Fatalf("expected masked chat id, got %q", masked)
	}

	if masked := maskTelegramTarget("1234"); masked != "1234" {
		t.Fatalf("expected short chat id to remain unchanged, got %q", masked)
	}
}