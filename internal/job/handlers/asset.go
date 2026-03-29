package handlers

import (
	"sentineldb/internal/job/domain"
	"sentineldb/internal/job/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Handler struct {
    Repo domain.AssetRepositoryInterface
}

func NewHandler(repo domain.AssetRepositoryInterface) *Handler {
	return &Handler{Repo: repo}
}

func (h *Handler) CreateAsset(c echo.Context) error {
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
        log.Error("Failed to create asset: ", err)
        return c.JSON(500, map[string]string{"error": "failed to create asset"})
    }

    return c.JSON(201, map[string]string{"message": "Asset created successfully"})
}

func (h *Handler) GetAssets(c echo.Context) error {
    assets, err := h.Repo.ListAssets()
    if err != nil {
        log.Error("Failed to list assets: ", err)

        return c.JSON(500, map[string]string{"error": "failed to list assets"})
    }
    return c.JSON(200, assets)
}

func (h *Handler) GetAsset(c echo.Context) error {
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

func (h *Handler) UpdateAsset(c echo.Context) error {
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
        log.Error("Failed to update asset: ", err)
        return c.JSON(500, map[string]string{"error": "failed to update asset"})
    }
    return c.JSON(200, map[string]string{"message": "Asset updated successfully"})
}

func (h *Handler) DeleteAsset(c echo.Context) error {
    id := c.Param("id")
    err := h.Repo.SoftDeleteAsset(id)
    if err != nil {
        if err.Error() == "not found" {
            return c.JSON(404, map[string]string{"error": "asset not found"})
        }
        log.Error("Failed to delete asset: ", err)
        return c.JSON(500, map[string]string{"error": "failed to delete asset"})
    }
    return c.JSON(200, map[string]string{"message": "Asset deleted (soft) successfully"})
}