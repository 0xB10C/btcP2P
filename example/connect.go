package main

import (
	"fmt"
	"log"

	".." 		// TODO: change back to "github.com/jonasnick/btcP2P" once merged
)

/*
 * adjust these settings
 */
var ownIpAddress = "127.0.0.1:8333"

// p2p
var nodeAddress = "127.0.0.1:8333"

func printResult(stateResultChan chan *btcP2P.PeerState) {
	peerState := <-stateResultChan
	peer, state := peerState.Get()
	switch state := state.(type) {
	case *btcP2P.StateEstablished:
		fmt.Println(peer, "connected (", state.UserAgent, ")")
	case *btcP2P.StateGotAddrs:
		fmt.Println(peer, "sent ", len(state.Addrs), "addresses")
	}
}

func main() {
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
	go peer.NegotiateVersionHandler(ownIpAddress, nodeAddress)
	printResult(stateResultChan)
	go peer.AddrRequest()
	printResult(stateResultChan)
}
