// internal/infrastructure/discord/bot.go

package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"fluencybe/pkg/chatbot"
	"fluencybe/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session    *discordgo.Session
	logger     *logger.PrettyLogger
	roleID     string
	webhookURL string
	chatbot    *chatbot.TogetherAIClient
}

func NewBot(logger *logger.PrettyLogger) (*Bot, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	roleID := os.Getenv("DISCORD_BOT_ROLE_ID")
	webhookURL := os.Getenv("DISCORD_CHATBOT_WEBBOOK_URL")

	if token == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN not set")
	}
	if roleID == "" {
		return nil, fmt.Errorf("DISCORD_BOT_ROLE_ID not set")
	}
	if webhookURL == "" {
		return nil, fmt.Errorf("DISCORD_CHATBOT_WEBBOOK_URL not set")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	bot := &Bot{
		session:    session,
		logger:     logger,
		roleID:     roleID,
		webhookURL: webhookURL,
		chatbot:    chatbot.NewTogetherAIClient(logger),
	}

	// Add handlers
	session.AddHandler(bot.messageHandler)

	return bot, nil
}

func (b *Bot) Start() error {
	// Open a websocket connection to Discord
	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("error opening connection to Discord: %w", err)
	}

	b.logger.Info("DISCORD_BOT_START", map[string]interface{}{
		"username": b.session.State.User.Username,
	}, "Discord bot started successfully")

	return nil
}

func (b *Bot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check for role mention
	roleMention := fmt.Sprintf("<@&%s>", b.roleID)
	if strings.Contains(m.Content, roleMention) {
		// Extract the actual content by removing the role mention
		content := strings.TrimSpace(strings.ReplaceAll(m.Content, roleMention, ""))

		// Handle Postman commands
		if strings.HasPrefix(content, "open") || strings.HasPrefix(content, "help") ||
			strings.HasPrefix(content, "GET") || strings.HasPrefix(content, "POST") ||
			strings.HasPrefix(content, "PUT") || strings.HasPrefix(content, "DELETE") {
			return
		}

		// Process chatbot queries
		if err := b.chatbot.ProcessQuery(content); err != nil {
			b.logger.Error("CHATBOT_PROCESS", map[string]interface{}{
				"error": err.Error(),
				"query": content,
			}, "Failed to process chatbot query")
			b.sendWebhookResponse("Sorry, I encountered an error processing your request.")
		}
	}
}

func (b *Bot) sendWebhookResponse(content string) {
	webhookURL := os.Getenv("DISCORD_POSTMAN_WEBHOOK_URL")
	if webhookURL == "" {
		b.logger.Error("DISCORD_WEBHOOK_URL", nil, "DISCORD_POSTMAN_WEBHOOK_URL not set")
		return
	}

	message := struct {
		Content string `json:"content"`
	}{
		Content: content,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		b.logger.Error("DISCORD_WEBHOOK_MARSHAL", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to marshal webhook message")
		return
	}

	// Use direct HTTP POST for regular messages
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		b.logger.Error("DISCORD_WEBHOOK_SEND", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to send webhook message")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b.logger.Error("DISCORD_WEBHOOK_RESPONSE", map[string]interface{}{
			"status_code": resp.StatusCode,
		}, "Unexpected webhook response status")
	}
}

func (b *Bot) sendInteractiveMessage(embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	webhookURL := os.Getenv("DISCORD_POSTMAN_WEBHOOK_URL")
	if webhookURL == "" {
		b.logger.Error("DISCORD_WEBHOOK_URL", nil, "DISCORD_POSTMAN_WEBHOOK_URL not set")
		return
	}

	// Extract webhook ID and token from URL
	parts := strings.Split(webhookURL, "/")
	if len(parts) < 2 {
		b.logger.Error("DISCORD_WEBHOOK_URL", nil, "Invalid webhook URL format")
		return
	}

	webhookID := parts[len(parts)-2]
	webhookToken := parts[len(parts)-1]

	// Create webhook message with components
	message := &discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	}

	// Execute webhook with wait parameter
	_, err := b.session.WebhookExecute(webhookID, webhookToken, true, message)
	if err != nil {
		b.logger.Error("DISCORD_WEBHOOK_SEND", map[string]interface{}{
			"error": err.Error(),
			"id":    webhookID,
		}, "Failed to send interactive message")
		return
	}

	b.logger.Info("DISCORD_WEBHOOK_SEND", map[string]interface{}{
		"id": webhookID,
	}, "Successfully sent interactive message")
}

func (b *Bot) updateInteractiveMessage(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	}
	s.InteractionRespond(i.Interaction, response)
}
