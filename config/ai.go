package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gokul-ms4/sop-manager/common/models"
	"google.golang.org/genai"
)

func Float32SliceToVectorString(values []float32) string {
	strValues := make([]string, len(values))

	for i, v := range values {
		strValues[i] = fmt.Sprintf("%f", v)
	}

	return "[" + strings.Join(strValues, ",") + "]"
}

func GenerateEmbedding(text string) ([]float32, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	resp, err := client.Models.EmbedContent(
		ctx,
		"gemini-embedding-001",
		genai.Text(text),
		nil,
	)
	if err != nil {
		return nil, err
	}

	return resp.Embeddings[0].Values, nil
}

func GenerateAnswer(question string, contextText string) (string, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`
You are an SOP assistant.
Answer the question using only the provided SOP context.
If the answer is not in the context, say: "I could not find this information in the SOP."

Context:
%s

Question:
%s
`, contextText, question)

	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", err
	}

	return resp.Text(), nil
}

func GenerateSopHeadingChunk(headingID int) error {
	var heading models.SopHeading
	if err := DB.
		Preload("SopItems").
		First(&heading, headingID).Error; err != nil {
		return err
	}
	if err := DB.
		Where("sop_heading_id = ?", heading.ID).
		Delete(&models.SopChunk{}).Error; err != nil {
		return err
	}
	for _, item := range heading.SopItems {
		chunkText := fmt.Sprintf(
			"SOP Heading: %s\nSOP Item: %s\nContent: %s",
			heading.Heading,
			item.Title,
			item.Content,
		)
		embedding, err := GenerateEmbedding(chunkText)
		if err != nil {
			return err
		}
		vectorString := Float32SliceToVectorString(embedding)
		chunk := models.SopChunk{
			SopHeadingID: uint(heading.ID),
			SopItemID:    uint(item.ID),
			ChunkText:    chunkText,
			Embedding:    vectorString,
		}
		if err := DB.Create(&chunk).Error; err != nil {
			return err
		}
	}
	return nil
}

func GenerateSopItemChunk(itemID int) error {
	var item models.SopItem
	if err := DB.
		First(&item, itemID).Error; err != nil {
		return err
	}
	if err := DB.
		Where("sop_item_id = ?", item.ID).
		Delete(&models.SopChunk{}).Error; err != nil {
		return err
	}
	var heading models.SopHeading
	if err := DB.
		First(&heading, item.HeadingID).Error; err != nil {
		return err
	}
	chunkText := fmt.Sprintf(
		"SOP Heading: %s\nSOP Item: %s\nContent: %s",
		heading.Heading,
		item.Title,
		item.Content,
	)
	embedding, err := GenerateEmbedding(chunkText)
	if err != nil {
		return err
	}
	vectorString := Float32SliceToVectorString(embedding)
	chunk := models.SopChunk{
		SopHeadingID: uint(heading.ID),
		SopItemID:    uint(item.ID),
		ChunkText:    chunkText,
		Embedding:    vectorString,
	}
	if err := DB.Create(&chunk).Error; err != nil {
		return err
	}
	return nil
}

func GetNextPosition(headingID int)(int, error){
	var position int
	var heading models.SopHeading
	if err := DB.First(&heading, headingID).Error; err != nil {
		return 0,err
	}
	if err := DB.Model(&models.SopItem{}).Where("heading_id = ?", heading.ID).
		Select("COALESCE(MAX(position), 0)").Scan(&position).Error; err != nil {
		return 0,err
	}
	return (position + 1), nil
}