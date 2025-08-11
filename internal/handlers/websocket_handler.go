package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// --- Stats Tracker ---
type StatsTracker struct {
	mu            sync.Mutex
	HumanizeCount int
	DetectCount   int
	PlagiarizeCount int
	ResearchCount int
	hub           *Hub
}

func NewStatsTracker(h *Hub) *StatsTracker {
	return &StatsTracker{hub: h}
}

func (st *StatsTracker) Increment(action string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	switch action {
	case "humanize":
		st.HumanizeCount++
	case "detect":
		st.DetectCount++
	case "plagiarize":
		st.PlagiarizeCount++
	case "research":
		st.ResearchCount++
	}
}

func (st *StatsTracker) GetStats() map[string]interface{} {
	st.mu.Lock()
	defer st.mu.Unlock()
	return map[string]interface{}{
		"type":            "stats",
		"humanize_count":  st.HumanizeCount,
		"detect_count":    st.DetectCount,
		"plagiarize_count": st.PlagiarizeCount,
		"research_count":  st.ResearchCount,
	}
}

// --- WebSocket Hub ---
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run(st *StatsTracker) {
	// Ticker to broadcast stats every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			// Send initial stats immediately on connection
			h.broadcastMessage(st.GetStats())
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
		case <-ticker.C:
			h.broadcastMessage(st.GetStats())
		}
	}
}

func (h *Hub) broadcastMessage(message map[string]interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling broadcast message:", err)
		return
	}

	for client := range h.clients {
		if err := client.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			log.Println("WebSocket write error:", err)
			client.Close()
			delete(h.clients, client)
		}
	}
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	h.register <- conn
}