package ws

import (
	"time"

	"github.com/gofiber/contrib/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second    // max time to write a message
	pongWait       = 60 * time.Second    // max time between pongs from client
	pingPeriod     = (pongWait * 9) / 10 // how often to send pings to client (must be less than pongWait)
	maxMessageSize = 512                 // maximum message size allowed from client
)

// Client is a single WebSocket connection.
// It bridges the Hub and the underlying WebSocket connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte // outbound messages to the client
	log  *zap.Logger
}

func NewClient(hub *Hub, conn *websocket.Conn, log *zap.Logger) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256), // buffered channel for outbound messages
		log:  log,
	}
}

// Run starts both pumps. Call this after creating the client.
// Blocks until the client connection is closed.
func (c *Client) Run() {
	go c.writePump()
	c.readPump()
}

// Read pump keeps the connection alive and drains any inbound message
// The ocean viewers doesn't send data to the server, so we just need
// to handle pings and close the connection when the client disconnects.
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		// Every time the client sends a pong, reset the read deadline to keep the connection alive.
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		// We don't expect to receive any messages from the client, but we need to read to detect disconnections and handle pings.
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break // connection closed or error occurred, exit the read pump to clean up
		}
	}
}

// Write pump sends messages to the client and handles periodic pings.
// It runs in a separate goroutine and blocks until the client connection is closed.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel, so we should close the connection.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.log.Warn("ws write failed", zap.Error(err))
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // client didn't respond to ping, close the connection
			}
		}
	}
}
