package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"global-comm-system/internal/api"
	"global-comm-system/internal/hub"
	"global-comm-system/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	execPath, err := os.Executable()
	baseDir := "."
	if err == nil {
		baseDir = filepath.Dir(execPath)
	}

	dbPath := filepath.Join(baseDir, "data", "global_comm.db")
	if envDB := os.Getenv("DB_PATH"); envDB != "" {
		dbPath = envDB
	}

	log.Printf("=========================================================")
	log.Printf("  Global Communication System (全渠道通訊整合系統)")
	log.Printf("  整合 LINE | Instagram | Facebook | 抖音 | Telegram")
	log.Printf("  高效能後端 (Go Native + SQLite WAL + WebSocket)")
	log.Printf("=========================================================")

	// 1. Initialize SQLite Database
	store, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("[Fatal] Failed to initialize storage: %v", err)
	}
	defer store.Close()
	log.Printf("[Database] SQLite WAL mode loaded at: %s", dbPath)

	// 2. Initialize WebSocket Hub
	wsHub := hub.NewHub()
	go wsHub.Run()
	log.Printf("[WebSocket] Real-time Hub started")

	// 3. Setup HTTP Server & API Handlers
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		candidates := []string{
			filepath.Join(baseDir, "..", "frontend", "dist"),
			filepath.Join(baseDir, "frontend", "dist"),
			"../frontend/dist",
			"frontend/dist",
			"/app/frontend/dist",
			filepath.Join(baseDir, "dist"),
		}
		for _, cand := range candidates {
			if fi, err := os.Stat(cand); err == nil && fi.IsDir() {
				staticDir = cand
				break
			}
		}
	}

	server := api.NewServer(store, wsHub)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux, staticDir)
	if staticDir != "" {
		log.Printf("[Static] Serving frontend SPA from: %s", staticDir)
	}

	// Enable CORS
	handler := server.CORS(mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 4. Start Server
	go func() {
		log.Printf("[Server] Listening on http://localhost:%s", port)
		log.Printf("[Server] Webhooks ready:")
		log.Printf("  - LINE:      POST http://localhost:%s/api/webhooks/line", port)
		log.Printf("  - Instagram: POST http://localhost:%s/api/webhooks/instagram", port)
		log.Printf("  - Facebook:  POST http://localhost:%s/api/webhooks/facebook", port)
		log.Printf("  - 抖音:      POST http://localhost:%s/api/webhooks/douyin", port)
		log.Printf("  - Telegram:  POST http://localhost:%s/api/webhooks/telegram", port)
		log.Printf("  - WebSocket: ws://localhost:%s/api/ws", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Fatal] Server failed: %v", err)
		}
	}()

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[Server] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[Fatal] Server forced shutdown: %v", err)
	}
	log.Println("[Server] Bye!")
	fmt.Println()
}
