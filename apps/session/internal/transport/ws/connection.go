package ws

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/coder/websocket"
)

const outboundBufferSize = 16

type ConnectionRole string

const (
	ConnectionRoleHost   ConnectionRole = "host"
	ConnectionRolePlayer ConnectionRole = "player"
)

type BootstrapData struct {
	Role         ConnectionRole
	HostUserID   string
	HostUserRole string
}

type Connection struct {
	conn      *websocket.Conn
	log       *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	readLimit int64

	bootstrap BootstrapData
	outbound  chan []byte
	closeOnce sync.Once
}

func NewConnection(parent context.Context, conn *websocket.Conn, log *slog.Logger, readLimit int64, bootstrap BootstrapData) *Connection {
	ctx, cancel := context.WithCancel(parent)

	return &Connection{
		conn:      conn,
		log:       log,
		ctx:       ctx,
		cancel:    cancel,
		readLimit: readLimit,
		bootstrap: bootstrap,
		outbound:  make(chan []byte, outboundBufferSize),
	}
}

func (c *Connection) Run() {
	c.conn.SetReadLimit(c.readLimit)

	c.log.InfoContext(c.ctx, "websocket connected", "role", c.bootstrap.Role, "host_user_id", c.bootstrap.HostUserID)

	go c.writeLoop()
	c.readLoop()
	c.close(websocket.StatusNormalClosure, "connection closed")
}

func (c *Connection) EnqueueText(payload []byte) bool {
	select {
	case <-c.ctx.Done():
		return false
	default:
	}

	select {
	case c.outbound <- payload:
		return true
	case <-c.ctx.Done():
		return false
	}
}

func (c *Connection) readLoop() {
	for {
		msgType, _, err := c.conn.Read(c.ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 || errors.Is(err, context.Canceled) {
				return
			}

			c.log.WarnContext(c.ctx, "websocket read failed", "role", c.bootstrap.Role, "error", err)
			return
		}

		if msgType != websocket.MessageText {
			c.log.WarnContext(c.ctx, "unsupported websocket message type", "role", c.bootstrap.Role, "type", msgType)
			return
		}
	}
}

func (c *Connection) writeLoop() {
	for payload := range c.outbound {
		if err := c.conn.Write(c.ctx, websocket.MessageText, payload); err != nil {
			if websocket.CloseStatus(err) == -1 && !errors.Is(err, context.Canceled) {
				c.log.WarnContext(c.ctx, "websocket write failed", "role", c.bootstrap.Role, "error", err)
			}
			c.close(websocket.StatusInternalError, "write failed")
			return
		}
	}
}

func (c *Connection) close(code websocket.StatusCode, reason string) {
	c.closeOnce.Do(func() {
		c.cancel()
		close(c.outbound)
		_ = c.conn.Close(code, reason)
		c.log.InfoContext(context.Background(), "websocket disconnected", "role", c.bootstrap.Role, "host_user_id", c.bootstrap.HostUserID)
	})
}
