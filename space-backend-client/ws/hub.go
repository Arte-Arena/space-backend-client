package ws

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub gerencia conexões e broadcast de mensagens
type Hub struct {
	clients    map[*websocket.Conn]bool
	clientsMux sync.Mutex // Mutex para proteger o mapa de clientes
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
			h.clientsMux.Lock()
			h.clients[conn] = true
			h.clientsMux.Unlock()

		case conn := <-h.unregister:
			h.clientsMux.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.clientsMux.Unlock()

		case msg := <-h.broadcast:
			h.clientsMux.Lock()
			clientsCopy := make([]*websocket.Conn, 0, len(h.clients))
			for conn := range h.clients {
				clientsCopy = append(clientsCopy, conn)
			}
			h.clientsMux.Unlock()

			for _, conn := range clientsCopy {
				go func(c *websocket.Conn) {
					log.Printf("[Hub] enviando para %s: %s", c.RemoteAddr(), string(msg))

          defer func() {
						if r := recover(); r != nil {
							log.Printf("Recovered from panic in websocket write: %v", r)
							h.unregister <- c
						}
					}()

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
	// log do hub exibindo tamanho e conteúdo
	h.clientsMux.Lock()
	n := len(h.clients)
	h.clientsMux.Unlock()
	log.Printf("[Hub] broadcast para %d clientes; payload: %s", n, string(message))

	h.broadcast <- message
}
