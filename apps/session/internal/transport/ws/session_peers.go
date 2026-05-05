package ws

type sessionPeers struct {
	host    *Connection
	players map[string]*Connection
}

func newSessionPeers() *sessionPeers {
	return &sessionPeers{players: make(map[string]*Connection)}
}

func (s *sessionPeers) isEmpty() bool {
	return s.host == nil && len(s.players) == 0
}
