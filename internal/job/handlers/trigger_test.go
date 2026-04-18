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

type MockTriggerRepository struct {
	ShouldFail bool
}

func (m MockTriggerRepository) RunTrigger() bool {
	return !m.ShouldFail
}

func (m MockTriggerRepository) GetTrigger(id string) (*models.Run, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("database error")
	}
	if id == "1" {
		return &models.Run{ID: "1", Status: "pending"}, nil
	}
	return nil, fmt.Errorf("not found")
}

func TestTriggerJob(t *testing.T) {
	tests := []struct {
		name       string
		mockFail   bool
		wantStatus int
	}{
		{"success", false, 200},
		{"trigger failure", true, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trigger", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			repo := MockTriggerRepository{ShouldFail: tt.mockFail}
			handler := NewTriggerHandler(repo, GetLogger())

			handler.TriggerJob(c)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestGetTrigger(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockFail   bool
		wantStatus int
	}{
		{"found", `{"id":"1"}`, false, 200},
		{"not found", `{"id":"999"}`, false, 500},
		{"database error", `{"id":"1"}`, true, 500},
		{"invalid body", `not json`, false, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trigger/1", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			repo := MockTriggerRepository{ShouldFail: tt.mockFail}
			handler := NewTriggerHandler(repo, GetLogger())

			handler.GetTrigger(c)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
