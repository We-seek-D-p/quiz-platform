package ws

import (
	"log/slog"
	"net/http"

	"github.com/coder/websocket"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/middleware"
)

type Handler struct {
	log       *slog.Logger
	readLimit int64
}

func NewHandler(cfg *config.Config, log *slog.Logger) *Handler {
	return &Handler{
		log:       log,
		readLimit: int64(cfg.WS.ReadLimitBytes),
	}
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
	wsConn.Run()
}
