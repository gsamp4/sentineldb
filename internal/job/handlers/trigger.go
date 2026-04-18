package handlers

import (
	"sentineldb/internal/job/domain"
	"sentineldb/pkg/logger"

	"github.com/labstack/echo/v4"
)

type TriggerRequest struct {
	ID        string `json:"id"`
}

type TriggerHandler struct {
	Repo   domain.TriggerRepositoryInterface
    Logger *logger.Logger
}

func NewTriggerHandler(repo domain.TriggerRepositoryInterface, log *logger.Logger) *TriggerHandler {
	return &TriggerHandler{Repo: repo, Logger: log}
}

func (h *TriggerHandler) TriggerJob(c echo.Context) error {
	success := h.Repo.RunTrigger()
	if !success {
		return c.JSON(500, map[string]string{"error": "Failed to trigger job"})
	}

	return c.JSON(202, map[string]string{"message": "Job triggered successfully"})
}

func (h *TriggerHandler) GetTrigger(c echo.Context) error {
	var trigger TriggerRequest

	if err := c.Bind(&trigger); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request body"})
	}

	triggerObj, err := h.Repo.GetTrigger(trigger.ID)
	if err != nil {
		h.Logger.Error("Error fetching trigger", err)
		return c.JSON(500, map[string]string{"error": "Failed to fetch trigger"})
	}

	return c.JSON(200, triggerObj)
}