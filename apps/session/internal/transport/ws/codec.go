package ws

import (
	"encoding/json"
)

type MessageEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// DecodeEnvelope parses and validates a raw WebSocket message envelope.
func DecodeEnvelope(raw []byte) (MessageEnvelope, error) {
	var envelope MessageEnvelope

	if err := json.Unmarshal(raw, &envelope); err != nil {
		return MessageEnvelope{}, wrapWSError(ErrCodeInvalidJSON, "invalid json", err)
	}

	if err := validateEnvelope(envelope); err != nil {
		return MessageEnvelope{}, err
	}

	return envelope, nil
}

// EncodeEnvelope wraps a typed payload into the common WebSocket envelope.
func EncodeEnvelope(messageType string, payload any) ([]byte, error) {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	envelope := MessageEnvelope{
		Type:    messageType,
		Payload: encodedPayload,
	}

	return json.Marshal(envelope)
}

// validateEnvelope verifies the common message envelope shape.
func validateEnvelope(envelope MessageEnvelope) error {
	if envelope.Type == "" {
		return NewWSError(ErrCodeInvalidEnvelope, "message type is required")
	}

	if len(envelope.Payload) == 0 {
		return NewWSError(ErrCodeInvalidPayload, "payload is required")
	}

	if !json.Valid(envelope.Payload) {
		return NewWSError(ErrCodeInvalidPayload, "payload must be valid json")
	}

	var payloadValue any
	if err := json.Unmarshal(envelope.Payload, &payloadValue); err != nil {
		return NewWSError(ErrCodeInvalidPayload, "payload must be an object")
	}

	if _, ok := payloadValue.(map[string]any); !ok {
		return NewWSError(ErrCodeInvalidPayload, "payload must be an object")
	}

	return validateMessageType(envelope.Type)
}
