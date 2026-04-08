package handlers

import (
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

func (h *TriggerHandler) TriggerJob(c echo.Context) error {
	return nil
}