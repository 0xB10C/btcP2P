package btcP2P

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/btcsuite/btcd/wire"
)

var BtcnetWire = wire.MainNet
var Pver = wire.ProtocolVersion
var nonce = uint64(rand.Int63())

type Connection struct {
	Err       error
	netConn   net.Conn
	IsInbound bool
	IpAddress string
}

func newConnection(conn net.Conn, isInbound bool) *Connection {
	return &Connection{nil, conn, isInbound, conn.RemoteAddr().String()}
}

func newConnectionFailed(err error, ipAddress string) *Connection {
	return &Connection{err, nil, false, ipAddress}
}

type Peer struct {
	conn net.Conn

	// use in handlers
	StateResultChan chan *PeerState
	MsgChan         chan wire.Message

	// low level channels
	readChan       chan wire.Message
	writeChan      chan wire.Message
	writeErrorChan chan error

	// info about peer
	IsInbound    bool
	IsSPV        bool
	UserAgent    string
	FilterTweak  uint32
	CreationTime time.Time
}

func NewPeer(conn *Connection, StateResultChan chan *PeerState) *Peer {
	msgChan := make(chan wire.Message)
	writeErrorChan := make(chan error)
	writeChan := writePump(conn.netConn, writeErrorChan)
	readChan := readPump(conn.netConn)

	strisInbound := "inbound"
	if !conn.IsInbound {
		strisInbound = "outbound"
	}
	log.Println(conn.netConn.RemoteAddr(), "New ", strisInbound, "peer")

	creationTime := time.Now()
	return &Peer{conn.netConn, StateResultChan, msgChan, readChan, writeChan, writeErrorChan, conn.IsInbound, false, "", 0, creationTime}
}

func (peer *Peer) SetState(state State) {
	peer.StateResultChan <- &PeerState{peer, state}
}

func (peer *Peer) RemoteAddr() string {
	return peer.conn.RemoteAddr().String()
}

func (peer *Peer) String() string {
	return peer.RemoteAddr()
}

// reads messages from the connection and writes them to a channel
// if an error occurs, the channel is closed and the function returns
func readPump(conn net.Conn) chan wire.Message {
	ch := make(chan wire.Message)
	go func() {
		for {
			msg, _, err := wire.ReadMessage(conn, Pver, BtcnetWire)
			if err != nil {
				close(ch)
				return
			}
			ch <- msg
		}
	}()
	return ch
}

// reads messages from a channel and writes them over the connection
// if an error is encountered during sending it emits an error over the error channel
// and goes into an errState where it reads messages from the channel but
// does not try to send them until the channel is closed and then it returns
func writePump(conn net.Conn, errCh chan<- error) chan wire.Message {
	ch := make(chan wire.Message)
	go func() {
		errState := false
		for {
			msg, more := <-ch
			if !more {
				close(errCh)
				return
			}
			if errState {
				continue
			}
			err := wire.WriteMessage(conn, msg, Pver, BtcnetWire)
			if err != nil {
				errState = true
				// non blocking send, because its possible that
				// nobody is available to receive the err
				select {
				case errCh <- err:
				case <-time.Tick(5 * time.Second):
				}
			}
		}
	}()
	return ch
}

func Connect(ipAddress string, newConnectionChan chan *Connection) {
	log.Println(ipAddress, "Connection attempt")
	timeout := time.Duration(5) * time.Second
	conn, err := net.DialTimeout("tcp", ipAddress, timeout)
	if err != nil {
		log.Println(ipAddress, "Connection attempt failed:", err)
		newConnectionChan <- newConnectionFailed(err, ipAddress)
		return
	}
	newConnectionChan <- newConnection(conn, false)
}

//TODO: do something more semantically obvious to trigger disconnect
func (peer *Peer) DeliberateDisconnect() {
	go func() {
		log.Println(peer, "Disconnect")
		defer func() {
			if r := recover(); r != nil {
				// already disconnected?
				log.Println(peer, "Recovered in disconnect", r)
			}
		}()

		peer.writeErrorChan <- errors.New("Disconnect")
	}()
}

func (peer *Peer) SendSimple(msg wire.Message) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// already disconnected
				log.Println(peer, "Recovered in disconnect", r)
			}
		}()
		for {
			select {
			case peer.writeChan <- msg:
				return
			case <-time.After(15 * time.Second):
				log.Println(peer, "Send timeout")
				return
			}
		}
	}()
}

func (peer *Peer) SendOwnAddr(ipAddress string) {
	msg := wire.NewMsgAddr()
	netAddr, err := StrIpAddrToNetAddr(ipAddress)
	if err != nil {
		log.Println(err)
		return
	}

	msg.AddAddress(netAddr)
	peer.SendSimple(msg)
}

func Listen(ipAddress string, newConnectionChan chan *Connection) {
	parsedAddr, err := net.ResolveTCPAddr("tcp", ipAddress)

	listener, err := net.Listen("tcp",
		fmt.Sprintf(":%d", parsedAddr.Port))
	CheckError(err)

	log.Println("Start listening on", parsedAddr.Port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			newConnectionChan <- newConnection(conn, true)
		}
	}
}
