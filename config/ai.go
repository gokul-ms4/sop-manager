package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"google.golang.org/genai"
)

func TestGemini(c echo.Context) error {

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error1":  err.Error(),
		})
	}

	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text("Hello Gemini"),
		nil,
	)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error2":  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":  true,
		"response": resp.Text(),
	})
}

func TestEmbedding(c echo.Context) error {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
	}

	text := "SOP Heading: Leave Application Process\nSOP Item: Manager Approval\nContent: Reporting Manager reviews and approves leave requests."

	resp, err := client.Models.EmbedContent(ctx, "gemini-embedding-001", genai.Text(text), nil)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
	}

	embedding := resp.Embeddings[0].Values

	sampleSize := 5
	if len(embedding) < sampleSize {
		sampleSize = len(embedding)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":    true,
		"dimensions": len(embedding),
		"sample":     embedding[:sampleSize],
	})
}

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
