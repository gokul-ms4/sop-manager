package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gokul-ms4/sop-manager/common/models"
	"github.com/gokul-ms4/sop-manager/config"
	"github.com/labstack/echo/v4"
)

func CreateSopHeading(c echo.Context) error {
	userID := config.GetUserID(c)
	var sop models.SopHeading
	if err := c.Bind(&sop); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid input data",
		})
	}
	if sop.Heading == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Sop Heading cannot be empty",
		})
	}
	var existingHeading models.SopHeading
	heading := strings.TrimSpace(strings.ToLower(sop.Heading))
	if err := config.DB.Where("LOWER(heading) = ?", heading).First(&existingHeading).Error; err == nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Sop Heading already exists",
		})
	}
	sop.CreatedBy = userID
	if err := config.DB.Create(&sop).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "failed to create the Sop Heading",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"response": map[string]any{
			"message": "Sop Headint Created successfully",
			"data":    sop,
		},
	})
}

func UpdateSopHeading(c echo.Context) error {
	userID := config.GetUserID(c)
	headingID := c.Param("heading_id")
	var sop models.SopHeading
	if err := config.DB.First(&sop, headingID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "Failed to find the sop heading",
		})
	}
	if sop.CreatedBy != userID {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Only the creator can updte the sop",
		})
	}
	var input models.SopHeading
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid input data",
		})
	}
	if input.Heading != "" {
		var existing models.SopHeading
		heading := strings.TrimSpace(strings.ToLower(input.Heading))
		if err := config.DB.Where("LOWER(heading) = ? AND id != ?", heading, headingID).First(&existing).Error; err == nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "Sop heading already exists",
			})
		}
		sop.Heading = input.Heading
	}
	if input.Description != "" {
		sop.Description = input.Description
	}
	if err := config.DB.Save(&sop).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "FAiled to update the SOP heading",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"response": map[string]any{
			"message": "Sop heading updated successfully",
			"data":    sop,
		},
	})
}

func CreateSopItem(c echo.Context) error {
	userID := config.GetUserID(c)
	headingID := c.Param("heading_id")
	var sop models.SopItem
	if err := c.Bind(&sop); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid input data",
		})
	}
	var heading models.SopHeading
	if err := config.DB.First(&heading, headingID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "Failed to find the Sop Heading",
		})
	}
	if sop.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Sop Item title required",
		})
	}
	if sop.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Sop Item content required",
		})
	}
	var existingTitle models.SopItem
	title := strings.TrimSpace(strings.ToLower(sop.Title))
	if err := config.DB.Where("heading_id = ? AND LOWER(title) = ?", heading.ID, title).First(&existingTitle).Error; err == nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Sop Item Title already exists",
		})
	}
	if sop.Position == nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Sop position should be proovided",
		})
	}
	var existing models.SopItem
	if err := config.DB.Model(&models.SopItem{}).Where("heading_id = ? AND position = ?", heading.ID, sop.Position).First(&existing).Error; err == nil {
		var maxValue int
		config.DB.Model(&models.SopItem{}).
			Where("heading_id = ?", heading.ID).
			Select("COALESCE(MAX(position), 0)").
			Scan(&maxValue)
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("position already taken, available position = %d", maxValue+1),
		})
	}
	sop.CreatedBy = userID
	sop.HeadingID = heading.ID
	if err := config.DB.Create(&sop).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to create the Sop Item",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"response": map[string]any{
			"message": "Sop Item created successfully",
			"data":    sop,
		},
	})
}

func UpdateSopItem(c echo.Context) error {
	userID := config.GetUserID(c)
	headingID := c.Param("heading_id")
	itemID := c.Param("item_id")
	var heading models.SopHeading
	if err := config.DB.First(&heading, headingID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "Failed to find the sop heading",
		})
	}
	var sop models.SopItem
	if err := config.DB.Where("id = ? AND heading_id = ?", itemID, headingID).First(&sop).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "Failed to find the sop item",
		})
	}
	if sop.CreatedBy != userID {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Only the creator can updte the sop",
		})
	}
	type SopInput struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var input SopInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid input data",
		})
	}
	if input.Title != "" {
		var existing models.SopItem
		title := strings.TrimSpace(strings.ToLower(input.Title))
		if err := config.DB.Where("LOWER(title) = ? AND id != ?", title, itemID).First(&existing).Error; err == nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "Sop Title already exists",
			})
		}
		sop.Title = input.Title
	}
	if input.Content != "" {
		sop.Content = input.Content
	}
	if err := config.DB.Save(&sop).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to update the SOP heading",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"response": map[string]any{
			"message": "Sop Item updated successfully",
			"data":    sop,
		},
	})
}

