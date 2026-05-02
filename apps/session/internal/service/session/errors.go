package session

import "errors"

var (
	ErrNotImplemented          = errors.New("not implemented")
	ErrSessionNotFound         = errors.New("session not found in management")
	ErrSessionRuntimeNotFound  = errors.New("session runtime not found")
	ErrSessionRuntimeConflict  = errors.New("session runtime conflict")
	ErrSessionAlreadyFinished  = errors.New("session already finished")
	ErrBootstrapFetchFailed    = errors.New("failed to fetch bootstrap data")
	ErrRoomCodeUnavailable     = errors.New("could not generate unique room code")
	ErrRuntimeStoreUnavailable = errors.New("runtime storage unavailable")
)
