package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

type WebChannel struct {
	BaseChannel
	cfg      *config.WebConfig
	server   *http.Server
	upgrader websocket.Upgrader
	clients  map[string]*websocket.Conn
	mu       sync.RWMutex
}

func NewWebChannel(cfg *config.WebConfig, bus *bus.MessageBus) *WebChannel {
	base := NewBaseChannel("web", bus)
	wc := &WebChannel{
		BaseChannel: *base,
		cfg:         cfg,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[string]*websocket.Conn),
	}
	return wc
}

func (c *WebChannel) Start(ctx context.Context) error {
	if !c.cfg.Enabled {
		logger.Info("web channel disabled")
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", c.handleWebSocket)
	mux.HandleFunc("/api/chat", c.handleChat)
	mux.HandleFunc("/api/health", c.handleHealth)

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	c.server = &http.Server{
		Addr:    addr,
		Handler: c.corsMiddleware(mux),
	}

	logger.Info("web channel starting on %s", addr)

	errCh := make(chan error, 1)
	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go c.handleOutbound(ctx)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return c.Stop()
	}
}

func (c *WebChannel) Stop(ctx context.Context) error {
	if c.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return c.server.Shutdown(shutdownCtx)
	}
	return nil
}

func (c *WebChannel) IsRunning() bool {
	return c.server != nil
}

func (c *WebChannel) Send(ctx context.Context, msg *bus.OutboundMessage) error {
	c.mu.RLock()
	conn, ok := c.clients[msg.ChatID]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no websocket connection for chat ID: %s", msg.ChatID)
	}

	response := map[string]interface{}{
		"type":    "message",
		"content": msg.Content,
		"chatId":  msg.ChatID,
	}

	return conn.WriteJSON(response)
}

func (c *WebChannel) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c *WebChannel) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("websocket upgrade failed: %v", err)
		return
	}

	chatID := fmt.Sprintf("web-%d", time.Now().UnixNano())
	c.mu.Lock()
	c.clients[chatID] = conn
	c.mu.Unlock()

	logger.Info("websocket client connected: %s", chatID)

	defer func() {
		c.mu.Lock()
		delete(c.clients, chatID)
		c.mu.Unlock()
		conn.Close()
		logger.Info("websocket client disconnected: %s", chatID)
	}()

	welcome := map[string]interface{}{
		"type":   "welcome",
		"chatId": chatID,
	}
	conn.WriteJSON(welcome)

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("websocket read error: %v", err)
			}
			break
		}

		content, ok := msg["content"].(string)
		if !ok {
			continue
		}

		inbound := &bus.InboundMessage{
			Channel:   c.name,
			ChatID:    chatID,
			UserID:    chatID,
			Content:   content,
			Timestamp: time.Now(),
		}

		if err := c.bus.PublishInbound(context.Background(), inbound); err != nil {
			logger.Error("failed to publish inbound message: %v", err)
		}
	}
}

func (c *WebChannel) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
		ChatID  string `json:"chatId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.ChatID == "" {
		req.ChatID = fmt.Sprintf("web-%d", time.Now().UnixNano())
	}

	inbound := &bus.InboundMessage{
		Channel:   c.name,
		ChatID:    req.ChatID,
		UserID:    req.ChatID,
		Content:   req.Content,
		Timestamp: time.Now(),
	}

	if err := c.bus.PublishInbound(context.Background(), inbound); err != nil {
		http.Error(w, "Failed to process message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"chatId": req.ChatID,
	})
}

func (c *WebChannel) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"channel": "web",
	})
}

func (c *WebChannel) handleOutbound(ctx context.Context) {
	sub := c.bus.SubscribeOutbound(c.name)
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-sub.Messages():
			if err := c.Send(ctx, msg); err != nil {
				logger.Error("failed to send outbound message: %v", err)
			}
		}
	}
}
