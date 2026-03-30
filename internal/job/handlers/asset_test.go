package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type MockAssetRepository struct {
	ShouldFail bool
}


func (m MockAssetRepository) RegisterAsset(asset *models.Asset) error {
    if m.ShouldFail {
        return fmt.Errorf("database error")
    }
    return nil
}

func (m MockAssetRepository) ListAssets() ([]models.Asset, error) {
    if m.ShouldFail {
        return nil, fmt.Errorf("database error")
    }
    return []models.Asset{
        {ID: "1", Type: "ip", Value: "192.168.1.1", Label: strPtr("router"), Active: true},
        {ID: "2", Type: "domain", Value: "example.com", Label: strPtr("site"), Active: true},
    }, nil
}

func (m MockAssetRepository) GetAssetByID(id string) (*models.Asset, error) {
    if m.ShouldFail {
        return nil, fmt.Errorf("database error")
    }
    if id == "1" {
        return &models.Asset{ID: "1", Type: "ip", Value: "192.168.1.1", Label: strPtr("router"), Active: true}, nil
    }
    return nil, nil
}

func (m MockAssetRepository) UpdateAsset(id string, label *string, active *bool) error {
    if m.ShouldFail {
        return fmt.Errorf("database error")
    }
    if id != "1" {
        return fmt.Errorf("not found")
    }
    return nil
}

func (m MockAssetRepository) SoftDeleteAsset(id string) error {
    if m.ShouldFail {
        return fmt.Errorf("database error")
    }
    if id != "1" {
        return fmt.Errorf("not found")
    }
    return nil
}

func GetLogger() *logger.Logger {
    return logger.New(logger.Options{
        Level:  logger.LevelInfo,
        Prefix: "TEST: ",
    })
}

func strPtr(s string) *string { return &s }

func TestCreateAsset(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        mockFail   bool
        wantStatus int
    }{
        {"valid ip",       `{"type":"ip","value":"192.168.1.1"}`, false, 201},
        {"invalid type",   `{"type":"url","value":"example.com"}`, false, 400},
        {"invalid body",   `not json`,                             false, 400},
        {"database error", `{"type":"ip","value":"192.168.1.1"}`, true,  500},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            e   := echo.New()
            req := httptest.NewRequest(
                http.MethodPost,
                "/api/v1/assets",
                strings.NewReader(tt.body),
            )
            req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
            rec := httptest.NewRecorder()
            c   := e.NewContext(req, rec)

            repo    := MockAssetRepository{ShouldFail: tt.mockFail}
            handler := NewAssetHandler(repo, GetLogger())

            handler.CreateAsset(c)

            if rec.Code != tt.wantStatus {
                t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
            }
        })
    }
}

func TestGetAssets(t *testing.T) {
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
            req := httptest.NewRequest(http.MethodGet, "/api/v1/assets", nil)
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            repo := MockAssetRepository{ShouldFail: tt.mockFail}
            handler := NewAssetHandler(repo, GetLogger())
            handler.GetAssets(c)
            if rec.Code != tt.wantStatus {
                t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
            }
        })
    }
}

func TestGetAssetByID(t *testing.T) {
    tests := []struct {
        name       string
        id         string
        mockFail   bool
        wantStatus int
    }{
        {"found", "1", false, 200},
        {"not found", "999", false, 404},
        {"database error", "1", true, 500},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            e := echo.New()
            req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/"+tt.id, nil)
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            c.SetParamNames("id")
            c.SetParamValues(tt.id)
            repo := MockAssetRepository{ShouldFail: tt.mockFail}
            handler := NewAssetHandler(repo, GetLogger())
            handler.GetAsset(c)
            if rec.Code != tt.wantStatus {
                t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
            }
        })
    }
}

func TestUpdateAsset(t *testing.T) {
    tests := []struct {
        name       string
        id         string
        body       string
        mockFail   bool
        wantStatus int
    }{
        {"success", "1", `{"label":"newlabel","active":false}`, false, 200},
        {"not found", "999", `{"label":"x"}`, false, 404},
        {"database error", "1", `{"label":"x"}`, true, 500},
        {"invalid body", "1", `not json`, false, 400},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            e := echo.New()
            req := httptest.NewRequest(http.MethodPut, "/api/v1/assets/"+tt.id, strings.NewReader(tt.body))
            req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            c.SetParamNames("id")
            c.SetParamValues(tt.id)
            repo := MockAssetRepository{ShouldFail: tt.mockFail}
            handler := NewAssetHandler(repo, GetLogger())
            handler.UpdateAsset(c)
            if rec.Code != tt.wantStatus {
                t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
            }
        })
    }
}

func TestDeleteAsset(t *testing.T) {
    tests := []struct {
        name       string
        id         string
        mockFail   bool
        wantStatus int
    }{
        {"success", "1", false, 200},
        {"not found", "999", false, 404},
        {"database error", "1", true, 500},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            e := echo.New()
            req := httptest.NewRequest(http.MethodDelete, "/api/v1/assets/"+tt.id, nil)
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            c.SetParamNames("id")
            c.SetParamValues(tt.id)
            repo := MockAssetRepository{ShouldFail: tt.mockFail}
            handler := NewAssetHandler(repo, GetLogger())
            handler.DeleteAsset(c)
            if rec.Code != tt.wantStatus {
                t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
            }
        })
    }
}