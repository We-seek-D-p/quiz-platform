package ws

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/coder/websocket"
	"github.com/google/uuid"
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
	conn         *websocket.Conn
	log          *slog.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	readLimit    int64
	connectionID string

	bootstrap BootstrapData
	outbound  chan []byte
	closeOnce sync.Once
}

func NewConnection(parent context.Context, conn *websocket.Conn, log *slog.Logger, readLimit int64, bootstrap BootstrapData) *Connection {
	ctx, cancel := context.WithCancel(parent)

	return &Connection{
		conn:         conn,
		log:          log,
		ctx:          ctx,
		cancel:       cancel,
		readLimit:    readLimit,
		connectionID: uuid.NewString(),
		bootstrap:    bootstrap,
		outbound:     make(chan []byte, outboundBufferSize),
	}
}

func (c *Connection) Run() {
	c.conn.SetReadLimit(c.readLimit)

	c.log.InfoContext(c.ctx, "websocket connected", "connection_id", c.connectionID, "role", c.bootstrap.Role, "host_user_id", c.bootstrap.HostUserID)

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
		msgType, payload, err := c.conn.Read(c.ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 || errors.Is(err, context.Canceled) {
				return
			}

			c.log.WarnContext(c.ctx, "websocket read failed", "connection_id", c.connectionID, "role", c.bootstrap.Role, "error", err)
			return
		}

		if msgType != websocket.MessageText {
			c.log.WarnContext(c.ctx, "unsupported websocket message type", "connection_id", c.connectionID, "role", c.bootstrap.Role, "type", msgType)
			return
		}

		envelope, err := DecodeEnvelope(payload)
		if err != nil {
			wsErr := ToWSError(err)
			c.log.WarnContext(c.ctx, "websocket message decode failed", "connection_id", c.connectionID, "role", c.bootstrap.Role, "error_code", wsErr.Code, "error", wsErr.Err)
			c.WriteError(wsErr.Code, wsErr.Message)
			continue
		}

		if err := dispatchIncomingMessage(envelope); err != nil {
			wsErr := ToWSError(err)
			c.log.DebugContext(c.ctx, "websocket message dispatch rejected", "connection_id", c.connectionID, "role", c.bootstrap.Role, "message_type", envelope.Type, "error_code", wsErr.Code)
			c.WriteError(wsErr.Code, wsErr.Message)
			continue
		}
	}
}

func (c *Connection) WriteError(code, message string) {
	payload, err := EncodeEnvelope(ServerEventError, ErrorPayload{Code: code, Message: message})
	if err != nil {
		c.log.WarnContext(c.ctx, "failed to encode websocket error message", "connection_id", c.connectionID, "role", c.bootstrap.Role, "error", err)
		return
	}

	if ok := c.EnqueueText(payload); !ok {
		c.log.WarnContext(c.ctx, "failed to enqueue websocket error message", "connection_id", c.connectionID, "role", c.bootstrap.Role)
	}
}

func (c *Connection) writeLoop() {
	for payload := range c.outbound {
		if err := c.conn.Write(c.ctx, websocket.MessageText, payload); err != nil {
			if websocket.CloseStatus(err) == -1 && !errors.Is(err, context.Canceled) {
				c.log.WarnContext(c.ctx, "websocket write failed", "connection_id", c.connectionID, "role", c.bootstrap.Role, "error", err)
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
		c.log.InfoContext(context.Background(), "websocket disconnected", "connection_id", c.connectionID, "role", c.bootstrap.Role, "host_user_id", c.bootstrap.HostUserID)
	})
}
