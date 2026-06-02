package domain

import "fmt"

type ErrorType int

const (
	ErrTypeInternal ErrorType = iota
	ErrTypeNotFound
	ErrTypeConflict
	ErrTypeForbidden
	ErrTypeInvalidInput
)

type AppError struct {
	Type    ErrorType
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}

	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewNotFound(code, msg string, err error) *AppError {
	return &AppError{Type: ErrTypeNotFound, Code: code, Message: msg, Err: err}
}

func NewConflict(code, msg string, err error) *AppError {
	return &AppError{Type: ErrTypeConflict, Code: code, Message: msg, Err: err}
}

func NewForbidden(code, msg string, err error) *AppError {
	return &AppError{Type: ErrTypeForbidden, Code: code, Message: msg, Err: err}
}

func NewInvalidInput(code, msg string, err error) *AppError {
	return &AppError{Type: ErrTypeInvalidInput, Code: code, Message: msg, Err: err}
}

func NewInternal(code, msg string, err error) *AppError {
	return &AppError{Type: ErrTypeInternal, Code: code, Message: msg, Err: err}
}
