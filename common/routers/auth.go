package routers

import (
	"github.com/gokul-ms4/sop-manager/common/controllers"
	"github.com/labstack/echo/v4"
)

func AuthRoutes(e *echo.Echo) {
	auth := e.Group("/api/v1/auth")
	auth.POST("/register", controllers.Register)
	auth.POST("/login",controllers.Login)
}
