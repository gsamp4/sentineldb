package routes

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	assetDomain "sentineldb/internal/job/domain"
	jobHandlers "sentineldb/internal/job/handlers"
	"sentineldb/pkg/logger"
)

func InitRoutes(e *echo.Echo, db *gorm.DB, log *logger.Logger) {
    assetRepo    := assetDomain.AssetRepository{DB: db, Logger: log}
    assetHandler := jobHandlers.NewAssetHandler(assetRepo, log)

    // runs
    runsRepo    := assetDomain.RunRepository{DB: db, Logger: log}
    runsHandler := jobHandlers.NewRunHandler(runsRepo, log)

    // triggers
    triggerRepo   := assetDomain.TriggerRepository{DB: db, Logger: log}
    triggerHandler := jobHandlers.NewTriggerHandler(triggerRepo, log)

    // findings
    findingsRepo    := assetDomain.FindingRepository{DB: db, Logger: log}
    findingsHandler := jobHandlers.NewFindingHandler(findingsRepo, log)

    v1 := e.Group("/api/v1")

    v1.POST("/assets",       assetHandler.CreateAsset)
    v1.GET("/assets",        assetHandler.GetAssets)
    v1.GET("/assets/:id",    assetHandler.GetAsset)
    v1.PUT("/assets/:id",    assetHandler.UpdateAsset)
    v1.DELETE("/assets/:id", assetHandler.DeleteAsset)

    v1.GET("/runs",      runsHandler.GetRuns)
    v1.GET("/runs/:id",  runsHandler.GetRunByID)
    v1.GET("/runs/:id/jobs",  runsHandler.GetRunJobs)

    v1.POST("/trigger", triggerHandler.TriggerJob)
    v1.POST("/trigger/:id", triggerHandler.GetTrigger)

    v1.GET("/findings", findingsHandler.GetFindings)
    v1.GET("/findings/:id", findingsHandler.GetFindingByID)
    v1.PATCH("/findings/:id/resolve", findingsHandler.UpdateFinding)
}