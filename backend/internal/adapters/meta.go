package adapters

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"global-comm-system/internal/models"
)

type MetaAdapter struct {
	httpClient *http.Client
}

func NewMetaAdapter() *MetaAdapter {
	return &MetaAdapter{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// VerifySignature validates X-Hub-Signature-256 against app secret
func (a *MetaAdapter) VerifySignature(appSecret, signatureHeader string, body []byte) bool {
	if appSecret == "" || strings.HasPrefix(appSecret, "meta_") {
		return true // Demo mode pass
	}
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}
	expectedHash := signatureHeader[7:]
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	calculatedHash := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(calculatedHash), []byte(expectedHash))
}

type MetaWebhookPayload struct {
	Object string `json:"object"` // "page" (FB) or "instagram" (IG)
	Entry  []struct {
		ID        string `json:"id"`
		Time      int64  `json:"time"`
		Messaging []struct {
			Sender struct {
				ID string `json:"id"`
			} `json:"sender"`
			Recipient struct {
				ID string `json:"id"`
			} `json:"recipient"`
			Timestamp int64 `json:"timestamp"`
			Message   *struct {
				MID  string `json:"mid"`
				Text string `json:"text"`
				Seq  int    `json:"seq"`
				Attachments []struct {
					Type    string `json:"type"` // "image", "video", "audio", "file"
					Payload struct {
						URL string `json:"url"`
					} `json:"payload"`
				} `json:"attachments"`
			} `json:"message"`
		} `json:"messaging"`
	} `json:"entry"`
}

// ParseWebhook converts Meta webhook payload into standardized messages
func (a *MetaAdapter) ParseWebhook(body []byte, defaultPlatform models.Platform) ([]*models.Message, error) {
	var payload MetaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("meta unmarshal err: %w", err)
	}

	platform := defaultPlatform
	if payload.Object == "instagram" {
		platform = models.PlatformInstagram
	} else if payload.Object == "page" {
		platform = models.PlatformFacebook
	}

	var results []*models.Message
	for _, entry := range payload.Entry {
		for _, m := range entry.Messaging {
			if m.Message == nil {
				continue
			}

			cType := models.ContentTypeText
			content := m.Message.Text
			var mediaURL string

			if len(m.Message.Attachments) > 0 {
				att := m.Message.Attachments[0]
				mediaURL = att.Payload.URL
				switch att.Type {
				case "image":
					cType = models.ContentTypeImage
					content = "[圖片]"
				case "video":
					cType = models.ContentTypeVideo
					content = "[影片]"
				case "audio":
					cType = models.ContentTypeAudio
					content = "[語音]"
				default:
					cType = models.ContentTypeFile
					content = "[檔案]"
				}
			}

			platformName := "Facebook"
			if platform == models.PlatformInstagram {
				platformName = "Instagram"
			}

			msg := &models.Message{
				ID:          m.Message.MID,
				Platform:    platform,
				SenderID:    m.Sender.ID,
				SenderName:  fmt.Sprintf("%s 用戶 (%s)", platformName, truncateStr(m.Sender.ID, 6)),
				SenderRole:  models.RoleUser,
				Avatar:      "https://api.dicebear.com/7.x/avataaars/svg?seed=" + m.Sender.ID,
				Content:     content,
				ContentType: cType,
				MediaURL:    mediaURL,
				Status:      models.StatusDelivered,
				Timestamp:   m.Timestamp,
				RawPayload:  string(body),
			}
			results = append(results, msg)
		}
	}
	return results, nil
}

// SendMessage sends message via Meta Graph API
func (a *MetaAdapter) SendMessage(accessToken, recipientID, text string) error {
	if accessToken == "" || strings.Contains(accessToken, "_demo") {
		log.Printf("[Meta Adapter] Simulated outbound message to %s: %s", recipientID, text)
		return nil
	}

	url := "https://graph.facebook.com/v21.0/me/messages"
	reqBody := map[string]any{
		"recipient": map[string]string{"id": recipientID},
		"message":   map[string]string{"text": text},
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
		return fmt.Errorf("meta api error (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}
