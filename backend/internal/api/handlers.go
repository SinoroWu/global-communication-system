package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"global-comm-system/internal/adapters"
	"global-comm-system/internal/hub"
	"global-comm-system/internal/models"
	"global-comm-system/internal/storage"

	"github.com/google/uuid"
)

type Server struct {
	storage         *storage.Storage
	hub             *hub.Hub
	lineAdapter     *adapters.LineAdapter
	metaAdapter     *adapters.MetaAdapter
	douyinAdapter   *adapters.DouyinAdapter
	telegramAdapter *adapters.TelegramAdapter
}

func NewServer(store *storage.Storage, h *hub.Hub) *Server {
	return &Server{
		storage:         store,
		hub:             h,
		lineAdapter:     adapters.NewLineAdapter(),
		metaAdapter:     adapters.NewMetaAdapter(),
		douyinAdapter:   adapters.NewDouyinAdapter(),
		telegramAdapter: adapters.NewTelegramAdapter(),
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux, staticDir string) {
	// Webhooks
	mux.HandleFunc("/api/webhooks/line", s.handleLineWebhook)
	mux.HandleFunc("/api/webhooks/instagram", s.handleMetaWebhook(models.PlatformInstagram))
	mux.HandleFunc("/api/webhooks/facebook", s.handleMetaWebhook(models.PlatformFacebook))
	mux.HandleFunc("/api/webhooks/meta", s.handleMetaWebhook(models.PlatformFacebook))
	mux.HandleFunc("/api/webhooks/douyin", s.handleDouyinWebhook)
	mux.HandleFunc("/api/webhooks/telegram", s.handleTelegramWebhook)

	// REST APIs
	mux.HandleFunc("/api/conversations", s.handleConversations)
	mux.HandleFunc("/api/conversations/", s.handleConversationSubroutes)
	mux.HandleFunc("/api/messages/send", s.handleSendMessage)
	mux.HandleFunc("/api/channels", s.handleChannels)
	mux.HandleFunc("/api/stats", s.handleStats)

	// Webhook Simulation & Testing Sandbox
	mux.HandleFunc("/api/simulator/trigger", s.handleSimulatorTrigger)

	// WebSocket
	mux.HandleFunc("/api/ws", s.hub.HandleWS)

	// Frontend SPA Static Files Serving
	if staticDir != "" {
		fileServer := http.FileServer(http.Dir(staticDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.NotFound(w, r)
				return
			}
			filePath := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
			if fi, err := os.Stat(filePath); err == nil && !fi.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		})
	}
}

func (s *Server) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Line-Signature, X-Hub-Signature-256, X-Telegram-Bot-Api-Secret-Token")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// -------------------------------------------------------------
// Webhook Handlers
// -------------------------------------------------------------

func (s *Server) handleLineWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read body error", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Line-Signature")
	// Look up channel config for secret
	configs, _ := s.storage.GetChannelConfigs()
	var secret string
	for _, c := range configs {
		if c.Platform == models.PlatformLine {
			secret = c.WebhookSecret
			break
		}
	}

	if !s.lineAdapter.VerifySignature(secret, sig, body) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	messages, err := s.lineAdapter.ParseWebhook(body)
	if err != nil {
		log.Printf("[Webhook] LINE parse error: %v", err)
		http.Error(w, "Parse error", http.StatusBadRequest)
		return
	}

	for _, msg := range messages {
		s.processIncomingMessage(msg)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleMetaWebhook(defaultPlatform models.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Meta webhook verification handshake (GET)
		if r.Method == "GET" {
			mode := r.URL.Query().Get("hub.mode")
			token := r.URL.Query().Get("hub.verify_token")
			challenge := r.URL.Query().Get("hub.challenge")

			if mode == "subscribe" && token != "" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(challenge))
				return
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Read body error", http.StatusBadRequest)
			return
		}

		sig := r.Header.Get("X-Hub-Signature-256")
		configs, _ := s.storage.GetChannelConfigs()
		var appSecret string
		for _, c := range configs {
			if c.Platform == defaultPlatform {
				appSecret = c.AppSecret
				break
			}
		}

		if !s.metaAdapter.VerifySignature(appSecret, sig, body) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		messages, err := s.metaAdapter.ParseWebhook(body, defaultPlatform)
		if err != nil {
			log.Printf("[Webhook] Meta parse error: %v", err)
			http.Error(w, "Parse error", http.StatusBadRequest)
			return
		}

		for _, msg := range messages {
			s.processIncomingMessage(msg)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`EVENT_RECEIVED`))
	}
}

