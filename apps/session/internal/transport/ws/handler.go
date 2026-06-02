package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	sessionservice "github.com/We-seek-D-p/quiz-platform/apps/session/internal/service/session"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/middleware"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/ws/dto"
)

var roomCodePattern = regexp.MustCompile(`^\d{6}$`)

type Handler struct {
	log       *slog.Logger
	readLimit int64
	hub       *Hub
	service   *sessionservice.Service
	timers    *timerLoop
}

func NewHandler(cfg *config.Config, log *slog.Logger, service *sessionservice.Service) *Handler {
	hub := NewHub(log)

	return &Handler{
		log:       log,
		readLimit: int64(cfg.WS.ReadLimitBytes),
		hub:       hub,
		service:   service,
		timers:    newTimerLoop(log, service, hub),
	}
}

// Hub exposes the in-memory connection registry for internal lifecycle operations.
func (h *Handler) Hub() *Hub {
	return h.hub
}

// StartTimerLoop starts background timer processing for active sessions.
func (h *Handler) StartTimerLoop(ctx context.Context) {
	go h.timers.Run(ctx)
}

// Host upgrades an authenticated host request to a WebSocket connection.
func (h *Handler) Host(w http.ResponseWriter, r *http.Request) {
	h.acceptAndServe(w, r, ConnectionRoleHost)
}

// Player upgrades an anonymous player request to a WebSocket connection.
func (h *Handler) Player(w http.ResponseWriter, r *http.Request) {
	h.acceptAndServe(w, r, ConnectionRolePlayer)
}

// acceptAndServe owns the lifecycle of a single accepted WebSocket connection.
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

		if conn.Role() != ConnectionRolePlayer || conn.SessionID() == "" || conn.ParticipantID() == "" {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := h.service.SetParticipantConnected(ctx, conn.SessionID(), conn.ParticipantID(), false); err != nil {
			h.log.WarnContext(ctx, "player presence update failed", "session_id", conn.SessionID(), "participant_id", conn.ParticipantID(), "error", err)
			return
		}

		h.broadcastLobbyUpdate(ctx, conn.SessionID())
	})
	wsConn.SetMessageHandler(h.dispatchIncomingMessage)
	wsConn.Run()
}

// broadcastLobbyUpdate sends presence and snapshot changes after player disconnects.
func (h *Handler) broadcastLobbyUpdate(ctx context.Context, sessionID string) {
	snapshot, err := h.service.GetSessionSnapshotForTimer(ctx, sessionID)
	if err != nil {
		h.log.WarnContext(ctx, "lobby update snapshot failed", "session_id", sessionID, "error", err)
		return
	}

	lobbyPayload, err := EncodeEnvelope("lobby_updated", dto.LobbyUpdated{PlayersCount: len(snapshot.Participants)})
	if err == nil {
		_ = h.hub.Broadcast(sessionID, lobbyPayload)
	}

	snapshotPayload, err := EncodeEnvelope("session_snapshot", mapSnapshot(snapshot))
	if err == nil {
		_ = h.hub.Broadcast(sessionID, snapshotPayload)
	}
}

type hostConnectPayload struct {
	SessionID string `json:"session_id"`
}

type playerJoinPayload struct {
	RoomCode string `json:"room_code"`
	Nickname string `json:"nickname"`
}

type playerReconnectPayload struct {
	RoomCode         string `json:"room_code"`
	ParticipantToken string `json:"participant_token"`
}

type startGamePayload struct {
	SessionID string `json:"session_id"`
}

type submitAnswerPayload struct {
	QuestionID        string   `json:"question_id"`
	SelectedOptionIDs []string `json:"selected_option_ids"`
}

type finishGamePayload struct {
	SessionID string `json:"session_id"`
}

// Validate checks host connection payload shape before service execution.
func (p *hostConnectPayload) Validate() error {
	p.SessionID = strings.TrimSpace(p.SessionID)
	if _, err := uuid.Parse(p.SessionID); err != nil {
		return domain.NewInvalidInput("invalid_payload", "session_id must be a valid uuid", err)
	}
	return nil
}

