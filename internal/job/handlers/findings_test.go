package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sentineldb/internal/job/models"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type MockFindingRepository struct {
	ShouldFail bool
}

func (m MockFindingRepository) ListFindings() ([]models.Finding, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("database error")
	}
	return []models.Finding{
		{ID: "1", AssetID: "a1", RunID: "r1", Source: "shodan", Severity: "high", Title: "Open port 22"},
		{ID: "2", AssetID: "a2", RunID: "r1", Source: "hibp", Severity: "critical", Title: "Leaked credentials"},
	}, nil
}

func (m MockFindingRepository) GetFindingByID(id string) (*models.Finding, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("database error")
	}
	if id == "1" {
		return &models.Finding{ID: "1", AssetID: "a1", RunID: "r1", Source: "shodan", Severity: "high", Title: "Open port 22"}, nil
	}
	return nil, nil
}

// Implement the UpdateFindingStatus method for the mock repository
func (m MockFindingRepository) UpdateFindingStatus(findingID string, status string) error {
	if m.ShouldFail {
		return fmt.Errorf("database error")
	}
	return nil
}

func TestGetFindings(t *testing.T) {
	tests := []struct {
		name       string
		mockFail   bool
		wantStatus int
	}{
		{"success", false, 200},
		{"database error", true, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/findings", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			repo := MockFindingRepository{ShouldFail: tt.mockFail}
			handler := NewFindingHandler(repo, GetLogger())

			handler.GetFindings(c)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestGetFindingByID(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockFail   bool
		wantStatus int
	}{
		{"found", `{"id":"1"}`, false, 200},
		{"not found", `{"id":"999"}`, false, 404},
		{"database error", `{"id":"1"}`, true, 500},
		{"invalid body", `not json`, false, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/findings/1", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			repo := MockFindingRepository{ShouldFail: tt.mockFail}
			handler := NewFindingHandler(repo, GetLogger())

			handler.GetFindingByID(c)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestUpdateFinding_InvalidBody(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/findings/1/resolve", strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	repo := MockFindingRepository{}
	handler := NewFindingHandler(repo, GetLogger())

	handler.UpdateFinding(c)

	if rec.Code != 400 {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
