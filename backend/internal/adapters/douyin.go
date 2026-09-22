package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"global-comm-system/internal/models"
)

type DouyinAdapter struct {
	httpClient *http.Client
}

func NewDouyinAdapter() *DouyinAdapter {
	return &DouyinAdapter{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type DouyinWebhookPayload struct {
	Event      string `json:"event"`
	ClientKey  string `json:"client_key"`
	CreateTime int64  `json:"create_time"`
	Content    struct {
		OpenID       string `json:"open_id"`
		Conversation string `json:"conversation_short_id"`
		ServerMsgID  string `json:"server_msg_id"`
		MsgType      string `json:"msg_type"` // "text", "image", "video"
		Text         string `json:"text"`
		MediaURL     string `json:"media_url,omitempty"`
	} `json:"content"`
}

func (a *DouyinAdapter) ParseWebhook(body []byte) ([]*models.Message, error) {
	var payload DouyinWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("douyin unmarshal err: %w", err)
	}

	cType := models.ContentTypeText
	text := payload.Content.Text
	switch payload.Content.MsgType {
	case "image":
		cType = models.ContentTypeImage
		if text == "" {
			text = "[抖音圖片]"
		}
	case "video":
		cType = models.ContentTypeVideo
		if text == "" {
			text = "[抖音短影音]"
		}
	}

	msgID := payload.Content.ServerMsgID
	if msgID == "" {
		msgID = fmt.Sprintf("dy_%d", time.Now().UnixNano())
	}

	msg := &models.Message{
		ID:          msgID,
		Platform:    models.PlatformDouyin,
		SenderID:    payload.Content.OpenID,
		SenderName:  fmt.Sprintf("抖音用戶 (%s)", truncateStr(payload.Content.OpenID, 6)),
		SenderRole:  models.RoleUser,
		Avatar:      "https://api.dicebear.com/7.x/personas/svg?seed=" + payload.Content.OpenID,
		Content:     text,
		ContentType: cType,
		MediaURL:    payload.Content.MediaURL,
		Status:      models.StatusDelivered,
		Timestamp:   payload.CreateTime,
		RawPayload:  string(body),
	}

	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}

	return []*models.Message{msg}, nil
}

func (a *DouyinAdapter) SendMessage(accessToken, openID, text string) error {
	if accessToken == "" || strings.Contains(accessToken, "_demo") {
		log.Printf("[Douyin Adapter] Simulated outbound message to %s: %s", openID, text)
		return nil
	}

	url := "https://open.douyin.com/im/send/msg/"
	reqBody := map[string]any{
		"to_user_id": openID,
		"msg_type":   "text",
		"content": map[string]string{
			"text": text,
		},
	}

	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access-token", accessToken)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("douyin api error (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}
