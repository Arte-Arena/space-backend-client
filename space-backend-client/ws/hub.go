// space-backend-client/ws/hub.go
package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Hub gerencia conexões e broadcast de mensagens
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// NewHub inicializa um Hub vazio
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run roda o loop principal do hub
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true

		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}

		case msg := <-h.broadcast:
			for conn := range h.clients {
				go func(c *websocket.Conn) {
					c.SetWriteDeadline(time.Now().Add(10 * time.Second))
					if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
						log.Printf("ws write error: %v", err)
						h.unregister <- c
					}
				}(conn)
			}
		}
	}
}

// Upgrader para WebSocket (ajuste CheckOrigin conforme desejar)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Handler retorna um http.HandlerFunc que faz upgrade e registra no Hub
func (h *Hub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade failed: %v", err)
			return
		}
		h.register <- conn

		// loop de leitura para detectar disconnects
		go func() {
			defer func() { h.unregister <- conn }()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	}
}

// Broadcast insere uma mensagem no canal de saída
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}
