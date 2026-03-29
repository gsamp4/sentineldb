package routes

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	assetDomain "sentineldb/internal/job/domain"
	assetHandler "sentineldb/internal/job/handlers"
)

func InitRoutes(e *echo.Echo, db *gorm.DB) {
    repo    := assetDomain.AssetRepository{DB: db}
    handler := assetHandler.NewHandler(repo)

    v1 := e.Group("/api/v1")
    v1.POST("/assets",       handler.CreateAsset)
    v1.GET("/assets",        handler.GetAssets)
    v1.GET("/assets/:id",    handler.GetAsset)
    v1.PUT("/assets/:id",    handler.UpdateAsset)
    v1.DELETE("/assets/:id", handler.DeleteAsset)
}