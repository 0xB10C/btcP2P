package btcP2P

import "github.com/conformal/btcwire"

type State interface{}
type StateDisconnected struct{}
type StateEstablished struct {
	UserAgent string
}
type StateHasFilter struct {
	Msg *btcwire.MsgFilterLoad
}

type StateGotAddrs struct {
	Addrs []*btcwire.NetAddress
}

type PeerState struct {
	peer  *Peer
	state State
}

func (peerState *PeerState) Get() (*Peer, State) {
	return peerState.peer, peerState.state
}
