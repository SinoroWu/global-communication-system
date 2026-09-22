package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"global-comm-system/internal/models"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewStorage(dbPath string) (*Storage, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Enable WAL mode and performance pragmas
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Set connection pool for SQLite
	db.SetMaxOpenConns(1) // SQLite works best with serialized writes or WAL
	db.SetMaxIdleConns(1)

	s := &Storage{db: db}
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init database schema: %w", err)
	}

	return s, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		platform TEXT NOT NULL,
		external_user_id TEXT NOT NULL,
		display_name TEXT NOT NULL,
		avatar_url TEXT,
		last_message TEXT,
		last_message_at INTEGER,
		unread_count INTEGER DEFAULT 0,
		tags TEXT,
		notes TEXT,
		created_at INTEGER,
		updated_at INTEGER,
		UNIQUE(platform, external_user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_conversations_platform ON conversations(platform);
	CREATE INDEX IF NOT EXISTS idx_conversations_updated_at ON conversations(updated_at DESC);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		conversation_id TEXT NOT NULL,
		platform TEXT NOT NULL,
		sender_id TEXT NOT NULL,
		sender_name TEXT NOT NULL,
		sender_role TEXT NOT NULL,
		avatar TEXT,
		content TEXT NOT NULL,
		content_type TEXT NOT NULL,
		media_url TEXT,
		status TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		raw_payload TEXT,
		FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conv_timestamp ON messages(conversation_id, timestamp ASC);

	CREATE TABLE IF NOT EXISTS channel_configs (
		platform TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		enabled INTEGER DEFAULT 1,
		webhook_secret TEXT,
		access_token TEXT,
		channel_id TEXT,
		app_secret TEXT,
		extra_json TEXT
	);
	`
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}

	// Seed default channel configurations if empty
	return s.seedDefaultChannels()
}

func (s *Storage) seedDefaultChannels() error {
	defaults := []models.ChannelConfig{
		{
			Platform:      models.PlatformLine,
			Name:          "LINE 官方帳號",
			Enabled:       true,
			WebhookURL:    "/api/webhooks/line",
			WebhookSecret: "line_channel_secret_demo",
			AccessToken:   "line_channel_access_token_demo",
			ChannelID:     "@my_line_oa",
		},
		{
			Platform:      models.PlatformInstagram,
			Name:          "Instagram Direct",
			Enabled:       true,
			WebhookURL:    "/api/webhooks/instagram",
			WebhookSecret: "meta_verify_token_demo",
			AccessToken:   "ig_graph_access_token_demo",
			AppSecret:     "meta_app_secret_demo",
			ChannelID:     "instagram_business_demo",
		},
		{
			Platform:      models.PlatformFacebook,
			Name:          "Facebook Messenger",
			Enabled:       true,
			WebhookURL:    "/api/webhooks/facebook",
			WebhookSecret: "fb_verify_token_demo",
			AccessToken:   "fb_page_access_token_demo",
			AppSecret:     "fb_app_secret_demo",
			ChannelID:     "fb_page_10023456",
		},
		{
			Platform:      models.PlatformDouyin,
			Name:          "抖音 / TikTok 企業私訊",
			Enabled:       true,
			WebhookURL:    "/api/webhooks/douyin",
			WebhookSecret: "douyin_webhook_token_demo",
			AccessToken:   "douyin_access_token_demo",
			ChannelID:     "douyin_creator_enterprise",
		},
		{
			Platform:      models.PlatformTelegram,
			Name:          "Telegram 機器人",
			Enabled:       true,
			WebhookURL:    "/api/webhooks/telegram",
			WebhookSecret: "telegram_secret_token_demo",
			AccessToken:   "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ_demo",
			ChannelID:     "@GlobalCommBot",
		},
	}

	for _, c := range defaults {
		var count int
		s.db.QueryRow("SELECT COUNT(*) FROM channel_configs WHERE platform = ?", c.Platform).Scan(&count)
		if count == 0 {
			extra, _ := json.Marshal(c.Extra)
			s.db.Exec(`INSERT INTO channel_configs (platform, name, enabled, webhook_secret, access_token, channel_id, app_secret, extra_json)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				c.Platform, c.Name, 1, c.WebhookSecret, c.AccessToken, c.ChannelID, c.AppSecret, string(extra))
		}
	}
	return nil
}

