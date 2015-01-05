btcP2P
===
btcP2P implements some more abstractions on top of conformals btcwire
library to make it easier to work with bitcoin's peer2peer network.

Handlers
---
Each peer is taken care of by one or more handlers.

Example:
```{go}
// sends a getaddr message and waits 
// some time to collect the responses
func (peer *Peer) AddrRequestBlocking() ([]*btcwire.NetAddress, error) {
	addrs := make([]*btcwire.NetAddress, 0)

	timeToWaitForAddr := time.Duration(5) * time.Second

	log.Println(peer, "Raw: Send getaddr")
	peer.SendSimple(btcwire.NewMsgGetAddr())
	timeout := time.After(timeToWaitForAddr)
GetAddrLoop:
	for {
		select {
		case msg, ok := <-peer.MsgChan:
			if !ok {
				return nil, errors.New("Peer disconnected unexpectedly")
			}
			switch msg := msg.(type) {
			case *btcwire.MsgAddr:
				log.Println(peer, "Raw: Received ", len(msg.AddrList), "addrs")
				for _, addr := range msg.AddrList {
					addrs = append(addrs, addr)
				}
			}
		case <-timeout:
			break GetAddrLoop
		}
	}
	return addrs, nil
}
```
It's important to check if peer.MsgChan is closed which happens in case the peer disconnects. 
If you want to disconnect the peer from within a handler call ```peer.DeliberateDisconnect```. 
You'll find some exemplary handlers in peerHandler.go.

If you do not want to return values from your handler or you have intermediary results you can set the "state" of a peer like this:
```{go}
func (peer *Peer) AddrRequest() {
	addrs, err := peer.AddrRequestBlocking()
	if err != nil {
		return
	}
	peer.SetState(&StateGotAddrs{addrs})
}
```
which sets the peers StateResultChan. 
It is intended to collect results for all peers.
For example, BasicHandlers sets a "disconnected" state once the peer disconnects.

Connection
---
The following example shows how to connect to a peer (./example/connect.go):
```{go}
newConnectionChan := make(chan *btcP2P.Connection)

// connect to node via p2p bitcoin protocol
go btcP2P.Connect(nodeAddress, newConnectionChan)
p2pConn := <-newConnectionChan
if p2pConn.Err != nil {
    log.Fatal(p2pConn.Err)
}
var peer *btcP2P.Peer
stateResultChan := make(chan *btcP2P.PeerState)
peer = btcP2P.NewPeer(p2pConn, stateResultChan)
go peer.BasicHandler()
go peer.NegotiateVersionHandler(ownIpAddress)
```

