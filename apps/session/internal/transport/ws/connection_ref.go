package ws

type connectionRef struct {
	connectionID  string
	sessionID     string
	participantID string
	role          ConnectionRole
	connection    *Connection
}

func newConnectionRef(conn *Connection) connectionRef {
	return connectionRef{
		connectionID:  conn.ID(),
		sessionID:     conn.SessionID(),
		participantID: conn.ParticipantID(),
		role:          conn.Role(),
		connection:    conn,
	}
}
