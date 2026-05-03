package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sentineldb/internal/job/domain"
	"testing"

	"github.com/labstack/echo/v4"
)

type MockTriggerRepository struct {
	ShouldFail         bool
	TriggerAllError    error
	TriggerByIDError   error
	TriggerAllResponse *domain.TriggerResult
	TriggerByIDResult  *domain.TriggerResult
}

func (m MockTriggerRepository) TriggerAllActiveAssets() (*domain.TriggerResult, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("database error")
	}
	if m.TriggerAllError != nil {
		return nil, m.TriggerAllError
	}
	if m.TriggerAllResponse != nil {
		return m.TriggerAllResponse, nil
	}
	return &domain.TriggerResult{RunID: "run-all", JobCount: 2}, nil
}

func (m MockTriggerRepository) TriggerAssetByID(id string) (*domain.TriggerResult, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("database error")
	}
	if m.TriggerByIDError != nil {
		return nil, m.TriggerByIDError
	}
	if m.TriggerByIDResult != nil {
		return m.TriggerByIDResult, nil
	}
	return &domain.TriggerResult{RunID: "run-one", JobCount: 1}, nil
}

func TestTriggerAll(t *testing.T) {
	tests := []struct {
		name       string
		repo       MockTriggerRepository
		wantStatus int
	}{
		{name: "success", repo: MockTriggerRepository{}, wantStatus: http.StatusAccepted},
		{name: "no supported assets", repo: MockTriggerRepository{TriggerAllError: domain.ErrTriggerNoSupportedAssets}, wantStatus: http.StatusNotFound},
		{name: "database error", repo: MockTriggerRepository{ShouldFail: true}, wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trigger", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := NewTriggerHandler(tt.repo, GetLogger())
			if err := handler.TriggerAll(c); err != nil {
				t.Fatalf("TriggerAll returned error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestTriggerByAssetID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		repo       MockTriggerRepository
		wantStatus int
	}{
		{name: "success", id: "asset-1", repo: MockTriggerRepository{}, wantStatus: http.StatusAccepted},
		{name: "missing id", id: "", repo: MockTriggerRepository{}, wantStatus: http.StatusBadRequest},
		{name: "asset not found", id: "asset-404", repo: MockTriggerRepository{TriggerByIDError: domain.ErrTriggerAssetNotFound}, wantStatus: http.StatusNotFound},
		{name: "unsupported asset", id: "asset-email", repo: MockTriggerRepository{TriggerByIDError: domain.ErrTriggerUnsupportedAsset}, wantStatus: http.StatusBadRequest},
		{name: "database error", id: "asset-1", repo: MockTriggerRepository{ShouldFail: true}, wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trigger/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.id != "" {
				c.SetParamNames("id")
				c.SetParamValues(tt.id)
			}

			handler := NewTriggerHandler(tt.repo, GetLogger())
			if err := handler.TriggerByAssetID(c); err != nil {
				t.Fatalf("TriggerByAssetID returned error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
