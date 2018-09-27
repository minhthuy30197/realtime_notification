package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// user id
	userID string 
}

type Message struct {
	UserID       string `json:"user_id"`
	MessageType string `json:"message_type"`
}

func (c *Client) readPump() {
	defer func() {	
		log.Println(c.userID, " đóng kết nối")
		c.conn.Close()
	}()
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		log.Println(time.Now())
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			// Xoá connection 
			delete(c.hub.clients[c.userID].list, c)
			if len(c.hub.clients[c.userID].list) == 0 {
				delete(c.hub.clients, c.userID)	 
			}
			break
		} 
		n := len(message)
		data := fmt.Sprintf("%s", (string(message[:n])))
		var msg Message
		err = json.Unmarshal([]byte(data), &msg)
		if err != nil {
			fmt.Println("error:", err)
			break
		}
		log.Println(msg)
		if msg.MessageType == "register" {
			var userID string 
			// Parse token ra userID
			userID = msg.UserID
			if _, ok := c.hub.clients[userID]; ok {
				c.userID = userID
				log.Println(userID, " mở một tab khác")
				c.hub.clients[userID].list[c] = true
			} else {
				log.Println(userID, " đăng kí")
				c.userID = userID
				tmp := &ListClient{
					list: make(map[*Client]bool),
				}
				tmp.list[c] = true 
				c.hub.clients[userID] = tmp
			}
		} else {
			
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// Kiểm tra c.conn đã tồn tại trong clients chưa
			if c.userID != "" {
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			} else {
				log.Println("Quá giờ chưa gửi token để lấy userID")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}			
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
