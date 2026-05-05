package ws

func dispatchIncomingMessage(envelope MessageEnvelope) error {
	switch envelope.Type {
	case "host_connect":
		return NewWSError(ErrCodeUnknownMessageType, "message type is not implemented yet")
	case "player_join":
		return NewWSError(ErrCodeUnknownMessageType, "message type is not implemented yet")
	case "player_reconnect":
		return NewWSError(ErrCodeUnknownMessageType, "message type is not implemented yet")
	case "start_game":
		return NewWSError(ErrCodeUnknownMessageType, "message type is not implemented yet")
	case "submit_answer":
		return NewWSError(ErrCodeUnknownMessageType, "message type is not implemented yet")
	case "finish_game":
		return NewWSError(ErrCodeUnknownMessageType, "message type is not implemented yet")
	default:
		return NewWSError(ErrCodeUnknownMessageType, "unknown message type")
	}
}

func validateMessageType(messageType string) error {
	switch messageType {
	case "host_connect", "player_join", "player_reconnect", "start_game", "submit_answer", "finish_game":
		return nil
	default:
		return NewWSError(ErrCodeUnknownMessageType, "unknown message type")
	}
}
