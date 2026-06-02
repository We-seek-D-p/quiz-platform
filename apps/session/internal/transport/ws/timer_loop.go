package ws

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	sessionservice "github.com/We-seek-D-p/quiz-platform/apps/session/internal/service/session"
)

const timerTickInterval = 250 * time.Millisecond
const maxConcurrentTimerSessions = 32

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

type playerEventTarget struct {
	participantID string
	payload       any
}

func newTimerLoop(log logger, service *sessionservice.Service, hub *Hub) *timerLoop {
	return &timerLoop{
		log:        log,
		service:    service,
		hub:        hub,
		processing: make(map[string]struct{}),
	}
}

// Run starts periodic timer processing until the context is canceled.
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

// tick schedules active sessions for bounded concurrent timer processing.
func (l *timerLoop) tick(ctx context.Context) {
	sessionIDs := l.hub.ActiveSessionIDs()
	semaphore := make(chan struct{}, maxConcurrentTimerSessions)
	for _, sessionID := range sessionIDs {
		sessionID = strings.TrimSpace(sessionID)
		if sessionID == "" {
			continue
		}

		select {
		case semaphore <- struct{}{}:
		case <-ctx.Done():
			return
		}

		go func(sessionID string) {
			defer func() { <-semaphore }()

			if !l.acquire(sessionID) {
				l.log.DebugContext(ctx, "timer_transition_skipped", "session_id", sessionID, "reason", "already_processing")
				return
			}
			defer l.release(sessionID)

			l.processSession(ctx, sessionID)
		}(sessionID)
	}
}

// processSession delegates timer decisions to the service state machine.
func (l *timerLoop) processSession(ctx context.Context, sessionID string) {
	events, err := l.service.HandleTick(ctx, sessionID, time.Now().UTC())
	if err != nil {
		if isAppErrorCode(err, "session_runtime_not_found") {
			l.log.DebugContext(ctx, "timer_transition_skipped", "session_id", sessionID, "reason", "session_not_found")
			return
		}

		l.log.WarnContext(ctx, "timer_transition_failed", "session_id", sessionID, "error_code", "internal_error", "error", err)
		return
	}

	for _, event := range events {
		l.dispatchDomainEvent(ctx, event)
	}
}

// dispatchDomainEvent maps a domain event to WebSocket notifications.
func (l *timerLoop) dispatchDomainEvent(ctx context.Context, event domain.SessionDomainEvent) {
	switch event.Type {
	case domain.EventAnswerRevealed:
		revealResult, ok := event.Payload.(sessionservice.RevealTransitionResult)
		if !ok {
			l.log.WarnContext(ctx, "timer_transition_failed", "session_id", event.SessionID, "error_code", "invalid_payload")
			return
		}
		playerTargets := buildPlayerTargets(revealResult.PlayerReveals, func(reveal sessionservice.ParticipantAnswerReveal) string {
			return reveal.ParticipantID
		}, func(reveal sessionservice.ParticipantAnswerReveal) any {
			return mapAnswerReveal(reveal.Payload)
		})
		l.emitTransitionEvents(event.SessionID, "question_reveal_host", mapQuestionRevealHost(revealResult.HostReveal), "answer_reveal", playerTargets, mapSnapshot(revealResult.SessionSnapshot))
	case domain.EventLeaderboardRevealed:
		revealResult, ok := event.Payload.(sessionservice.LeaderboardTransitionResult)
		if !ok {
			l.log.WarnContext(ctx, "timer_transition_failed", "session_id", event.SessionID, "error_code", "invalid_payload")
			return
		}
		playerTargets := buildPlayerTargets(revealResult.PlayerReveals, func(reveal sessionservice.ParticipantLeaderboardReveal) string {
			return reveal.ParticipantID
		}, func(reveal sessionservice.ParticipantLeaderboardReveal) any {
			return mapLeaderboardReveal(reveal.Payload)
		})
		l.emitTransitionEvents(event.SessionID, "leaderboard_reveal_host", mapLeaderboardRevealHost(revealResult.HostReveal), "leaderboard_reveal", playerTargets, mapSnapshot(revealResult.SessionSnapshot))
	case domain.EventQuestionOpened, domain.EventGameFinished:
		snapshot, ok := event.Payload.(sessionservice.Snapshot)
		if !ok {
			l.log.WarnContext(ctx, "timer_transition_failed", "session_id", event.SessionID, "error_code", "invalid_payload")
			return
		}
		l.emitSnapshotEvent(ctx, event.SessionID, snapshot)
	}
}

