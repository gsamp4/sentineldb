package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sentineldb/internal/job/domain"
	"sentineldb/internal/job/models"
	"testing"

	"github.com/labstack/echo/v4"
)

type MockRunRepository struct {
    ShouldFail bool
}

func (m MockRunRepository) ListRuns() ([]models.Run, error) {
    if m.ShouldFail {
        return nil, fmt.Errorf("database error")
    }
    return []models.Run{
        {ID: "1", Status: "pending"},
        {ID: "2", Status: "done"},
    }, nil
}

func (m MockRunRepository) GetRunByID(id string) (*models.Run, error) {
    if m.ShouldFail {
        return nil, fmt.Errorf("database error")
    }
    if id == "1" {
        return &models.Run{ID: "1", Status: "pending"}, nil
    }
    return nil, nil
}

func NewTestRunHandler(repo domain.RunRepositoryInterface) *RunHandler {
    return NewRunHandler(repo, GetLogger())
}

func TestGetRuns(t *testing.T) {
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
            req := httptest.NewRequest(http.MethodGet, "/api/v1/runs", nil)
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            repo := MockRunRepository{ShouldFail: tt.mockFail}
            handler := NewTestRunHandler(repo)
            handler.GetRuns(c)
            if rec.Code != tt.wantStatus {
                t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
            }
        })
    }
}

func TestGetRunByID(t *testing.T) {
    tests := []struct {
        name       string
        id         string
        mockFail   bool
        wantStatus int
    }{
        {"found", "1", false, 200},
        {"not found", "999", false, 404},
        {"database error", "1", true, 500},
        {"missing id", "", false, 400},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            e := echo.New()
            req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/"+tt.id, nil)
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            if tt.id != "" {
                c.SetParamNames("id")
                c.SetParamValues(tt.id)
            }
            repo := MockRunRepository{ShouldFail: tt.mockFail}
            handler := NewTestRunHandler(repo)
            handler.GetRunByID(c)
            if rec.Code != tt.wantStatus {
                t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
            }
        })
    }
}

func (m MockRunRepository) GetRunJobs(id string) ([]models.Outbox, error) {
    if m.ShouldFail {
        return nil, fmt.Errorf("database error")
    }
    // Return dummy jobs for testing
    if id == "1" {
        return []models.Outbox{
            {ID: "job1", RunID: "1", Status: "pending"},
            {ID: "job2", RunID: "1", Status: "done"},
        }, nil
    }
    return []models.Outbox{}, nil
}