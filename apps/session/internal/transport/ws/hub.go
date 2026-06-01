package ws

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

var ErrInvalidBinding = errors.New("invalid hub binding")

type Hub struct {
	log      *slog.Logger
	mu       sync.RWMutex
	sessions map[string]*sessionPeers
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{log: log, sessions: make(map[string]*sessionPeers)}
}

// BindHost attaches a host connection to a runtime session.
func (h *Hub) BindHost(sessionID string, conn *Connection) error {
	if sessionID == "" || conn == nil {
		return fmt.Errorf("bind host: %w", ErrInvalidBinding)
	}

	conn.BindSession(sessionID)
	conn.BindParticipant("")

	h.mu.Lock()
	peers := h.ensureSessionPeersLocked(sessionID)
	peers.host = new(newPeerRef(conn, sessionID, "", ConnectionRoleHost))
	h.mu.Unlock()

	h.log.Debug("hub host bound", "connection_id", conn.ID(), "session_id", sessionID, "role", conn.Role())
	return nil
}

// BindPlayer attaches a player connection to a runtime session.
func (h *Hub) BindPlayer(sessionID, participantID string, conn *Connection) error {
	if sessionID == "" || participantID == "" || conn == nil {
		return fmt.Errorf("bind player: %w", ErrInvalidBinding)
	}

	conn.BindSession(sessionID)
	conn.BindParticipant(participantID)

	h.mu.Lock()
	peers := h.ensureSessionPeersLocked(sessionID)
	peers.players[participantID] = new(newPeerRef(conn, sessionID, participantID, ConnectionRolePlayer))
	h.mu.Unlock()

	h.log.Debug("hub player bound", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "role", conn.Role())
	return nil
}

// Unbind removes a connection from its current session registry.
func (h *Hub) Unbind(conn *Connection) bool {
	if conn == nil {
		return false
	}

	ref := peerRef{
		connectionID:  conn.ID(),
		sessionID:     conn.SessionID(),
		participantID: conn.ParticipantID(),
		role:          conn.Role(),
		connection:    conn,
	}
	if ref.sessionID == "" {
		return false
	}

	h.mu.Lock()
	removed := h.unbindByRefLocked(ref)
	h.mu.Unlock()

	if removed {
		h.log.Debug("hub connection unbound", "connection_id", ref.connectionID, "session_id", ref.sessionID, "participant_id", ref.participantID, "role", ref.role)
	}

	return removed
}

// Broadcast sends a payload to every connection in a session.
func (h *Hub) Broadcast(sessionID string, payload []byte) int {
	recipients := h.snapshotAll(sessionID)
	stale := h.sendToRecipients(recipients, payload)
	h.dropStale(stale)
	return len(recipients) - len(stale)
}

// SendHost sends a payload to the session host connection.
func (h *Hub) SendHost(sessionID string, payload []byte) bool {
	recipient := h.snapshotHost(sessionID)
	if recipient.connection == nil {
		return false
	}

	if ok := recipient.connection.EnqueueText(payload); !ok {
		h.dropStale([]peerRef{recipient})
		return false
	}

	return true
}

// SendPlayer sends a payload to one player connection.
func (h *Hub) SendPlayer(sessionID, participantID string, payload []byte) bool {
	recipient := h.snapshotPlayer(sessionID, participantID)
	if recipient.connection == nil {
		return false
	}

	if ok := recipient.connection.EnqueueText(payload); !ok {
		h.dropStale([]peerRef{recipient})
		return false
	}

	return true
}

// BroadcastPlayers sends a payload to all player connections in a session.
func (h *Hub) BroadcastPlayers(sessionID string, payload []byte) int {
	recipients := h.snapshotPlayers(sessionID)
	stale := h.sendToRecipients(recipients, payload)
	h.dropStale(stale)
	return len(recipients) - len(stale)
}

// ActiveSessionIDs returns sessions that currently have active WebSocket peers.
func (h *Hub) ActiveSessionIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.sessions))
	for sessionID := range h.sessions {
		ids = append(ids, sessionID)
	}

	return ids
}

func (h *Hub) ensureSessionPeersLocked(sessionID string) *sessionPeers {
	peers, ok := h.sessions[sessionID]
	if ok {
		return peers
	}

	peers = newSessionPeers()
	h.sessions[sessionID] = peers
	return peers
}

func (h *Hub) snapshotHost(sessionID string) peerRef {
	h.mu.RLock()
	defer h.mu.RUnlock()

	peers, ok := h.sessions[sessionID]
	if !ok || peers.host == nil {
		return peerRef{}
	}

	return *peers.host
}

func (h *Hub) snapshotPlayer(sessionID, participantID string) peerRef {
	h.mu.RLock()
	defer h.mu.RUnlock()

	peers, ok := h.sessions[sessionID]
	if !ok {
		return peerRef{}
	}

	conn, ok := peers.players[participantID]
	if !ok {
		return peerRef{}
	}

	return *conn
}

func (h *Hub) snapshotPlayers(sessionID string) []peerRef {
	h.mu.RLock()
	defer h.mu.RUnlock()

	peers, ok := h.sessions[sessionID]
	if !ok {
		return nil
	}

	res := make([]peerRef, 0, len(peers.players))
	for _, conn := range peers.players {
		res = append(res, *conn)
	}

	return res
}

func (h *Hub) snapshotAll(sessionID string) []peerRef {
	h.mu.RLock()
	defer h.mu.RUnlock()

	peers, ok := h.sessions[sessionID]
	if !ok {
		return nil
	}

	capacity := len(peers.players)
	if peers.host != nil {
		capacity++
	}

	res := make([]peerRef, 0, capacity)
	if peers.host != nil {
		res = append(res, *peers.host)
	}

	for _, conn := range peers.players {
		res = append(res, *conn)
	}

	return res
}

// sendToRecipients enqueues a payload and returns stale recipients for cleanup.
func (h *Hub) sendToRecipients(recipients []peerRef, payload []byte) []peerRef {
	stale := make([]peerRef, 0)
	for _, recipient := range recipients {
		if recipient.connection == nil {
			continue
		}

		if ok := recipient.connection.EnqueueText(payload); !ok {
			stale = append(stale, recipient)
		}
	}

	return stale
}

// dropStale removes recipients that rejected outbound messages.
func (h *Hub) dropStale(stale []peerRef) {
	if len(stale) == 0 {
		return
	}

	h.mu.Lock()
	for _, ref := range stale {
		if h.unbindByRefLocked(ref) {
			h.log.Debug("hub dropped stale connection", "connection_id", ref.connectionID, "session_id", ref.sessionID, "participant_id", ref.participantID, "role", ref.role)
		}
	}
	h.mu.Unlock()
}

func (h *Hub) unbindByRefLocked(ref peerRef) bool {
	peers, ok := h.sessions[ref.sessionID]
	if !ok {
		return false
	}

	removed := false
	if ref.role == ConnectionRoleHost {
		if peers.host != nil && peers.host.connection == ref.connection {
			peers.host = nil
			removed = true
		}
	} else {
		current, exists := peers.players[ref.participantID]
		if exists && current.connection == ref.connection {
			delete(peers.players, ref.participantID)
			removed = true
		}
	}

	if peers.isEmpty() {
		delete(h.sessions, ref.sessionID)
	}

	return removed
}
