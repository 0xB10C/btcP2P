package btcP2P

import (
	"errors"
	"log"
	"net"
	"time"

	"github.com/conformal/btcwire"
)

// handles standard messages like ping and forwards interesting messages over the msgChan
// returns if it receives a write error or if the readChan closes
func (peer *Peer) BasicHandler() {
OuterLoop:
	for {
		select {
		case <-peer.writeErrorChan:
			break OuterLoop
		case msg, ok := <-peer.readChan:
			if !ok {
				break OuterLoop
			}
			switch msg := msg.(type) {
			case *btcwire.MsgPing:
				peer.SendSimple(btcwire.NewMsgPong(msg.Nonce))
			case btcwire.Message:
				peer.MsgChan <- msg
			default:
				// disconnect
				break OuterLoop
			}
		}
	}
	log.Println(peer, "Lost connection")
	peer.conn.Close()
	// closing write chan is dangerous, because we are not
	// the sender here, but sendSimple is going to catch that
	close(peer.writeChan)
	close(peer.MsgChan)
	peer.SetState(&StateDisconnected{})
}

func MyNewMsgVersionFromConn(conn net.Conn, ownIpAddress string) (*btcwire.MsgVersion, error) {
	msg, err := btcwire.NewMsgVersionFromConn(conn, nonce, 332753)
	if err != nil {
		return nil, err
	}
	msg.Services = 1

	addrMe, err := StrIpAddrToNetAddr(ownIpAddress)
	if err != nil {
		return nil, err
	}

	msg.AddrMe = *addrMe
	return msg, nil
}

func (peer *Peer) NegotiateVersionHandler(ownIpAddress string) {
	msg, err := MyNewMsgVersionFromConn(peer.conn, ownIpAddress)
	CheckError(err)
	var otherVersion *btcwire.MsgVersion
	otherVerack := false

	log.Println(peer, "Send version")
	peer.SendSimple(msg)

	for {
		select {
		case msg, ok := <-peer.MsgChan:
			if !ok {
				return
			}
			switch msg := msg.(type) {
			case *btcwire.MsgVersion:
				log.Println(peer, "UserAgent: ", msg.UserAgent)

				if msg.Nonce == nonce {
					peer.DeliberateDisconnect()
					log.Println(peer, "is myself")
					return
				}

				otherVersion = msg
				peer.SendSimple(btcwire.NewMsgVerAck())

				if otherVerack {
					peer.SetState(&StateEstablished{otherVersion.UserAgent})
					return
				}
			case *btcwire.MsgVerAck:
				log.Println(peer, "Raw: verack")
				if otherVersion != nil {
					peer.SetState(&StateEstablished{otherVersion.UserAgent})
					return
				}
				otherVerack = true
			default:
				continue
			}
		case <-time.After(15 * time.Second):
			log.Println(peer, "Version Timeout")
			peer.DeliberateDisconnect()
			return
		}
	}
}

// sends a getaddr message and waits
// some time to collect the responses
func (peer *Peer) AddrRequestBlocking() ([]*btcwire.NetAddress, error) {
	addrs := make([]*btcwire.NetAddress, 0)

	timeToWaitForAddr := time.Duration(10) * time.Second

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

func (peer *Peer) AddrRequest() {
	addrs, err := peer.AddrRequestBlocking()
	if err != nil {
		return
	}
	peer.SetState(&StateGotAddrs{addrs})
}
