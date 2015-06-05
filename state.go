package btcP2P

import "github.com/btcsuite/btcd/wire"

type State interface{}
type StateDisconnected struct{}
type StateEstablished struct {
	UserAgent string
}
type StateHasFilter struct {
	Msg *wire.MsgFilterLoad
}

type StateGotAddrs struct {
	Addrs []*wire.NetAddress
}

type PeerState struct {
	peer  *Peer
	state State
}

func (peerState *PeerState) Get() (*Peer, State) {
	return peerState.peer, peerState.state
}
