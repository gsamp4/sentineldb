package handlers

import (
	"sentineldb/internal/job/domain"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"github.com/labstack/echo/v4"
)

type CreateAssetRequest struct {
	Type  string `json:"type" validate:"required,oneof=ip domain email"`
	Value string `json:"value" validate:"required"`
	Label string `json:"label"`
}

type AssetResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

type AssetHandler struct {
    Repo   domain.AssetRepositoryInterface
    Logger *logger.Logger
}

func NewAssetHandler(repo domain.AssetRepositoryInterface, log *logger.Logger) *AssetHandler {
    return &AssetHandler{Repo: repo, Logger: log}
}

func (h *AssetHandler) CreateAsset(c echo.Context) error {
    var body CreateAssetRequest

    if err := c.Bind(&body); err != nil {
        return c.JSON(400, map[string]string{"error": "invalid request body"})
    }

    if err := domain.ValidateAsset(body.Type, body.Value); err != nil {
        return c.JSON(400, map[string]string{"error": err.Error()})
    }

    asset := models.Asset{
        Type:  body.Type,
        Value: body.Value,
        Label: &body.Label,
    }

    if err := h.Repo.RegisterAsset(&asset); err != nil {
        h.Logger.Error("Failed to create asset", err)
        return c.JSON(500, map[string]string{"error": "failed to create asset"})
    }

    return c.JSON(201, map[string]string{"message": "Asset created successfully"})
}

func (h *AssetHandler) GetAssets(c echo.Context) error {
    assets, err := h.Repo.ListAssets()
    if err != nil {
        h.Logger.Error("Failed to list assets", err)
        return c.JSON(500, map[string]string{"error": "failed to list assets"})
    }
    return c.JSON(200, assets)
}

func (h *AssetHandler) GetAsset(c echo.Context) error {
    id := c.Param("id")
    asset, err := h.Repo.GetAssetByID(id)
    if err != nil {
        return c.JSON(500, map[string]string{"error": "failed to get asset"})
    }
    if asset == nil {
        return c.JSON(404, map[string]string{"error": "asset not found"})
    }
    return c.JSON(200, asset)
}

func (h *AssetHandler) UpdateAsset(c echo.Context) error {
    id := c.Param("id")
    var body struct {
        Label  *string `json:"label"`
        Active *bool  `json:"active"`
    }
    if err := c.Bind(&body); err != nil {
        return c.JSON(400, map[string]string{"error": "invalid request body"})
    }
    err := h.Repo.UpdateAsset(id, body.Label, body.Active)
    if err != nil {
        if err.Error() == "not found" {
            return c.JSON(404, map[string]string{"error": "asset not found"})
        }
        h.Logger.Error("Failed to update asset", err)
        return c.JSON(500, map[string]string{"error": "failed to update asset"})
    }
    return c.JSON(200, map[string]string{"message": "Asset updated successfully"})
}

func (h *AssetHandler) DeleteAsset(c echo.Context) error {
    id := c.Param("id")
    err := h.Repo.SoftDeleteAsset(id)
    if err != nil {
        if err.Error() == "not found" {
            return c.JSON(404, map[string]string{"error": "asset not found"})
        }
        h.Logger.Error("Failed to delete asset", err)
        return c.JSON(500, map[string]string{"error": "failed to delete asset"})
    }
    return c.JSON(200, map[string]string{"message": "Asset deleted (soft) successfully"})
}