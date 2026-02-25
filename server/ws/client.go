package ws

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second    // max time to write a message
	pongWait       = 60 * time.Second    // max time between pongs from client
	pingPeriod     = (pongWait * 9) / 10 // how often to send pings to client (must be less than pongWait)
	maxMessageSize = 1024                // increased since subscriptions can be large, but still prevent DoS with huge messages
)

// SubMessage is what the client sends to subscribe or unsubscribe.
// {"action": "subscribe", "topic": "bottle:42"}
// {"action": "unsubscribe", "topic": "bottle:42"}
type SubMessage struct {
	Action string `json:"action"` // "subscribe" or "unsubscribe"
	Topic  string `json:"topic"`
}

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
	c.hub.Register(c)
	go c.writePump()
	c.readPump()
}

// readPump keeps the connection alive by draining inbound frames.
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var sub SubMessage
		if err := json.Unmarshal(msg, &sub); err != nil {
			c.log.Warn("invalid ws message", zap.Error(err))
			continue
		}

		switch sub.Action {
		case "subscribe":
			if err := validateTopic(sub.Topic); err != nil {
				c.log.Warn("invalid topic", zap.String("topic", sub.Topic))
				continue
			}
			c.hub.Subscribe(c, sub.Topic)
			c.log.Debug("client subscribe", zap.String("topic", sub.Topic))
		case "unsubscribe":
			c.hub.Unsubscribe(c, sub.Topic)
			c.log.Debug("client unsubscribed", zap.String("topic", sub.Topic))
		default:
			c.log.Warn("unknown ws action", zap.String("action", sub.Action))
		}
	}
}

// writePump drains the send channel and keeps the connection alive with pings.
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
				return
			}
		}
	}
}

// this function ensures topic string ffollow known formats.
// prevents clients from subscribing to arbitrary topics.
func validateTopic(topic string) error {
	if len(topic) == 0 || len(topic) > 64 {
		return fmt.Errorf("topic length invalid")
	}
	if strings.HasPrefix(topic, "bottle:") {
		idStr := strings.TrimPrefix(topic, "bottle:")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil || id <= 0 {
			return fmt.Errorf("invalid bottle topic")
		}
		return nil
	}

	switch topic {
	case "region:north_atlantic",
		"region:south_atlantic",
		"region:north_pacific",
		"region:south_pacific",
		"region:indian_ocean",
		"region:other":
		return nil
	default:
		return fmt.Errorf("unsupported topic")
	}
}
