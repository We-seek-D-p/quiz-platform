package redis

import "fmt"

func sessionMetaKey(sessionID string) string {
	return fmt.Sprintf("session:%s:meta", sessionID)
}

func sessionQuizSnapshotKey(sessionID string) string {
	return fmt.Sprintf("session:%s:quiz_snapshot", sessionID)
}

func roomCodeKey(roomCode string) string {
	return fmt.Sprintf("room_code:%s", roomCode)
}
