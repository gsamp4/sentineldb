package handlers

import (
	"sentineldb/internal/job/domain"
	"sentineldb/internal/job/models"

	"github.com/labstack/echo/v4"
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
        return c.JSON(500, map[string]string{"error": "failed to create asset"})
    }

    return c.JSON(201, map[string]string{"message": "Asset created successfully"})
}

func (h *Handler) GetAssets() {}

func (h *Handler) UpdateAsset() {}

func (h *Handler) DeleteAsset() {}