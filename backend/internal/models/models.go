package models

type Platform string

const (
	PlatformLine      Platform = "line"
	PlatformInstagram Platform = "instagram"
	PlatformFacebook  Platform = "facebook"
	PlatformDouyin    Platform = "douyin"
	PlatformTelegram  Platform = "telegram"
)

type SenderRole string

const (
	RoleUser   SenderRole = "user"
	RoleAgent  SenderRole = "agent"
	RoleSystem SenderRole = "system"
)

type ContentType string

const (
	ContentTypeText    ContentType = "text"
	ContentTypeImage   ContentType = "image"
	ContentTypeVideo   ContentType = "video"
	ContentTypeAudio   ContentType = "audio"
	ContentTypeFile    ContentType = "file"
	ContentTypeSticker ContentType = "sticker"
)

type MessageStatus string

const (
	StatusSent      MessageStatus = "sent"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
	StatusFailed    MessageStatus = "failed"
)

// Message represents a standardized message across all platforms
type Message struct {
	ID             string        `json:"id"`
	ConversationID string        `json:"conversationId"`
	Platform       Platform      `json:"platform"`
	SenderID       string        `json:"senderId"`
	SenderName     string        `json:"senderName"`
	SenderRole     SenderRole    `json:"senderRole"`
	Avatar         string        `json:"avatar"`
	Content        string        `json:"content"`
	ContentType    ContentType   `json:"contentType"`
	MediaURL       string        `json:"mediaUrl,omitempty"`
	Status         MessageStatus `json:"status"`
	Timestamp      int64         `json:"timestamp"`
	RawPayload     string        `json:"rawPayload,omitempty"`
}

// Conversation represents an aggregated customer dialog
type Conversation struct {
	ID             string   `json:"id"`
	Platform       Platform `json:"platform"`
	ExternalUserID string   `json:"externalUserId"`
	DisplayName    string   `json:"displayName"`
	AvatarURL      string   `json:"avatarUrl"`
	LastMessage    string   `json:"lastMessage"`
	LastMessageAt  int64    `json:"lastMessageAt"`
	UnreadCount    int      `json:"unreadCount"`
	Tags           []string `json:"tags"`
	Notes          string   `json:"notes"`
	CreatedAt      int64    `json:"createdAt"`
	UpdatedAt      int64    `json:"updatedAt"`
}

// ChannelConfig represents platform-specific API credentials & settings
type ChannelConfig struct {
	Platform      Platform          `json:"platform"`
	Name          string            `json:"name"`
	Enabled       bool              `json:"enabled"`
	WebhookURL    string            `json:"webhookUrl"`
	WebhookSecret string            `json:"webhookSecret"`
	AccessToken   string            `json:"accessToken"`
	ChannelID     string            `json:"channelId"`
	AppSecret     string            `json:"appSecret,omitempty"`
	Extra         map[string]string `json:"extra,omitempty"`
}

// WSMessage represents a real-time event pushed to Vue clients
type WSMessage struct {
	Type string `json:"type"` // "new_message", "conversation_updated", "channel_status", "system_notification"
	Data any    `json:"data"`
}

// OutgoingMessageRequest represents client request to send a message
type OutgoingMessageRequest struct {
	ConversationID string      `json:"conversationId"`
	Content        string      `json:"content"`
	ContentType    ContentType `json:"contentType"`
	MediaURL       string      `json:"mediaUrl,omitempty"`
}

// SimulatorRequest for triggering simulated webhook messages
type SimulatorRequest struct {
	Platform    Platform `json:"platform"`
	SenderID    string   `json:"senderId"`
	SenderName  string   `json:"senderName"`
	AvatarURL   string   `json:"avatarUrl"`
	Content     string   `json:"content"`
	ContentType string   `json:"contentType"`
	MediaURL    string   `json:"mediaUrl"`
}
