package ws

import (
	"encoding/json"
)

type MessageEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

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

func EncodeEnvelope(messageType string, payload any) ([]byte, error) {
	envelope := MessageEnvelope{Type: messageType, Payload: nil}

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	envelope.Payload = encodedPayload

	return json.Marshal(envelope)
}

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

	if err := validateMessageType(envelope.Type); err != nil {
		return err
	}

	return nil
}
