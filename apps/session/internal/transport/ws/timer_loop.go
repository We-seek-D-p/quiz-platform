package ws

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	sessionservice "github.com/We-seek-D-p/quiz-platform/apps/session/internal/service/session"
)

const timerTickInterval = 250 * time.Millisecond

type timerLoop struct {
	log     logger
	service *sessionservice.Service
	hub     *Hub

	mu         sync.Mutex
	processing map[string]struct{}
}

type logger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

func newTimerLoop(log logger, service *sessionservice.Service, hub *Hub) *timerLoop {
	return &timerLoop{
		log:        log,
		service:    service,
		hub:        hub,
		processing: make(map[string]struct{}),
	}
}

func (l *timerLoop) Run(ctx context.Context) {
	ticker := time.NewTicker(timerTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.tick(ctx)
		}
	}
}

func (l *timerLoop) tick(ctx context.Context) {
	sessionIDs := l.hub.ActiveSessionIDs()
	for _, sessionID := range sessionIDs {
		sessionID = strings.TrimSpace(sessionID)
		if sessionID == "" {
			continue
		}

		if !l.acquire(sessionID) {
			l.log.DebugContext(ctx, "timer_transition_skipped", "session_id", sessionID, "reason", "already_processing")
			continue
		}

		l.processSession(ctx, sessionID)
		l.release(sessionID)
	}
}

func (l *timerLoop) processSession(ctx context.Context, sessionID string) {
	snapshot, err := l.service.GetSessionSnapshotForTimer(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sessionservice.ErrSessionRuntimeNotFound) {
			l.log.DebugContext(ctx, "timer_transition_skipped", "session_id", sessionID, "reason", "session_not_found")
			return
		}

		l.log.WarnContext(ctx, "timer_transition_failed", "session_id", sessionID, "error_code", "internal_error", "error", err)
		return
	}

	now := time.Now().UTC()
	status := snapshot.Status

	if status == string(domain.RuntimeStatusQuestionOpen) && snapshot.DeadlineAt != nil && !snapshot.DeadlineAt.After(now) {
		l.handleDeadlineTransition(ctx, sessionID, status)
		return
	}

	if status == string(domain.RuntimeStatusAnswerReveal) && snapshot.RevealUntil != nil && !snapshot.RevealUntil.After(now) {
		l.handleRevealTimeoutTransition(ctx, sessionID, status)
	}
}

func (l *timerLoop) handleDeadlineTransition(ctx context.Context, sessionID, fromStatus string) {
	l.log.DebugContext(ctx, "timer_transition_started", "session_id", sessionID, "from_status", fromStatus, "reason", "deadline")

	revealResult, err := l.service.CloseCurrentQuestionAndBuildReveal(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sessionservice.ErrInvalidStateTransition) {
			l.log.DebugContext(ctx, "timer_transition_skipped", "session_id", sessionID, "from_status", fromStatus, "reason", "deadline", "error_code", "invalid_state_transition")
			return
		}

		l.log.WarnContext(ctx, "timer_transition_failed", "session_id", sessionID, "from_status", fromStatus, "reason", "deadline", "error_code", "internal_error", "error", err)
		return
	}

	hostPayload, err := EncodeEnvelope("question_reveal_host", revealResult.HostReveal)
	if err == nil {
		_ = l.hub.SendHost(sessionID, hostPayload)
	}

	for i := range revealResult.PlayerReveals {
		playerReveal := &revealResult.PlayerReveals[i]
		playerPayload, encodeErr := EncodeEnvelope("answer_reveal", playerReveal.Payload)
		if encodeErr != nil {
			continue
		}
		_ = l.hub.SendPlayer(sessionID, playerReveal.ParticipantID, playerPayload)
	}

	snapshotPayload, err := EncodeEnvelope("session_snapshot", revealResult.SessionSnapshot)
	if err == nil {
		_ = l.hub.Broadcast(sessionID, snapshotPayload)
	}

	l.log.DebugContext(ctx, "timer_transition_succeeded", "session_id", sessionID, "from_status", fromStatus, "to_status", string(domain.RuntimeStatusAnswerReveal), "reason", "deadline")
}

func (l *timerLoop) handleRevealTimeoutTransition(ctx context.Context, sessionID, fromStatus string) {
	l.log.DebugContext(ctx, "timer_transition_started", "session_id", sessionID, "from_status", fromStatus, "reason", "reveal_timeout")

	snapshot, err := l.service.AdvanceToNextQuestion(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sessionservice.ErrInvalidStateTransition) {
			l.log.DebugContext(ctx, "timer_transition_skipped", "session_id", sessionID, "from_status", fromStatus, "reason", "reveal_timeout", "error_code", "invalid_state_transition")
			return
		}

		l.log.WarnContext(ctx, "timer_transition_failed", "session_id", sessionID, "from_status", fromStatus, "reason", "reveal_timeout", "error_code", "internal_error", "error", err)
		return
	}

	if snapshot.Status == string(domain.RuntimeStatusQuestionOpen) {
		if snapshot.CurrentQuestion == nil || snapshot.DeadlineAt == nil {
			l.log.WarnContext(ctx, "timer_transition_failed", "session_id", sessionID, "from_status", fromStatus, "to_status", snapshot.Status, "reason", "reveal_timeout", "error_code", "invalid_payload")
			return
		}

		payload := sessionservice.QuestionOpenedDTO{
			QuestionIndex:  snapshot.CurrentQuestionIndex,
			TotalQuestions: snapshot.TotalQuestions,
			Question:       *snapshot.CurrentQuestion,
			DeadlineAt:     *snapshot.DeadlineAt,
		}

		questionPayload, encErr := EncodeEnvelope("question_opened", payload)
		if encErr == nil {
			_ = l.hub.Broadcast(sessionID, questionPayload)
		}

		snapshotPayload, encErr := EncodeEnvelope("session_snapshot", snapshot)
		if encErr == nil {
			_ = l.hub.Broadcast(sessionID, snapshotPayload)
		}

		l.log.DebugContext(ctx, "timer_transition_succeeded", "session_id", sessionID, "from_status", fromStatus, "to_status", snapshot.Status, "reason", "reveal_timeout")
		return
	}

	if snapshot.Status == string(domain.RuntimeStatusFinished) {
		finishedPayload := sessionservice.FinishedDTO{LeaderboardTop: snapshot.LeaderboardTop}
		encodedFinished, encErr := EncodeEnvelope("session_finished", finishedPayload)
		if encErr == nil {
			_ = l.hub.Broadcast(sessionID, encodedFinished)
		}

		snapshotPayload, encErr := EncodeEnvelope("session_snapshot", snapshot)
		if encErr == nil {
			_ = l.hub.Broadcast(sessionID, snapshotPayload)
		}

		l.log.DebugContext(ctx, "timer_transition_succeeded", "session_id", sessionID, "from_status", fromStatus, "to_status", snapshot.Status, "reason", "reveal_timeout")
		return
	}

	l.log.DebugContext(ctx, "timer_transition_skipped", "session_id", sessionID, "from_status", fromStatus, "to_status", snapshot.Status, "reason", "reveal_timeout")
}

func (l *timerLoop) acquire(sessionID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.processing[sessionID]; exists {
		return false
	}

	l.processing[sessionID] = struct{}{}
	return true
}

func (l *timerLoop) release(sessionID string) {
	l.mu.Lock()
	delete(l.processing, sessionID)
	l.mu.Unlock()
}
