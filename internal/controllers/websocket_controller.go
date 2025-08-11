package controllers

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

type Client struct {
	Conn     *websocket.Conn
	UserID   int64
	Username string
	Send     chan []byte
}

type Message struct {
	SenderID   int64     `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	ReceiverID int64     `json:"receiver_id"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
}

type WebSocketController struct {
	clients    map[int64]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewWebSocketController() *WebSocketController {
	return &WebSocketController{
		clients:    make(map[int64]*Client),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (wsc *WebSocketController) HandleConnections(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Atualizar para conexão WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading to websocket: %v", err)
		return
	}

	client := &Client{
		Conn:     conn,
		UserID:   userID.(int64),
		Username: username.(string),
		Send:     make(chan []byte, 256),
	}

	wsc.register <- client

	// Goroutine para ler mensagens do cliente
	go wsc.readMessages(client)
	// Goroutine para escrever mensagens para o cliente
	go wsc.writeMessages(client)
}

func (wsc *WebSocketController) readMessages(client *Client) {
	defer func() {
		wsc.unregister <- client
		client.Conn.Close()
	}()

	for {
		var msg Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		msg.SenderID = client.UserID
		msg.SenderName = client.Username
		msg.Timestamp = time.Now()

		wsc.broadcast <- msg
	}
}

func (wsc *WebSocketController) writeMessages(client *Client) {
	defer func() {
		client.Conn.Close()
	}()

	for message := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error writing message: %v", err)
			return
		}
	}
}

func (wsc *WebSocketController) Run() {
	for {
		select {
		case client := <-wsc.register:
			wsc.mu.Lock()
			wsc.clients[client.UserID] = client
			wsc.mu.Unlock()

			// Notificar outros usuários sobre a nova conexão
			wsc.notifyUserStatus(client.UserID, client.Username, true)

		case client := <-wsc.unregister:
			wsc.mu.Lock()
			if _, ok := wsc.clients[client.UserID]; ok {
				delete(wsc.clients, client.UserID)
				close(client.Send)
			}
			wsc.mu.Unlock()

			// Notificar outros usuários sobre a desconexão
			wsc.notifyUserStatus(client.UserID, client.Username, false)

		case message := <-wsc.broadcast:
			// Enviar mensagem para o destinatário específico
			wsc.mu.Lock()
			if receiver, ok := wsc.clients[message.ReceiverID]; ok {
				receiver.Send <- []byte(message.Content)
			}
			wsc.mu.Unlock()
		}
	}
}

func (wsc *WebSocketController) notifyUserStatus(userID int64, username string, online bool) {
	statusMessage := struct {
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
		Online   bool   `json:"online"`
	}{
		UserID:   userID,
		Username: username,
		Online:   online,
	}

	for id, client := range wsc.clients {
		if id != userID { 
			client.Send <- []byte(fmt.Sprintf(`{"type":"status","data":%s}`, statusMessage))
		}
	}
}