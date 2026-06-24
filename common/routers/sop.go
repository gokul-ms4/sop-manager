package routers

import (
	"github.com/gokul-ms4/sop-manager/common/controllers"
	"github.com/labstack/echo/v4"
)

func SopRoutes(protected *echo.Group) {
	sop := protected.Group("/sop_heading")
	sop.GET("", controllers.GetAllSopHeadings)
	sop.POST("", controllers.CreateSopHeading)
	sop.GET("/:heading_id", controllers.GetSopHeadingByID)
	sop.PATCH("/:heading_id", controllers.UpdateSopHeading)
	sop.DELETE("/:heading_id", controllers.DeleteSopHeading)

	sop.POST("/:heading_id/sop_item", controllers.CreateSopItem)
	sop.GET("/:heading_id/list_items", controllers.GetSopItems)
	sop.PATCH("/:heading_id/sop_item/:item_id", controllers.UpdateSopItem)
	sop.DELETE("/:heading_id/sop_item/:item_id", controllers.DeleteSopItem)

	sop.GET("/:heading_id/generate_sop_chunk", controllers.GenerateSopChunk)
	sop.POST("/ask_question", controllers.AskSopQuestion)

}
