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

func (h *Handler) Hub() *Hub {
	return h.hub
}

func (h *Handler) StartTimerLoop(ctx context.Context) {
	go h.timers.Run(ctx)
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

func (h *Handler) handlePlayerReconnect(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload playerReconnectPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	payload.RoomCode = strings.TrimSpace(payload.RoomCode)
	payload.ParticipantToken = strings.TrimSpace(payload.ParticipantToken)
	if payload.RoomCode == "" || payload.ParticipantToken == "" {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	h.log.DebugContext(ctx, "player_reconnect received", "connection_id", conn.ID(), "room_code", payload.RoomCode)

	result, err := h.service.PlayerReconnect(ctx, sessionservice.PlayerReconnectParams{
		RoomCode:         payload.RoomCode,
		ParticipantToken: payload.ParticipantToken,
	})
	if err != nil {
		wsErr := ToWSError(mapServicePlayerReconnectError(err))
		h.log.WarnContext(ctx, "player_reconnect failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "error_code", wsErr.Code)
		return wsErr
	}

	sessionID := result.SessionSnapshot.SessionID
	participantID := result.ParticipantID

	if err := h.hub.BindPlayer(sessionID, participantID, conn); err != nil {
		h.log.WarnContext(ctx, "player_reconnect bind failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	if err := conn.WriteEvent("session_snapshot", result.SessionSnapshot); err != nil {
		h.log.WarnContext(ctx, "player_reconnect snapshot send failed", "connection_id", conn.ID(), "room_code", payload.RoomCode, "session_id", sessionID, "participant_id", participantID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	h.log.DebugContext(ctx, "player_reconnect success", "connection_id", conn.ID(), "room_code", payload.RoomCode, "session_id", sessionID, "participant_id", participantID)
	return nil
}

func mapServicePlayerReconnectError(err error) error {
	switch {
	case errors.Is(err, sessionservice.ErrRoomNotFound):
		return NewWSError("room_not_found", "room not found")
	case errors.Is(err, sessionservice.ErrInvalidParticipantToken):
		return NewWSError("invalid_participant_token", "invalid participant token")
	case errors.Is(err, sessionservice.ErrParticipantNotFound):
		return NewWSError("participant_not_found", "participant not found")
	case errors.Is(err, sessionservice.ErrInvalidParams):
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	case errors.Is(err, sessionservice.ErrRuntimeStoreUnavailable):
		return NewWSError("internal_error", "internal error")
	default:
		return NewWSError("internal_error", "internal error")
	}
}

func (h *Handler) handleStartGame(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload startGamePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	payload.SessionID = strings.TrimSpace(payload.SessionID)
	if payload.SessionID == "" {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	h.log.DebugContext(ctx, "start_game received", "connection_id", conn.ID(), "session_id", payload.SessionID)

	result, err := h.service.StartGame(ctx, sessionservice.StartGameParams{
		SessionID:  payload.SessionID,
		HostUserID: conn.HostUserID(),
	})
	if err != nil {
		wsErr := ToWSError(mapServiceStartGameError(err))
		h.log.WarnContext(ctx, "start_game failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error_code", wsErr.Code)
		return wsErr
	}

	if err := h.hub.BindHost(payload.SessionID, conn); err != nil {
		h.log.WarnContext(ctx, "start_game bind failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	questionOpenedPayload, err := EncodeEnvelope("question_opened", result.QuestionOpened)
	if err != nil {
		h.log.WarnContext(ctx, "start_game encode question_opened failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}
	_ = h.hub.Broadcast(payload.SessionID, questionOpenedPayload)

	if err := conn.WriteEvent("session_snapshot", result.SessionSnapshot); err != nil {
		h.log.WarnContext(ctx, "start_game snapshot send failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	h.log.DebugContext(ctx, "start_game success", "connection_id", conn.ID(), "session_id", payload.SessionID)
	return nil
}

func mapServiceStartGameError(err error) error {
	switch {
	case errors.Is(err, sessionservice.ErrForbidden):
		return NewWSError("forbidden", "forbidden")
	case errors.Is(err, sessionservice.ErrGameAlreadyStarted):
		return NewWSError("game_already_started", "game already started")
	case errors.Is(err, sessionservice.ErrGameAlreadyFinished):
		return NewWSError("game_already_finished", "game already finished")
	case errors.Is(err, sessionservice.ErrInvalidParams):
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	default:
		return NewWSError("internal_error", "internal error")
	}
}

func (h *Handler) handleSubmitAnswer(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload submitAnswerPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	payload.QuestionID = strings.TrimSpace(payload.QuestionID)
	if payload.QuestionID == "" || len(payload.SelectedOptionIDs) == 0 {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
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
		wsErr := ToWSError(mapServiceSubmitAnswerError(err))
		h.log.WarnContext(ctx, "submit_answer failed", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "question_id", payload.QuestionID, "error_code", wsErr.Code)
		return wsErr
	}

	if err := conn.WriteEvent("answer_accepted", result.AnswerAccepted); err != nil {
		h.log.WarnContext(ctx, "submit_answer answer_accepted send failed", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "question_id", payload.QuestionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	if result.HostProgress != nil {
		hostProgressPayload, err := EncodeEnvelope("question_progress", result.HostProgress)
		if err != nil {
			h.log.WarnContext(ctx, "submit_answer encode question_progress failed", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "question_id", payload.QuestionID, "error", err)
			return NewWSError("internal_error", "internal error")
		}
		_ = h.hub.SendHost(sessionID, hostProgressPayload)
	}

	h.log.DebugContext(ctx, "submit_answer success", "connection_id", conn.ID(), "session_id", sessionID, "participant_id", participantID, "question_id", payload.QuestionID)
	return nil
}

func mapServiceSubmitAnswerError(err error) error {
	switch {
	case errors.Is(err, sessionservice.ErrQuestionNotActive):
		return NewWSError("question_not_active", "question not active")
	case errors.Is(err, sessionservice.ErrAnswerAlreadySubmitted):
		return NewWSError("answer_already_submitted", "answer already submitted")
	case errors.Is(err, sessionservice.ErrOptionNotInQuestion):
		return NewWSError("option_not_in_question", "option not in question")
	case errors.Is(err, sessionservice.ErrSelectionCountInvalid):
		return NewWSError("selection_count_invalid", "selection count invalid")
	case errors.Is(err, sessionservice.ErrInvalidAnswerPayload), errors.Is(err, sessionservice.ErrInvalidParams):
		return NewWSError("invalid_answer_payload", "invalid answer payload")
	case errors.Is(err, sessionservice.ErrParticipantNotFound):
		return NewWSError("participant_not_found", "participant not found")
	default:
		return NewWSError("internal_error", "internal error")
	}
}

func (h *Handler) handleFinishGame(ctx context.Context, conn *Connection, envelope MessageEnvelope) error {
	var payload finishGamePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	payload.SessionID = strings.TrimSpace(payload.SessionID)
	if payload.SessionID == "" {
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	}

	h.log.DebugContext(ctx, "finish_game received", "connection_id", conn.ID(), "session_id", payload.SessionID)

	result, err := h.service.FinishGame(ctx, sessionservice.FinishGameParams{
		SessionID:  payload.SessionID,
		HostUserID: conn.HostUserID(),
	})
	if err != nil {
		wsErr := ToWSError(mapServiceFinishGameError(err))
		h.log.WarnContext(ctx, "finish_game failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error_code", wsErr.Code)
		return wsErr
	}

	if err := h.hub.BindHost(payload.SessionID, conn); err != nil {
		h.log.WarnContext(ctx, "finish_game bind failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}

	finishedPayload, err := EncodeEnvelope("session_finished", result.SessionFinished)
	if err != nil {
		h.log.WarnContext(ctx, "finish_game encode session_finished failed", "connection_id", conn.ID(), "session_id", payload.SessionID, "error", err)
		return NewWSError("internal_error", "internal error")
	}
	_ = h.hub.Broadcast(payload.SessionID, finishedPayload)

	h.log.DebugContext(ctx, "finish_game success", "connection_id", conn.ID(), "session_id", payload.SessionID)
	return nil
}

func mapServiceFinishGameError(err error) error {
	switch {
	case errors.Is(err, sessionservice.ErrForbidden):
		return NewWSError("forbidden", "forbidden")
	case errors.Is(err, sessionservice.ErrGameAlreadyFinished):
		return NewWSError("game_already_finished", "game already finished")
	case errors.Is(err, sessionservice.ErrInvalidStateTransition):
		return NewWSError("invalid_state_transition", "invalid state transition")
	case errors.Is(err, sessionservice.ErrInvalidParams):
		return NewWSError(ErrCodeInvalidPayload, "invalid payload")
	default:
		return NewWSError("internal_error", "internal error")
	}
}
