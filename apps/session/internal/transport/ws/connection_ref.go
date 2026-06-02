package ws

type peerRef struct {
	connectionID  string
	sessionID     string
	participantID string
	role          ConnectionRole
	connection    *Connection
}

func newPeerRef(conn *Connection, sessionID, participantID string, role ConnectionRole) peerRef {
	return peerRef{
		connectionID:  conn.ID(),
		sessionID:     sessionID,
		participantID: participantID,
		role:          role,
		connection:    conn,
	}
}
