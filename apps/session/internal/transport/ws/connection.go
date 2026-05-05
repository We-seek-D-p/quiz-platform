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
	conn          *websocket.Conn
	log           *slog.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	readLimit     int64
	connectionID  string
	metaMu        sync.RWMutex
	sessionID     string
	participantID string
	onClose       func(*Connection)
	onMessage     func(context.Context, *Connection, MessageEnvelope) error

	bootstrap BootstrapData
	outbound  chan []byte
	closeOnce sync.Once
}

func NewConnection(parent context.Context, conn *websocket.Conn, log *slog.Logger, readLimit int64, bootstrap BootstrapData) *Connection {
	return &Connection{
		conn:         conn,
		log:          log,
		ctx:          parent,
		cancel:       nil,
		readLimit:    readLimit,
		connectionID: uuid.NewString(),
		bootstrap:    bootstrap,
		outbound:     make(chan []byte, outboundBufferSize),
	}
}

func (c *Connection) Run() {
	runCtx, cancel := context.WithCancel(c.ctx)
	c.metaMu.Lock()
	c.ctx = runCtx
	c.cancel = cancel
	c.metaMu.Unlock()

	c.conn.SetReadLimit(c.readLimit)

	c.log.InfoContext(c.ctx, "websocket connected", "connection_id", c.connectionID, "role", c.bootstrap.Role, "host_user_id", c.bootstrap.HostUserID)

	go c.writeLoop()
	c.readLoop()
	c.close(websocket.StatusNormalClosure, "connection closed")
}

func (c *Connection) ID() string {
	return c.connectionID
}

func (c *Connection) Role() ConnectionRole {
	return c.bootstrap.Role
}

func (c *Connection) SessionID() string {
	c.metaMu.RLock()
	defer c.metaMu.RUnlock()

	return c.sessionID
}

func (c *Connection) ParticipantID() string {
	c.metaMu.RLock()
	defer c.metaMu.RUnlock()

	return c.participantID
}

func (c *Connection) HostUserID() string {
	return c.bootstrap.HostUserID
}

func (c *Connection) BindSession(sessionID string) {
	c.metaMu.Lock()
	c.sessionID = sessionID
	c.metaMu.Unlock()
}

func (c *Connection) BindParticipant(participantID string) {
	c.metaMu.Lock()
	c.participantID = participantID
	c.metaMu.Unlock()
}

func (c *Connection) SetOnClose(callback func(*Connection)) {
	c.metaMu.Lock()
	c.onClose = callback
	c.metaMu.Unlock()
}

func (c *Connection) SetMessageHandler(handler func(context.Context, *Connection, MessageEnvelope) error) {
	c.metaMu.Lock()
	c.onMessage = handler
	c.metaMu.Unlock()
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

		c.metaMu.RLock()
		onMessage := c.onMessage
		c.metaMu.RUnlock()

		if onMessage == nil {
			c.WriteError("internal_error", "internal error")
			continue
		}

		if err := onMessage(c.ctx, c, envelope); err != nil {
			wsErr := ToWSError(err)
			c.log.DebugContext(c.ctx, "websocket message dispatch rejected", "connection_id", c.connectionID, "role", c.bootstrap.Role, "message_type", envelope.Type, "error_code", wsErr.Code)
			c.WriteError(wsErr.Code, wsErr.Message)
			continue
		}
	}
}

func (c *Connection) WriteEvent(messageType string, payload any) error {
	encoded, err := EncodeEnvelope(messageType, payload)
	if err != nil {
		return err
	}

	if ok := c.EnqueueText(encoded); !ok {
		return NewWSError("internal_error", "connection is closed")
	}

	return nil
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
		c.metaMu.RLock()
		onClose := c.onClose
		c.metaMu.RUnlock()

		if c.cancel != nil {
			c.cancel()
		}
		close(c.outbound)
		_ = c.conn.Close(code, reason)
		if onClose != nil {
			onClose(c)
		}
		c.log.InfoContext(context.Background(), "websocket disconnected", "connection_id", c.connectionID, "role", c.bootstrap.Role, "host_user_id", c.bootstrap.HostUserID)
	})
}
