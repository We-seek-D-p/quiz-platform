package ws

func validateMessageType(messageType string) error {
	switch messageType {
	case "host_connect", "player_join", "player_reconnect", "start_game", "submit_answer", "finish_game":
		return nil
	default:
		return NewWSError(ErrCodeUnknownMessageType, "unknown message type")
	}
}
