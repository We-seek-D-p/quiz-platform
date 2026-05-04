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

func sessionParticipantsKey(sessionID string) string {
	return fmt.Sprintf("session:%s:participants", sessionID)
}

func sessionParticipantTokenIndexKey(sessionID string) string {
	return fmt.Sprintf("session:%s:participant_token_idx", sessionID)
}

func sessionParticipantNicknameIndexKey(sessionID string) string {
	return fmt.Sprintf("session:%s:participant_nickname_idx", sessionID)
}

func sessionLeaderboardKey(sessionID string) string {
	return fmt.Sprintf("session:%s:leaderboard", sessionID)
}
