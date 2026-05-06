package ws

import (
	"errors"
	"fmt"
)

const (
	ErrCodeInvalidJSON        = "invalid_json"
	ErrCodeInvalidEnvelope    = "invalid_envelope"
	ErrCodeInvalidPayload     = "invalid_payload"
	ErrCodeUnknownMessageType = "unknown_message_type"

	ServerEventError = "error"
)

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}

	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NewWSError(code, message string) error {
	return &Error{Code: code, Message: message}
}

func wrapWSError(code, message string, err error) error {
	return &Error{Code: code, Message: message, Err: err}
}

func ToWSError(err error) *Error {
	if err == nil {
		return nil
	}

	if wsErr, ok := errors.AsType[*Error](err); ok {
		return wsErr
	}

	return &Error{Code: ErrCodeInvalidEnvelope, Message: "invalid message", Err: err}
}