// GetOrCreateConversation retrieves existing or inserts new conversation
func (s *Storage) GetOrCreateConversation(platform models.Platform, extUserID, displayName, avatarURL string) (*models.Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var conv models.Conversation
	var tagsJSON sql.NullString
	var notes sql.NullString
	var avatar sql.NullString

	query := `SELECT id, platform, external_user_id, display_name, avatar_url, last_message, last_message_at, unread_count, tags, notes, created_at, updated_at 
		FROM conversations WHERE platform = ? AND external_user_id = ?`

	err := s.db.QueryRow(query, platform, extUserID).Scan(
		&conv.ID, &conv.Platform, &conv.ExternalUserID, &conv.DisplayName,
		&avatar, &conv.LastMessage, &conv.LastMessageAt, &conv.UnreadCount,
		&tagsJSON, &notes, &conv.CreatedAt, &conv.UpdatedAt,
	)

	now := time.Now().UnixMilli()

	if err == sql.ErrNoRows {
		// Create new conversation
		conv.ID = uuid.New().String()
		conv.Platform = platform
		conv.ExternalUserID = extUserID
		conv.DisplayName = displayName
		conv.AvatarURL = avatarURL
		conv.LastMessage = ""
		conv.LastMessageAt = now
		conv.UnreadCount = 0
		conv.Tags = []string{"新訪客"}
		conv.Notes = ""
		conv.CreatedAt = now
		conv.UpdatedAt = now

		tagsBytes, _ := json.Marshal(conv.Tags)
		insertQuery := `INSERT INTO conversations (id, platform, external_user_id, display_name, avatar_url, last_message, last_message_at, unread_count, tags, notes, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, insErr := s.db.Exec(insertQuery, conv.ID, conv.Platform, conv.ExternalUserID, conv.DisplayName, conv.AvatarURL, conv.LastMessage, conv.LastMessageAt, conv.UnreadCount, string(tagsBytes), conv.Notes, conv.CreatedAt, conv.UpdatedAt)
		if insErr != nil {
			return nil, insErr
		}
		return &conv, nil
	} else if err != nil {
		return nil, err
	}

	if avatar.Valid {
		conv.AvatarURL = avatar.String
	}
	if notes.Valid {
		conv.Notes = notes.String
	}
	if tagsJSON.Valid && tagsJSON.String != "" {
		_ = json.Unmarshal([]byte(tagsJSON.String), &conv.Tags)
	}

	// Update display name or avatar if updated
	if displayName != "" && displayName != conv.DisplayName {
		conv.DisplayName = displayName
		s.db.Exec("UPDATE conversations SET display_name = ?, updated_at = ? WHERE id = ?", displayName, now, conv.ID)
	}

	return &conv, nil
}

// SaveMessage stores message and updates conversation stats
func (s *Storage) SaveMessage(msg *models.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	msgQuery := `INSERT INTO messages (id, conversation_id, platform, sender_id, sender_name, sender_role, avatar, content, content_type, media_url, status, timestamp, raw_payload)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.Exec(msgQuery, msg.ID, msg.ConversationID, msg.Platform, msg.SenderID, msg.SenderName, msg.SenderRole, msg.Avatar, msg.Content, msg.ContentType, msg.MediaURL, msg.Status, msg.Timestamp, msg.RawPayload)
	if err != nil {
		return fmt.Errorf("insert message err: %w", err)
	}

	// Update conversation last message and unread count if sent by user
	unreadDelta := 0
	if msg.SenderRole == models.RoleUser {
		unreadDelta = 1
	}

	preview := msg.Content
	if msg.ContentType != models.ContentTypeText {
		preview = fmt.Sprintf("[%s]", string(msg.ContentType))
	}

	convQuery := `UPDATE conversations 
		SET last_message = ?, last_message_at = ?, unread_count = unread_count + ?, updated_at = ?
		WHERE id = ?`
	_, err = tx.Exec(convQuery, preview, msg.Timestamp, unreadDelta, msg.Timestamp, msg.ConversationID)
	if err != nil {
		return fmt.Errorf("update conv stats err: %w", err)
	}

	return tx.Commit()
}

// GetConversations lists conversations optionally filtered by platform
func (s *Storage) GetConversations(platformFilter models.Platform) ([]*models.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var query string
	var rows *sql.Rows
	var err error

	if platformFilter == "" || platformFilter == "all" {
		query = `SELECT id, platform, external_user_id, display_name, avatar_url, last_message, last_message_at, unread_count, tags, notes, created_at, updated_at
			FROM conversations ORDER BY updated_at DESC`
		rows, err = s.db.Query(query)
	} else {
		query = `SELECT id, platform, external_user_id, display_name, avatar_url, last_message, last_message_at, unread_count, tags, notes, created_at, updated_at
			FROM conversations WHERE platform = ? ORDER BY updated_at DESC`
		rows, err = s.db.Query(query, platformFilter)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*models.Conversation, 0)
	for rows.Next() {
		var c models.Conversation
		var tagsJSON, notes, avatar sql.NullString
		if err := rows.Scan(&c.ID, &c.Platform, &c.ExternalUserID, &c.DisplayName, &avatar, &c.LastMessage, &c.LastMessageAt, &c.UnreadCount, &tagsJSON, &notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			log.Printf("scan conversation row err: %v", err)
			continue
		}
		if avatar.Valid {
			c.AvatarURL = avatar.String
		}
		if notes.Valid {
			c.Notes = notes.String
		}
		if tagsJSON.Valid && tagsJSON.String != "" {
			_ = json.Unmarshal([]byte(tagsJSON.String), &c.Tags)
		} else {
			c.Tags = []string{}
		}
		list = append(list, &c)
	}
	return list, nil
}

// GetConversation returns single conversation
func (s *Storage) GetConversation(id string) (*models.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var c models.Conversation
	var tagsJSON, notes, avatar sql.NullString
	query := `SELECT id, platform, external_user_id, display_name, avatar_url, last_message, last_message_at, unread_count, tags, notes, created_at, updated_at
		FROM conversations WHERE id = ?`
	err := s.db.QueryRow(query, id).Scan(&c.ID, &c.Platform, &c.ExternalUserID, &c.DisplayName, &avatar, &c.LastMessage, &c.LastMessageAt, &c.UnreadCount, &tagsJSON, &notes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if avatar.Valid {
		c.AvatarURL = avatar.String
	}
	if notes.Valid {
		c.Notes = notes.String
	}
	if tagsJSON.Valid && tagsJSON.String != "" {
		_ = json.Unmarshal([]byte(tagsJSON.String), &c.Tags)
	} else {
		c.Tags = []string{}
	}
	return &c, nil
}

// GetMessages loads messages for conversation ordered chronologically
func (s *Storage) GetMessages(conversationID string, limit int) ([]*models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	query := `SELECT id, conversation_id, platform, sender_id, sender_name, sender_role, avatar, content, content_type, media_url, status, timestamp, raw_payload
		FROM messages WHERE conversation_id = ? ORDER BY timestamp ASC LIMIT ?`

	rows, err := s.db.Query(query, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*models.Message, 0)
	for rows.Next() {
		var m models.Message
		var mediaURL, rawPayload, avatar sql.NullString
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Platform, &m.SenderID, &m.SenderName, &m.SenderRole, &avatar, &m.Content, &m.ContentType, &mediaURL, &m.Status, &m.Timestamp, &rawPayload); err != nil {
			continue
		}
		if avatar.Valid {
			m.Avatar = avatar.String
		}
		if mediaURL.Valid {
			m.MediaURL = mediaURL.String
		}
		if rawPayload.Valid {
			m.RawPayload = rawPayload.String
		}
		list = append(list, &m)
	}
	return list, nil
}

// MarkConversationAsRead resets unread count
func (s *Storage) MarkConversationAsRead(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("UPDATE conversations SET unread_count = 0 WHERE id = ?", id)
	return err
}

// UpdateConversationNotesAndTags updates customer notes and tags
func (s *Storage) UpdateConversationNotesAndTags(id string, notes string, tags []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tagsBytes, _ := json.Marshal(tags)
	_, err := s.db.Exec("UPDATE conversations SET notes = ?, tags = ? WHERE id = ?", notes, string(tagsBytes), id)
	return err
}

// GetChannelConfigs returns all platform settings
func (s *Storage) GetChannelConfigs() ([]*models.ChannelConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT platform, name, enabled, webhook_secret, access_token, channel_id, app_secret, extra_json FROM channel_configs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*models.ChannelConfig, 0)
	for rows.Next() {
		var c models.ChannelConfig
		var enabledInt int
		var extraStr sql.NullString
		var secret, token, cid, appSec sql.NullString
		if err := rows.Scan(&c.Platform, &c.Name, &enabledInt, &secret, &token, &cid, &appSec, &extraStr); err != nil {
			continue
		}
		c.Enabled = enabledInt == 1
		if secret.Valid {
			c.WebhookSecret = secret.String
		}
		if token.Valid {
			c.AccessToken = token.String
		}
		if cid.Valid {
			c.ChannelID = cid.String
		}
		if appSec.Valid {
			c.AppSecret = appSec.String
		}
		c.WebhookURL = fmt.Sprintf("/api/webhooks/%s", c.Platform)
		if extraStr.Valid && extraStr.String != "" {
			_ = json.Unmarshal([]byte(extraStr.String), &c.Extra)
		}
		list = append(list, &c)
	}
	return list, nil
}

// SaveChannelConfig updates credentials
func (s *Storage) SaveChannelConfig(c *models.ChannelConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	extraBytes, _ := json.Marshal(c.Extra)
	enabled := 0
	if c.Enabled {
		enabled = 1
	}

	query := `INSERT INTO channel_configs (platform, name, enabled, webhook_secret, access_token, channel_id, app_secret, extra_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(platform) DO UPDATE SET
			name = excluded.name,
			enabled = excluded.enabled,
			webhook_secret = excluded.webhook_secret,
			access_token = excluded.access_token,
			channel_id = excluded.channel_id,
			app_secret = excluded.app_secret,
			extra_json = excluded.extra_json`

	_, err := s.db.Exec(query, c.Platform, c.Name, enabled, c.WebhookSecret, c.AccessToken, c.ChannelID, c.AppSecret, string(extraBytes))
	return err
}
