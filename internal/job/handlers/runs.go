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
	return nil
}

func (h *RunHandler) GetRunByID(c echo.Context) error {
	param := c.Param("id")
	if param == "" {
		return c.JSON(400, map[string]string{"message": "id parameter is required"})
	}
	return nil
}

func (h *RunHandler) GetRunJobs(c echo.Context) error {
	var req RunRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"message": "invalid request body"})
	}

	jobs, err := h.Repo.GetRunJobs(req.ID)
	if err != nil {
		return c.JSON(404, map[string]string{"message": "run jobs not found"})
	}
	return c.JSON(200, jobs)
}