package adapters

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"global-comm-system/internal/models"
)

type LineAdapter struct {
	httpClient *http.Client
}

func NewLineAdapter() *LineAdapter {
	return &LineAdapter{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// VerifySignature validates X-Line-Signature against channel secret
func (a *LineAdapter) VerifySignature(channelSecret, signature string, body []byte) bool {
	if channelSecret == "" {
		return true // Skip verification if not configured for dev
	}
	hash := hmac.New(sha256.New, []byte(channelSecret))
	hash.Write(body)
	expectedSignature := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

type LineWebhookPayload struct {
	Destination string `json:"destination"`
	Events      []struct {
		Type       string `json:"type"` // "message", "follow", "unfollow"
		ReplyToken string `json:"replyToken"`
		Timestamp  int64  `json:"timestamp"`
		Source     struct {
			Type   string `json:"type"`
			UserID string `json:"userId"`
		} `json:"source"`
		Message struct {
			ID        string  `json:"id"`
			Type      string  `json:"type"` // "text", "image", "sticker"
			Text      string  `json:"text"`
			PackageID string  `json:"packageId,omitempty"`
			StickerID string  `json:"stickerId,omitempty"`
			Duration  int     `json:"duration,omitempty"`
			Title     string  `json:"title,omitempty"`
			Address   string  `json:"address,omitempty"`
			Latitude  float64 `json:"latitude,omitempty"`
			Longitude float64 `json:"longitude,omitempty"`
		} `json:"message"`
	} `json:"events"`
}

// ParseWebhook parses LINE raw JSON into standard models
func (a *LineAdapter) ParseWebhook(body []byte) ([]*models.Message, error) {
	var payload LineWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal line payload: %w", err)
	}

	var results []*models.Message
	for _, event := range payload.Events {
		if event.Type != "message" {
			continue
		}

		cType := models.ContentTypeText
		content := event.Message.Text
		if event.Message.Type == "image" {
			cType = models.ContentTypeImage
			content = "[圖片訊息]"
		} else if event.Message.Type == "sticker" {
			cType = models.ContentTypeSticker
			content = "[貼圖]"
		}

		msg := &models.Message{
			ID:          event.Message.ID,
			Platform:    models.PlatformLine,
			SenderID:    event.Source.UserID,
			SenderName:  fmt.Sprintf("LINE 用戶 (%s)", truncateStr(event.Source.UserID, 6)),
			SenderRole:  models.RoleUser,
			Avatar:      "https://api.dicebear.com/7.x/bottts/svg?seed=" + event.Source.UserID,
			Content:     content,
			ContentType: cType,
			Status:      models.StatusDelivered,
			Timestamp:   event.Timestamp,
			RawPayload:  string(body),
		}
		results = append(results, msg)
	}
	return results, nil
}

// SendMessage sends reply or push message to LINE user
func (a *LineAdapter) SendMessage(accessToken, toUserID, text string) error {
	if accessToken == "" || accessToken == "line_channel_access_token_demo" {
		log.Printf("[LINE Adapter] Simulated outbound message to %s: %s", toUserID, text)
		return nil // Success in simulation mode
	}

	url := "https://api.line.me/v2/bot/message/push"
	reqBody := map[string]any{
		"to": toUserID,
		"messages": []map[string]string{
			{
				"type": "text",
				"text": text,
			},
		},
	}

	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("line api error (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
