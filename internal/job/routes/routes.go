package routes

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	assetHandler "sentineldb/internal/job/handlers"
)

func InitRoutes(e *echo.Echo, db *gorm.DB) {
	h := assetHandler.NewHandler(db)
	v1 := e.Group("/api/v1")
	v1.POST("/assets", h.CreateAsset)
	v1.GET("/assets", h.GetAssets)
	v1.PUT("/assets/:id", h.UpdateAsset)
	v1.DELETE("/assets/:id", h.DeleteAsset)
}
