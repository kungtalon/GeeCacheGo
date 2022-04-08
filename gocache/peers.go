package gocache

// PeerPicker locates the peer that owns the key
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter must be implemented by a peer with Get
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