func DeleteSopHeading(c echo.Context) error {
	userID := config.GetUserID(c)
	headingID := c.Param("heading_id")
	var sop models.SopHeading
	if err := config.DB.Model(&models.SopHeading{}).First(&sop, headingID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "Failed to find the sop heading",
		})
	}
	if sop.CreatedBy != userID {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Only the Creator can delete the sop heading",
		})
	}
	if err := config.DB.Delete(&sop).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to delete the sop heading",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "SOP heading deleted successfully",
	})
}

func DeleteSopItem(c echo.Context) error {
	userID := config.GetUserID(c)
	headingID := c.Param("heading_id")
	itemID := c.Param("item_id")
	var sopHeading models.SopHeading
	if err := config.DB.Model(&models.SopHeading{}).First(&sopHeading, headingID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "Failed to find the sop heading",
		})
	}
	var sopItem models.SopItem
	if err := config.DB.Model(&models.SopItem{}).Where("id = ? AND heading_id = ?", itemID, headingID).First(&sopItem).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "Failed to find the sop item",
		})
	}
	if sopItem.CreatedBy != userID {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Only the Creator can delete the sop item",
		})
	}
	if err := config.DB.Delete(&sopItem).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to delete the sop item",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "SOP item deleted successfully",
	})
}

func GetAllSopHeadings(c echo.Context) error {
	var sops []models.SopHeading
	if err := config.DB.Preload("SopItems").Find(&sops).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to fetch the sop headings",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"response": map[string]any{
			"message": "Sop Headings listed successfully",
			"data":    sops,
		},
	})
}

func GetSopHeadingByID(c echo.Context) error {
	headingID := c.Param("heading_id")
	var sop models.SopHeading
	if err := config.DB.Preload("SopItems").First(headingID, &sop).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to find the sop heading",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"response": map[string]any{
			"message": "Sop Heading listed successfully",
			"data":    sop,
		},
	})
}

func GetSopItems(c echo.Context) error {
	headingID := c.Param("heading_id")
	var sops []models.SopHeading
	if err := config.DB.Where("heading_id = ?", headingID).Find(&sops).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to fetch the sop items",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"response": map[string]any{
			"message": "Sop items listed successfully",
			"data":    sops,
		},
	})
}

func GenerateSopChunk(c echo.Context) error {
	var heading models.SopHeading
	headingID := c.Param("heading_id")
	if err := config.DB.Preload("SopItems").First(&heading, headingID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "Failed to find the sop heading",
		})
	}
	if err := config.DB.
		Where("sop_heading_id = ?", heading.ID).
		Delete(&models.SopChunk{}).Error; err != nil {

		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to clear old chunks",
		})
	}
	for _, item := range heading.SopItems {
		chunkText := fmt.Sprintf(
			"SOP Heading: %s\nSOP Item: %s\nContent: %s",
			heading.Heading,
			item.Title,
			item.Content,
		)
		embedding, err := config.GenerateEmbedding(chunkText)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "Failed to generate embedding",
			})
		}

		vectorString := config.Float32SliceToVectorString(embedding)
		chunk := models.SopChunk{
			SopHeadingID: uint(heading.ID),
			SopItemID:    uint(item.ID),
			ChunkText:    chunkText,
			Embedding:    vectorString,
		}
		if err := config.DB.Create(&chunk).Error; err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "Failed to create sop chunks",
			})
		}
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "Sop chunk added successfully",
	})
}
func AskSopQuestion(c echo.Context) error {
	type Request struct {
		Question string `json:"question"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request",
		})
	}
	if strings.TrimSpace(req.Question) == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Question is required",
		})
	}
	embedding, err := config.GenerateEmbedding(req.Question)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to generate question embedding",
		})
	}
	vectorString := config.Float32SliceToVectorString(embedding)
	type SearchResult struct {
		models.SopChunk
		Distance float64 `json:"distance"`
	}
	var chunks []SearchResult
	err = config.DB.Raw(`
	SELECT *,
	       embedding <=> ?::vector AS distance
	FROM sop_chunks
	ORDER BY distance
	LIMIT 3
`, vectorString).Scan(&chunks).Error
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to search SOP chunks",
		})
	}
	if len(chunks) == 0 || chunks[0].Distance > 0.7 {
		return c.JSON(http.StatusOK, map[string]any{
			"success": true,
			"answer":  "I could not find this information in the SOP.",
		})
	}
	var contextText string
	for _, chunk := range chunks {
		contextText += chunk.ChunkText + "\n\n"
	}
	answer, err := config.GenerateAnswer(req.Question, contextText)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Failed to generate answer",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success":  true,
		"question": req.Question,
		"answer":   answer,
		// "sources":  chunks,
	})
}