func (s *Server) handleDouyinWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read body error", http.StatusBadRequest)
		return
	}

	messages, err := s.douyinAdapter.ParseWebhook(body)
	if err != nil {
		log.Printf("[Webhook] Douyin parse error: %v", err)
		http.Error(w, "Parse error", http.StatusBadRequest)
		return
	}

	for _, msg := range messages {
		s.processIncomingMessage(msg)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"success"}`))
}

func (s *Server) handleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read body error", http.StatusBadRequest)
		return
	}

	messages, err := s.telegramAdapter.ParseWebhook(body)
	if err != nil {
		log.Printf("[Webhook] Telegram parse error: %v", err)
		http.Error(w, "Parse error", http.StatusBadRequest)
		return
	}

	for _, msg := range messages {
		s.processIncomingMessage(msg)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// processIncomingMessage normalizes and stores incoming messages, then broadcasts via WebSocket
func (s *Server) processIncomingMessage(msg *models.Message) {
	// 1. Get or create corresponding conversation
	conv, err := s.storage.GetOrCreateConversation(msg.Platform, msg.SenderID, msg.SenderName, msg.Avatar)
	if err != nil {
		log.Printf("[Server] GetOrCreateConversation err: %v", err)
		return
	}

	msg.ConversationID = conv.ID

	// 2. Persist message in SQLite WAL
	if err := s.storage.SaveMessage(msg); err != nil {
		log.Printf("[Server] SaveMessage err: %v", err)
		return
	}

	// 3. Re-read updated conversation
	updatedConv, _ := s.storage.GetConversation(conv.ID)

	// 4. Broadcast via WebSocket in microsecond latency
	s.hub.Broadcast("new_message", msg)
	if updatedConv != nil {
		s.hub.Broadcast("conversation_updated", updatedConv)
	}
}

// -------------------------------------------------------------
// REST API Handlers
// -------------------------------------------------------------

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	platform := models.Platform(r.URL.Query().Get("platform"))
	list, err := s.storage.GetConversations(platform)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, list)
}

func (s *Server) handleConversationSubroutes(w http.ResponseWriter, r *http.Request) {
	// URL paths: /api/conversations/{id}, /api/conversations/{id}/messages, /api/conversations/{id}/read, /api/conversations/{id}/metadata
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/conversations/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	convID := parts[0]

	if len(parts) == 1 {
		// GET /api/conversations/{id}
		conv, err := s.storage.GetConversation(convID)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		respondJSON(w, http.StatusOK, conv)
		return
	}

	sub := parts[1]
	switch sub {
	case "messages":
		if r.Method == "GET" {
			messages, err := s.storage.GetMessages(convID, 100)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, http.StatusOK, messages)
			return
		}

	case "read":
		if r.Method == "POST" {
			if err := s.storage.MarkConversationAsRead(convID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			updatedConv, _ := s.storage.GetConversation(convID)
			s.hub.Broadcast("conversation_updated", updatedConv)
			respondJSON(w, http.StatusOK, map[string]any{"success": true})
			return
		}

	case "metadata":
		if r.Method == "PUT" {
			var body struct {
				Notes string   `json:"notes"`
				Tags  []string `json:"tags"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			if err := s.storage.UpdateConversationNotesAndTags(convID, body.Notes, body.Tags); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			updatedConv, _ := s.storage.GetConversation(convID)
			s.hub.Broadcast("conversation_updated", updatedConv)
			respondJSON(w, http.StatusOK, updatedConv)
			return
		}
	}

	http.NotFound(w, r)
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.OutgoingMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad payload", http.StatusBadRequest)
		return
	}

	conv, err := s.storage.GetConversation(req.ConversationID)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// 1. Create Outgoing message record
	now := time.Now().UnixMilli()
	msg := &models.Message{
		ID:             uuid.New().String(),
		ConversationID: conv.ID,
		Platform:       conv.Platform,
		SenderID:       "agent_001",
		SenderName:     "客服專員 (您)",
		SenderRole:     models.RoleAgent,
		Avatar:         "https://api.dicebear.com/7.x/bottts/svg?seed=agent",
		Content:        req.Content,
		ContentType:    req.ContentType,
		MediaURL:       req.MediaURL,
		Status:         models.StatusSent,
		Timestamp:      now,
	}

	if msg.ContentType == "" {
		msg.ContentType = models.ContentTypeText
	}

	// 2. Dispatch to external platform API (Goroutine for non-blocking latency)
	go func() {
		configs, _ := s.storage.GetChannelConfigs()
		var channelCfg *models.ChannelConfig
		for _, c := range configs {
			if c.Platform == conv.Platform {
				channelCfg = c
				break
			}
		}

		var sendErr error
		if channelCfg != nil {
			switch conv.Platform {
			case models.PlatformLine:
				sendErr = s.lineAdapter.SendMessage(channelCfg.AccessToken, conv.ExternalUserID, req.Content)
			case models.PlatformInstagram:
				sendErr = s.metaAdapter.SendMessage(channelCfg.AccessToken, conv.ExternalUserID, req.Content)
			case models.PlatformFacebook:
				sendErr = s.metaAdapter.SendMessage(channelCfg.AccessToken, conv.ExternalUserID, req.Content)
			case models.PlatformDouyin:
				sendErr = s.douyinAdapter.SendMessage(channelCfg.AccessToken, conv.ExternalUserID, req.Content)
			case models.PlatformTelegram:
				sendErr = s.telegramAdapter.SendMessage(channelCfg.AccessToken, conv.ExternalUserID, req.Content)
			}
		}

		if sendErr != nil {
			log.Printf("[Outbound Dispatch Error] Platform %s: %v", conv.Platform, sendErr)
		}
	}()

	// 3. Save message and update conversation stats
	if err := s.storage.SaveMessage(msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedConv, _ := s.storage.GetConversation(conv.ID)

	// 4. Real-time broadcast
	s.hub.Broadcast("new_message", msg)
	if updatedConv != nil {
		s.hub.Broadcast("conversation_updated", updatedConv)
	}

	respondJSON(w, http.StatusOK, msg)
}

