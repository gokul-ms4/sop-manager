package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// "github.com/gokul-ms4/sop-manager/common/models"
	"github.com/gokul-ms4/sop-manager/common/routers"
	"github.com/gokul-ms4/sop-manager/config"
)

func main() {
	godotenv.Load() // ✅ just ignore the error

	config.ConnectDB()

	e := echo.New()

	e.Server.MaxHeaderBytes = 100 << 20
	e.Use(middleware.BodyLimit("100M"))

	// MUST be before InitRoutes so CORS covers static files
	routers.InitRoutes(e)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on :" + port)
	e.Logger.Fatal(e.Start(":" + port))
}