// Validate checks player join payload shape before service execution.
func (p *playerJoinPayload) Validate() error {
	p.RoomCode = strings.TrimSpace(p.RoomCode)
	p.Nickname = strings.TrimSpace(p.Nickname)
	if !roomCodePattern.MatchString(p.RoomCode) || p.Nickname == "" {
		return domain.NewInvalidInput("invalid_payload", "room_code and nickname are required", nil)
	}
	return nil
}

// Validate checks reconnect payload shape before service execution.
func (p *playerReconnectPayload) Validate() error {
	p.RoomCode = strings.TrimSpace(p.RoomCode)
	p.ParticipantToken = strings.TrimSpace(p.ParticipantToken)
	if !roomCodePattern.MatchString(p.RoomCode) || p.ParticipantToken == "" {
		return domain.NewInvalidInput("invalid_payload", "room_code and participant_token are required", nil)
	}
	return nil
}

// Validate checks start game payload shape before service execution.
func (p *startGamePayload) Validate() error {
	p.SessionID = strings.TrimSpace(p.SessionID)
	if _, err := uuid.Parse(p.SessionID); err != nil {
		return domain.NewInvalidInput("invalid_payload", "session_id must be a valid uuid", err)
	}
	return nil
}

// Validate checks answer submission payload shape before service execution.
func (p *submitAnswerPayload) Validate() error {
	p.QuestionID = strings.TrimSpace(p.QuestionID)
	if _, err := uuid.Parse(p.QuestionID); err != nil {
		return domain.NewInvalidInput("invalid_answer_payload", "question_id must be a valid uuid", err)
	}
	if len(p.SelectedOptionIDs) == 0 {
		return domain.NewInvalidInput("invalid_answer_payload", "selected_option_ids cannot be empty", nil)
	}

	seen := make(map[string]struct{}, len(p.SelectedOptionIDs))
	for i, optionID := range p.SelectedOptionIDs {
		optionID = strings.TrimSpace(optionID)
		if _, err := uuid.Parse(optionID); err != nil {
			return domain.NewInvalidInput("invalid_answer_payload", "selected option id must be a valid uuid", err)
		}
		if _, exists := seen[optionID]; exists {
			return domain.NewInvalidInput("selection_count_invalid", "selected option ids must be unique", nil)
		}
		seen[optionID] = struct{}{}
		p.SelectedOptionIDs[i] = optionID
	}
	return nil
}

// Validate checks finish game payload shape before service execution.
func (p *finishGamePayload) Validate() error {
	p.SessionID = strings.TrimSpace(p.SessionID)
	if _, err := uuid.Parse(p.SessionID); err != nil {
		return domain.NewInvalidInput("invalid_payload", "session_id must be a valid uuid", err)
	}
	return nil
}

// writeMessageError converts service errors into WebSocket error envelopes.
func (h *Handler) writeMessageError(ctx context.Context, err error) error {
	wsErr := ToWSError(err)
	if wsErr == nil {
		return nil
	}

	if appErr, ok := errors.AsType[*domain.AppError](err); ok {
		if appErr.Type == domain.ErrTypeInvalidInput {
			h.log.DebugContext(ctx, "ws validation rejected", "code", appErr.Code)
		} else {
			h.log.WarnContext(ctx, "ws domain error", "code", appErr.Code, "error", appErr.Error())
		}
		return wsErr
	}

	h.log.ErrorContext(ctx, "ws unhandled error", "error", err)
	return NewWSError("internal_error", "internal error")
}

// dispatchIncomingMessage routes a decoded client envelope to its handler.
func (h *Handler) dispatchIncomingMessage(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	switch envelope.Type {
	case "host_connect":
		return h.handleHostConnect(ctx, conn, envelope)
	case "player_join":
		return h.handlePlayerJoin(ctx, conn, envelope)
	case "player_reconnect":
		return h.handlePlayerReconnect(ctx, conn, envelope)
	case "start_game":
		return h.handleStartGame(ctx, conn, envelope)
	case "submit_answer":
		return h.handleSubmitAnswer(ctx, conn, envelope)
	case "finish_game":
		return h.handleFinishGame(ctx, conn, envelope)
	default:
		return NewWSError(ErrCodeUnknownMessageType, "unknown message type")
	}
}

