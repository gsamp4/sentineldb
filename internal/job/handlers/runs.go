package handlers

import (
	"sentineldb/internal/job/domain"
	"sentineldb/pkg/logger"

	"github.com/labstack/echo/v4"
)

type RunRequest struct {
	ID        string `json:"id"`
}

type RunHandler struct {
	Repo   domain.RunRepositoryInterface
    Logger *logger.Logger
}

func NewRunHandler(repo domain.RunRepositoryInterface, log *logger.Logger) *RunHandler {
	return &RunHandler{Repo: repo, Logger: log}
}

func (h *RunHandler) GetRuns(c echo.Context) error {
	runs, err := h.Repo.ListRuns()
	if err != nil {
		h.Logger.Error("Failed to list runs", err)
		return c.JSON(500, map[string]string{"message": "failed to list runs"})
	}

	return c.JSON(200, runs)
}

func (h *RunHandler) GetRunByID(c echo.Context) error {
	param := c.Param("id")
	if param == "" {
		return c.JSON(400, map[string]string{"message": "id parameter is required"})
	}

	run, err := h.Repo.GetRunByID(param)
	if err != nil {
		h.Logger.Error("Failed to get run by id", err)
		return c.JSON(500, map[string]string{"message": "failed to get run"})
	}
	if run == nil {
		return c.JSON(404, map[string]string{"message": "run not found"})
	}

	return c.JSON(200, run)
}