// buildPlayerTargets prepares per-player payloads for targeted broadcasts.
func buildPlayerTargets[T any](items []T, participantID func(T) string, payload func(T) any) []playerEventTarget {
	targets := make([]playerEventTarget, 0, len(items))
	for i := range items {
		item := items[i]
		targets = append(targets, playerEventTarget{
			participantID: participantID(item),
			payload:       payload(item),
		})
	}

	return targets
}

// emitSnapshotEvent broadcasts question opening or final session events.
func (l *timerLoop) emitSnapshotEvent(ctx context.Context, sessionID string, snapshot sessionservice.Snapshot) {
	if snapshot.Runtime.Status == domain.RuntimeStatusQuestionOpen {
		if snapshot.CurrentQuestion == nil || snapshot.Runtime.Progress.DeadlineAt == nil {
			l.log.WarnContext(ctx, "timer_transition_failed", "session_id", sessionID, "to_status", snapshot.Runtime.Status, "reason", "reveal_timeout", "error_code", "invalid_payload")
			return
		}

		payload := sessionservice.QuestionOpened{
			QuestionIndex:  snapshot.Runtime.Progress.CurrentQuestionIndex,
			TotalQuestions: snapshot.Runtime.Progress.TotalQuestions,
			Question:       *snapshot.CurrentQuestion,
			DeadlineAt:     *snapshot.Runtime.Progress.DeadlineAt,
		}

		questionPayload, encErr := EncodeEnvelope("question_opened", mapQuestionOpened(payload))
		if encErr == nil {
			_ = l.hub.Broadcast(sessionID, questionPayload)
		}

		snapshotPayload, encErr := EncodeEnvelope("session_snapshot", mapSnapshot(snapshot))
		if encErr == nil {
			_ = l.hub.Broadcast(sessionID, snapshotPayload)
		}

		l.log.DebugContext(ctx, "timer_transition_succeeded", "session_id", sessionID, "to_status", snapshot.Runtime.Status, "reason", "reveal_timeout")
		return
	}

	if snapshot.Runtime.Status == domain.RuntimeStatusFinished {
		finishedPayload := sessionservice.Finished{LeaderboardTop: snapshot.LeaderboardTop}
		encodedFinished, encErr := EncodeEnvelope("session_finished", mapFinished(finishedPayload))
		if encErr == nil {
			_ = l.hub.Broadcast(sessionID, encodedFinished)
		}

		snapshotPayload, encErr := EncodeEnvelope("session_snapshot", mapSnapshot(snapshot))
		if encErr == nil {
			_ = l.hub.Broadcast(sessionID, snapshotPayload)
		}

		l.log.DebugContext(ctx, "timer_transition_succeeded", "session_id", sessionID, "to_status", snapshot.Runtime.Status, "reason", "reveal_timeout")
		return
	}

	l.log.DebugContext(ctx, "timer_transition_skipped", "session_id", sessionID, "to_status", snapshot.Runtime.Status, "reason", "reveal_timeout")
}

// emitTransitionEvents sends host, player, and snapshot notifications for a phase transition.
func (l *timerLoop) emitTransitionEvents(
	sessionID string,
	hostEvent string,
	hostPayload any,
	playerEvent string,
	playerTargets []playerEventTarget,
	snapshot any,
) {
	encodedHost, err := EncodeEnvelope(hostEvent, hostPayload)
	if err == nil {
		_ = l.hub.SendHost(sessionID, encodedHost)
	}

	for i := range playerTargets {
		playerTarget := playerTargets[i]
		encodedPlayer, encodeErr := EncodeEnvelope(playerEvent, playerTarget.payload)
		if encodeErr != nil {
			continue
		}
		_ = l.hub.SendPlayer(sessionID, playerTarget.participantID, encodedPlayer)
	}

	encodedSnapshot, err := EncodeEnvelope("session_snapshot", snapshot)
	if err == nil {
		_ = l.hub.Broadcast(sessionID, encodedSnapshot)
	}
}

// acquire prevents overlapping timer work for the same session.
func (l *timerLoop) acquire(sessionID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.processing[sessionID]; exists {
		return false
	}

	l.processing[sessionID] = struct{}{}
	return true
}

// release marks a session timer job as finished.
func (l *timerLoop) release(sessionID string) {
	l.mu.Lock()
	delete(l.processing, sessionID)
	l.mu.Unlock()
}
