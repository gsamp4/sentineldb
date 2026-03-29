package handlers

import "github.com/labstack/echo/v4"

func GetRuns(c echo.Context) error {
	return nil
}

func GetRunByID(c echo.Context) error {
	param := c.Param("id")
	if param == "" {
		return c.JSON(400, map[string]string{"message": "id parameter is required"})
	}
	return nil
}