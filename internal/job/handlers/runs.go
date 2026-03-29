package handlers

import (
	"sentineldb/internal/job/domain"
	"sentineldb/pkg/logger"

	"github.com/labstack/echo/v4"
)

type RunHandler struct {
	Repo   domain.RunRepositoryInterface
    Logger *logger.Logger
}

func NewRunHandler(repo domain.RunRepositoryInterface, log *logger.Logger) *RunHandler {
	return &RunHandler{Repo: repo, Logger: log}
}

func (h *RunHandler) GetRuns(c echo.Context) error {
	return nil
}

func (h *RunHandler) GetRunByID(c echo.Context) error {
	param := c.Param("id")
	if param == "" {
		return c.JSON(400, map[string]string{"message": "id parameter is required"})
	}
	return nil
}