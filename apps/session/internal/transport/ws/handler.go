package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/coder/websocket"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	sessionservice "github.com/We-seek-D-p/quiz-platform/apps/session/internal/service/session"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/middleware"
)

type Handler struct {
	log       *slog.Logger
	readLimit int64
	hub       *Hub
	service   *sessionservice.Service
}

func NewHandler(cfg *config.Config, log *slog.Logger, service *sessionservice.Service) *Handler {
	return &Handler{
		log:       log,
		readLimit: int64(cfg.WS.ReadLimitBytes),
		hub:       NewHub(log),
		service:   service,
	}
}

func (h *Handler) Hub() *Hub {
	return h.hub
}

func (h *Handler) Host(w http.ResponseWriter, r *http.Request) {
	h.acceptAndServe(w, r, ConnectionRoleHost)
}

func (h *Handler) Player(w http.ResponseWriter, r *http.Request) {
	h.acceptAndServe(w, r, ConnectionRolePlayer)
}

func (h *Handler) acceptAndServe(w http.ResponseWriter, r *http.Request, role ConnectionRole) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		h.log.WarnContext(r.Context(), "websocket accept failed", "role", role, "error", err)
		return
	}

	bootstrap := BootstrapData{Role: role}
	if role == ConnectionRoleHost {
		bootstrap.HostUserID = r.Header.Get(middleware.UserIDHeader)
		bootstrap.HostUserRole = r.Header.Get(middleware.UserRoleHeader)
	}

	wsConn := NewConnection(r.Context(), conn, h.log, h.readLimit, bootstrap)
	wsConn.SetOnClose(func(conn *Connection) {
		h.hub.Unbind(conn)
	})
	wsConn.SetMessageHandler(h.dispatchIncomingMessage)
	wsConn.Run()
}

type hostConnectPayload struct {
	SessionID string `json:"session_id"`
}

type playerJoinPayload struct {
	RoomCode string `json:"room_code"`
	Nickname string `json:"nickname"`
}

func (h *Handler) dispatchIncomingMessage(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	switch envelope.Type {
	case "host_connect":
		return h.handleHostConnect(ctx, conn, envelope)
	case "player_join":
		return h.handlePlayerJoin(ctx, conn, envelope)
	default:
		return NewWSError(ErrCodeUnknownMessageType, "unknown message type")
	}
}

func (h *Handler) handleHostConnect(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload hostConnectPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	payload.SessionID = strings.TrimSpace(payload.SessionID)
	if payload.SessionID == "" {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	h.log.DebugContext(ctx, "host_connect received", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID)

	result, err := h.service.HostConnect(ctx, sessionservice.HostConnectParams{
		SessionID:  payload.SessionID,
		HostUserID: conn.HostUserID(),
	})
	if err != nil {
		wsErr := ToWSError(mapServiceHostConnectError(err))
		h.log.WarnContext(ctx, "host_connect failed", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID, "error_code", wsErr.Code)
		return wsErr
	}

	if err := h.hub.BindHost(payload.SessionID, conn); err != nil {
		h.log.WarnContext(ctx, "host_connect bind failed", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	if err := conn.WriteEvent("session_snapshot", result.SessionSnapshot); err != nil {
		h.log.WarnContext(ctx, "host_connect snapshot send failed", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	h.log.DebugContext(ctx, "host_connect success", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID)
	return nil
}

func mapServiceHostConnectError(err error) error {
	switch {
	case errors.Is(err, sessionservice.ErrForbidden):
		return NewWSError("forbidden", "forbidden")
	case errors.Is(err, sessionservice.ErrSessionRuntimeNotFound), errors.Is(err, sessionservice.ErrSessionNotFound):
		return NewWSError("session_not_found", "session not found")
	case errors.Is(err, sessionservice.ErrRuntimeStoreUnavailable):
		return NewWSError("internal_error", "internal error")
	case errors.Is(err, sessionservice.ErrInvalidParams):
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	default:
		return NewWSError("internal_error", "internal error")
	}
}

func (h *Handler) handlePlayerJoin(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload playerJoinPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	payload.RoomCode = strings.TrimSpace(payload.RoomCode)
	payload.Nickname = strings.TrimSpace(payload.Nickname)
	if payload.RoomCode == "" || payload.Nickname == "" {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	h.log.DebugContext(ctx, "player_join received", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname)

	result, err := h.service.PlayerJoin(ctx, sessionservice.PlayerJoinParams{
		RoomCode: payload.RoomCode,
		Nickname: payload.Nickname,
	})
	if err != nil {
		wsErr := ToWSError(mapServicePlayerJoinError(err))
		h.log.WarnContext(ctx, "player_join failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "error_code", wsErr.Code)
		return wsErr
	}

	sessionID := result.SessionSnapshot.SessionID
	participantID := result.JoinedLobby.ParticipantID
	if err := h.hub.BindPlayer(sessionID, participantID, conn); err != nil {
		h.log.WarnContext(ctx, "player_join bind failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	if err := conn.WriteEvent("joined_lobby", result.JoinedLobby); err != nil {
		h.log.WarnContext(ctx, "player_join joined_lobby send failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	lobbyUpdatedPayload, err := EncodeEnvelope("lobby_updated", result.LobbyUpdated)
	if err != nil {
		h.log.WarnContext(ctx, "player_join lobby_updated encode failed", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}
	_ = h.hub.Broadcast(sessionID, lobbyUpdatedPayload)

	if err := conn.WriteEvent("session_snapshot", result.SessionSnapshot); err != nil {
		h.log.WarnContext(ctx, "player_join snapshot send failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	h.log.DebugContext(ctx, "player_join success", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "session_id", sessionID, "participant_id", participantID)
	return nil
}

func mapServicePlayerJoinError(err error) error {
	switch {
	case errors.Is(err, sessionservice.ErrRoomNotFound):
		return NewWSError("room_not_found", "room not found")
	case errors.Is(err, sessionservice.ErrNicknameTaken):
		return NewWSError("nickname_taken", "nickname already taken")
	case errors.Is(err, sessionservice.ErrGameAlreadyFinished):
		return NewWSError("game_already_finished", "game already finished")
	case errors.Is(err, sessionservice.ErrInvalidParams):
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	case errors.Is(err, sessionservice.ErrRuntimeStoreUnavailable):
		return NewWSError("internal_error", "internal error")
	default:
		return NewWSError("internal_error", "internal error")
	}
}
