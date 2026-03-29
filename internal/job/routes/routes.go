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

    v1 := e.Group("/api/v1")

    v1.POST("/assets",       assetHandler.CreateAsset)
    v1.GET("/assets",        assetHandler.GetAssets)
    v1.GET("/assets/:id",    assetHandler.GetAsset)
    v1.PUT("/assets/:id",    assetHandler.UpdateAsset)
    v1.DELETE("/assets/:id", assetHandler.DeleteAsset)

    v1.GET("/runs",      runsHandler.GetRuns)
    v1.GET("/runs/:id",  runsHandler.GetRunByID)
}