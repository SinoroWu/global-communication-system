package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"global-comm-system/internal/models"
)

type TelegramAdapter struct {
	httpClient *http.Client
}

func NewTelegramAdapter() *TelegramAdapter {
	return &TelegramAdapter{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type TelegramUpdate struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID           int64  `json:"id"`
			IsBot        bool   `json:"is_bot"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Username     string `json:"username"`
			LanguageCode string `json:"language_code"`
		} `json:"from"`
		Chat struct {
			ID       int64  `json:"id"`
			Type     string `json:"type"`
			Title    string `json:"title"`
			Username string `json:"username"`
		} `json:"chat"`
		Date  int64  `json:"date"`
		Text  string `json:"text"`
		Photo []struct {
			FileID   string `json:"file_id"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
			FileSize int    `json:"file_size"`
		} `json:"photo"`
	} `json:"message"`
}

func (a *TelegramAdapter) ParseWebhook(body []byte) ([]*models.Message, error) {
	var update TelegramUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		return nil, fmt.Errorf("telegram unmarshal err: %w", err)
	}

	if update.Message == nil {
		return nil, nil // Ignore non-message updates (e.g. inline queries)
	}

	m := update.Message
	senderID := strconv.FormatInt(m.From.ID, 10)
	senderName := m.From.FirstName
	if m.From.LastName != "" {
		senderName += " " + m.From.LastName
	}
	if senderName == "" {
		senderName = m.From.Username
	}
	if senderName == "" {
		senderName = fmt.Sprintf("TG 用戶 (%s)", senderID)
	}

	cType := models.ContentTypeText
	content := m.Text
	if len(m.Photo) > 0 {
		cType = models.ContentTypeImage
		if content == "" {
			content = "[Telegram 相片]"
		}
	}

	msg := &models.Message{
		ID:          fmt.Sprintf("tg_%d", m.MessageID),
		Platform:    models.PlatformTelegram,
		SenderID:    senderID,
		SenderName:  senderName,
		SenderRole:  models.RoleUser,
		Avatar:      "https://api.dicebear.com/7.x/identicon/svg?seed=" + senderID,
		Content:     content,
		ContentType: cType,
		Status:      models.StatusDelivered,
		Timestamp:   m.Date * 1000,
		RawPayload:  string(body),
	}

	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}

	return []*models.Message{msg}, nil
}

func (a *TelegramAdapter) SendMessage(botToken, chatIDStr, text string) error {
	if botToken == "" || strings.Contains(botToken, "_demo") {
		log.Printf("[Telegram Adapter] Simulated outbound message to %s: %s", chatIDStr, text)
		return nil
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat id %s: %w", chatIDStr, err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	reqBody := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}

	data, _ := json.Marshal(reqBody)
	resp, err := a.httpClient.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram api error (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}