func (h *Handler) handleHostConnect(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload hostConnectPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	if err := payload.Validate(); err != nil {
		return h.writeMessageError(ctx, err)
	}

	h.log.DebugContext(ctx, "host_connect received", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID)

	result, err := h.service.HostConnect(ctx, sessionservice.HostConnectParams{
		SessionID:  payload.SessionID,
		HostUserID: conn.HostUserID(),
	})
	if err != nil {
		wsErr := ToWSError(h.writeMessageError(ctx, err))
		h.log.WarnContext(ctx, "host_connect failed", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID, "error_code", wsErr.Code)
		return wsErr
	}

	if err := h.hub.BindHost(payload.SessionID, conn); err != nil {
		h.log.WarnContext(ctx, "host_connect bind failed", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	if err := conn.WriteEvent("session_snapshot", mapSnapshot(result.SessionSnapshot)); err != nil {
		h.log.WarnContext(ctx, "host_connect snapshot send failed", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	h.log.DebugContext(ctx, "host_connect success", "connection_id", conn.ID(), "host_user_id", conn.HostUserID(), "session_id", payload.SessionID)
	return nil
}

func (h *Handler) handlePlayerJoin(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload playerJoinPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	if err := payload.Validate(); err != nil {
		return h.writeMessageError(ctx, err)
	}

	h.log.DebugContext(ctx, "player_join received", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname)

	result, err := h.service.PlayerJoin(ctx, sessionservice.PlayerJoinParams{
		RoomCode: payload.RoomCode,
		Nickname: payload.Nickname,
	})
	if err != nil {
		wsErr := ToWSError(h.writeMessageError(ctx, err))
		h.log.WarnContext(ctx, "player_join failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "error_code", wsErr.Code)
		return wsErr
	}

	sessionID := result.SessionSnapshot.Runtime.SessionID
	participantID := result.JoinedLobby.ParticipantID
	if err := h.hub.BindPlayer(sessionID, participantID, conn); err != nil {
		h.log.WarnContext(ctx, "player_join bind failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	if err := conn.WriteEvent("joined_lobby", mapJoinedLobby(result.JoinedLobby)); err != nil {
		h.log.WarnContext(ctx, "player_join joined_lobby send failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	lobbyUpdatedPayload, err := EncodeEnvelope("lobby_updated", mapLobbyUpdated(result.LobbyUpdated))
	if err != nil {
		h.log.WarnContext(ctx, "player_join lobby_updated encode failed", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}
	_ = h.hub.Broadcast(sessionID, lobbyUpdatedPayload)

	if err := conn.WriteEvent("session_snapshot", mapSnapshot(result.SessionSnapshot)); err != nil {
		h.log.WarnContext(ctx, "player_join snapshot send failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	h.log.DebugContext(ctx, "player_join success", "connection_id", conn.ID(), "room_code", payload.RoomCode, "nickname", payload.Nickname, "session_id", sessionID, "participant_id", participantID)
	return nil
}

func (h *Handler) handlePlayerReconnect(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload playerReconnectPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	if err := payload.Validate(); err != nil {
		return h.writeMessageError(ctx, err)
	}

	h.log.DebugContext(ctx, "player_reconnect received", "connection_id", conn.ID(), "room_code", payload.RoomCode)

	result, err := h.service.PlayerReconnect(ctx, sessionservice.PlayerReconnectParams{
		RoomCode:         payload.RoomCode,
		ParticipantToken: payload.ParticipantToken,
	})
	if err != nil {
		wsErr := ToWSError(h.writeMessageError(ctx, err))
		h.log.WarnContext(ctx, "player_reconnect failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "error_code", wsErr.Code)
		return wsErr
	}

	sessionID := result.SessionSnapshot.Runtime.SessionID
	participantID := result.ParticipantID

	if err := h.hub.BindPlayer(sessionID, participantID, conn); err != nil {
		h.log.WarnContext(ctx, "player_reconnect bind failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	if err := conn.WriteEvent("session_snapshot", mapSnapshot(result.SessionSnapshot)); err != nil {
		h.log.WarnContext(ctx, "player_reconnect snapshot send failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	h.log.DebugContext(ctx, "player_reconnect success", "connection_id", conn.ID(), "room_code", payload.RoomCode, "session_id", sessionID, "participant_id", participantID)
	return nil
}

func (h *Handler) handleStartGame(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload startGamePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	if err := payload.Validate(); err != nil {
		return h.writeMessageError(ctx, err)
	}

	h.log.DebugContext(ctx, "start_game received", "connection_id", conn.ID(), "session_id", payload.SessionID)

	result, err := h.service.StartGame(ctx, sessionservice.StartGameParams{
		SessionID:  payload.SessionID,
		HostUserID: conn.HostUserID(),
	})
	if err != nil {
		wsErr := ToWSError(h.writeMessageError(ctx, err))
		h.log.WarnContext(ctx, "start_game failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error_code", wsErr.Code)
		return wsErr
	}

	if err := h.hub.BindHost(payload.SessionID, conn); err != nil {
		h.log.WarnContext(ctx, "start_game bind failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	questionOpenedPayload, err := EncodeEnvelope("question_opened", mapQuestionOpened(result.QuestionOpened))
	if err != nil {
		h.log.WarnContext(ctx, "start_game encode question_opened failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}
	_ = h.hub.Broadcast(payload.SessionID, questionOpenedPayload)

	if err := conn.WriteEvent("session_snapshot", mapSnapshot(result.SessionSnapshot)); err != nil {
		h.log.WarnContext(ctx, "start_game snapshot send failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	h.log.DebugContext(ctx, "start_game success", "connection_id", conn.ID(), "session_id", payload.SessionID)
	return nil
}

func (h *Handler) handleSubmitAnswer(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload submitAnswerPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	if err := payload.Validate(); err != nil {
		return h.writeMessageError(ctx, err)
	}

	sessionID := strings.TrimSpace(conn.SessionID())
	participantID := strings.TrimSpace(conn.ParticipantID())
	if sessionID == "" || participantID == "" {
		return NewWSError("invalid_participant_token", "invalid participant token")
	}

	h.log.DebugContext(ctx, "submit_answer received", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "question_id", payload.QuestionID)

	result, err := h.service.SubmitAnswer(ctx, sessionservice.SubmitAnswerParams{
		SessionID:         sessionID,
		ParticipantID:     participantID,
		QuestionID:        payload.QuestionID,
		SelectedOptionIDs: payload.SelectedOptionIDs,
	})
	if err != nil {
		wsErr := ToWSError(h.writeMessageError(ctx, err))
		h.log.WarnContext(ctx, "submit_answer failed", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "question_id", payload.QuestionID, "error_code", wsErr.Code)
		return wsErr
	}

	if err := conn.WriteEvent("answer_accepted", mapAnswerAccepted(result.AnswerAccepted)); err != nil {
		h.log.WarnContext(ctx, "submit_answer answer_accepted send failed", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "question_id", payload.QuestionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	if result.HostProgress != nil {
		hostProgressPayload, err := EncodeEnvelope("question_progress", mapQuestionProgress(*result.HostProgress))
		if err != nil {
			h.log.WarnContext(ctx, "submit_answer encode question_progress failed", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "question_id", payload.QuestionID, "error", err)
			return NewWSError("internal_error", "internal error")
		}
		_ = h.hub.SendHost(sessionID, hostProgressPayload)
	}

	h.log.DebugContext(ctx, "submit_answer success", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "question_id", payload.QuestionID)
	return nil
}

func (h *Handler) handleFinishGame(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload finishGamePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	if err := payload.Validate(); err != nil {
		return h.writeMessageError(ctx, err)
	}

	h.log.DebugContext(ctx, "finish_game received", "connection_id", conn.ID(), "session_id", payload.SessionID)

	result, err := h.service.FinishGame(ctx, sessionservice.FinishGameParams{
		SessionID:  payload.SessionID,
		HostUserID: conn.HostUserID(),
	})
	if err != nil {
		wsErr := ToWSError(h.writeMessageError(ctx, err))
		h.log.WarnContext(ctx, "finish_game failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error_code", wsErr.Code)
		return wsErr
	}

	if err := h.hub.BindHost(payload.SessionID, conn); err != nil {
		h.log.WarnContext(ctx, "finish_game bind failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	finishedPayload, err := EncodeEnvelope("session_finished", mapFinished(result.SessionFinished))
	if err != nil {
		h.log.WarnContext(ctx, "finish_game encode session_finished failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}
	_ = h.hub.Broadcast(payload.SessionID, finishedPayload)

	h.log.DebugContext(ctx, "finish_game success", "connection_id", conn.ID(), "session_id", payload.SessionID)
	return nil
}
