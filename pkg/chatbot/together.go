package chatbot

import (
	"bytes"
	"encoding/json"
	"fluencybe/pkg/logger"
	"fmt"
	"net/http"
	"os"
)

type TogetherAIClient struct {
	apiKey     string
	webhookURL string
	logger     *logger.PrettyLogger
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewTogetherAIClient(logger *logger.PrettyLogger) *TogetherAIClient {
	return &TogetherAIClient{
		apiKey:     os.Getenv("TOGETHERAI_API_KEY"),
		webhookURL: os.Getenv("DISCORD_CHATBOT_WEBBOOK_URL"),
		logger:     logger,
	}
}

func (c *TogetherAIClient) ProcessQuery(query string) error {
	// Create chat request
	req := ChatRequest{
		Model: "meta-llama/Llama-3.3-70B-Instruct-Turbo-Free",
		Messages: []ChatMessage{
			{Role: "user", Content: query},
		},
	}

	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", "https://api.together.xyz/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Send response to Discord webhook
	if len(chatResp.Choices) > 0 {
		content := chatResp.Choices[0].Message.Content
		webhookMsg := struct {
			Content string `json:"content"`
		}{
			Content: content,
		}

		webhookBody, err := json.Marshal(webhookMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal webhook message: %w", err)
		}

		webhookResp, err := http.Post(c.webhookURL, "application/json", bytes.NewBuffer(webhookBody))
		if err != nil {
			return fmt.Errorf("failed to send webhook: %w", err)
		}
		defer webhookResp.Body.Close()

		if webhookResp.StatusCode != http.StatusOK && webhookResp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("webhook returned status %d", webhookResp.StatusCode)
		}
	}

	return nil
}
