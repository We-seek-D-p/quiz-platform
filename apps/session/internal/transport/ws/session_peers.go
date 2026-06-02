package ws

type sessionPeers struct {
	host    *peerRef
	players map[string]*peerRef
}

func newSessionPeers() *sessionPeers {
	return &sessionPeers{players: make(map[string]*peerRef)}
}

func (s *sessionPeers) isEmpty() bool {
	return s.host == nil && len(s.players) == 0
}
