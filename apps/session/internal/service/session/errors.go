package session

import "errors"

var (
	ErrNotImplemented          = errors.New("not implemented")
	ErrSessionRuntimeNotFound  = errors.New("session runtime not found")
	ErrSessionRuntimeConflict  = errors.New("session runtime conflict")
	ErrSessionAlreadyFinished  = errors.New("session already finished")
	ErrBootstrapFetchFailed    = errors.New("bootstrap fetch failed")
	ErrRoomCodeUnavailable     = errors.New("room code unavailable")
	ErrRuntimeStoreUnavailable = errors.New("runtime store unavailable")
)