func (s *Server) handleChannels(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		configs, err := s.storage.GetChannelConfigs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, configs)
		return
	}

	if r.Method == "POST" {
		var cfg models.ChannelConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if err := s.storage.SaveChannelConfig(&cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{"success": true})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	convs, err := s.storage.GetConversations("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalConvs := len(convs)
	totalUnread := 0
	platformCounts := make(map[string]int)

	for _, c := range convs {
		totalUnread += c.UnreadCount
		platformCounts[string(c.Platform)]++
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"totalConversations": totalConvs,
		"totalUnread":        totalUnread,
		"platformCounts":     platformCounts,
		"serverTime":         time.Now().UnixMilli(),
	})
}

// -------------------------------------------------------------
// Webhook Simulator Handler (for instant testing & demos)
// -------------------------------------------------------------

func (s *Server) handleSimulatorTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SimulatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if req.SenderID == "" {
		req.SenderID = fmt.Sprintf("user_%d", time.Now().Unix()%100000)
	}

	if req.SenderName == "" {
		switch req.Platform {
		case models.PlatformLine:
			req.SenderName = "LINE 訪客 " + req.SenderID[len(req.SenderID)-4:]
		case models.PlatformInstagram:
			req.SenderName = "IG_user_" + req.SenderID[len(req.SenderID)-4:]
		case models.PlatformFacebook:
			req.SenderName = "FB 顧客 " + req.SenderID[len(req.SenderID)-4:]
		case models.PlatformDouyin:
			req.SenderName = "抖音粉絲 " + req.SenderID[len(req.SenderID)-4:]
		case models.PlatformTelegram:
			req.SenderName = "TG @" + req.SenderID
		default:
			req.SenderName = "訪客 " + req.SenderID
		}
	}

	if req.AvatarURL == "" {
		req.AvatarURL = fmt.Sprintf("https://api.dicebear.com/7.x/bottts/svg?seed=%s_%s", req.Platform, req.SenderID)
	}

	if req.Content == "" {
		req.Content = fmt.Sprintf("你好！這是一則來自 %s 的即時測試諮詢訊息。", strings.ToUpper(string(req.Platform)))
	}

	cType := models.ContentTypeText
	if req.ContentType != "" {
		cType = models.ContentType(req.ContentType)
	}

	msg := &models.Message{
		ID:          uuid.New().String(),
		Platform:    req.Platform,
		SenderID:    req.SenderID,
		SenderName:  req.SenderName,
		SenderRole:  models.RoleUser,
		Avatar:      req.AvatarURL,
		Content:     req.Content,
		ContentType: cType,
		MediaURL:    req.MediaURL,
		Status:      models.StatusDelivered,
		Timestamp:   time.Now().UnixMilli(),
		RawPayload:  `{"simulated": true}`,
	}

	s.processIncomingMessage(msg)

	respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": msg,
	})
}

func respondJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
