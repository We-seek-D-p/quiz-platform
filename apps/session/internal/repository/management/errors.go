package management

import "errors"

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrAlreadyFinished     = errors.New("session already finished")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrUpstreamUnavailable = errors.New("management service unavailable")
	ErrInvalidResponse     = errors.New("invalid response from management")
	ErrUnexpectedStatus    = errors.New("unexpected response status")
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
