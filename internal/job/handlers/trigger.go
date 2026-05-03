package handlers

import (
	"errors"
	"net/http"

	"sentineldb/internal/job/domain"
	"sentineldb/pkg/logger"

	"github.com/labstack/echo/v4"
)

type TriggerHandler struct {
	Repo   domain.TriggerRepositoryInterface
	Logger *logger.Logger
}

func NewTriggerHandler(repo domain.TriggerRepositoryInterface, log *logger.Logger) *TriggerHandler {
	return &TriggerHandler{Repo: repo, Logger: log}
}

func (h *TriggerHandler) TriggerAll(c echo.Context) error {
	result, err := h.Repo.TriggerAllActiveAssets()
	if err != nil {
		return h.handleTriggerError(c, err)
	}

	return c.JSON(http.StatusAccepted, result)
}

func (h *TriggerHandler) TriggerByAssetID(c echo.Context) error {
	assetID := c.Param("id")
	if assetID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "asset id parameter is required"})
	}

	result, err := h.Repo.TriggerAssetByID(assetID)
	if err != nil {
		return h.handleTriggerError(c, err)
	}

	return c.JSON(http.StatusAccepted, result)
}

func (h *TriggerHandler) handleTriggerError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrTriggerAssetNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"message": "asset not found"})
	case errors.Is(err, domain.ErrTriggerNoSupportedAssets):
		return c.JSON(http.StatusNotFound, map[string]string{"message": "no supported active assets found"})
	case errors.Is(err, domain.ErrTriggerUnsupportedAsset):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "asset type is not supported for trigger yet"})
	default:
		h.Logger.Error("Failed to trigger run", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "failed to create trigger run"})
	}
}
