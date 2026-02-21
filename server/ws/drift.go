package ws

import (
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

// upgrader performs the HTTP→WebSocket handshake via fasthttp.
// CORS is already handled by the global middleware, so CheckOrigin always passes.
var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(_ *fasthttp.RequestCtx) bool { return true },
}

// NewDriftHandler returns a fiber.Handler that upgrades the connection to
// WebSocket and registers the new client with the Hub so it receives real-time
// bottle drift events.
func NewDriftHandler(hub *Hub, log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		return upgrader.Upgrade(c.RequestCtx(), func(conn *websocket.Conn) {
			client := NewClient(hub, conn, log)
			client.Run()
		})
	}
}
