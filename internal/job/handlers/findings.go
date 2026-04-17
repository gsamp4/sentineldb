package handlers

import (
	"sentineldb/internal/job/domain"
	"sentineldb/pkg/logger"

	"github.com/labstack/echo/v4"
)

type FindingRequest struct {
	ID        string `json:"id"`
}

type FindingHandler struct {
	Repo   domain.FindingRepositoryInterface
    Logger *logger.Logger
}

func NewFindingHandler(repo domain.FindingRepositoryInterface, log *logger.Logger) *FindingHandler {
	return &FindingHandler{Repo: repo, Logger: log}
}

func (h *FindingHandler) GetFindings(c echo.Context) error {
	findings, err := h.Repo.ListFindings()
	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to fetch findings"})
	}
	return c.JSON(200, findings)
}

func (h *FindingHandler) GetFindingByID(c echo.Context) error {
	var req FindingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request body"})
	}

	finding, err := h.Repo.GetFindingByID(req.ID)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to fetch finding"})
	}

	if finding == nil {
		return c.JSON(404, map[string]string{"error": "Finding not found"})
	}
	return c.JSON(200, finding)
}