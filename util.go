package btcP2P

import (
	"fmt"
	"net"

	"github.com/btcsuite/btcd/wire"
)

// a slice of net addresses is ordered based on the timestamp
type NetAddressSlice []*wire.NetAddress

func (s NetAddressSlice) Len() int {
	return len(s)
}
func (s NetAddressSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s NetAddressSlice) Less(i, j int) bool {
	return s[i].Timestamp.Before(s[j].Timestamp)
}

func AddrToString(addr *wire.NetAddress) string {
	return fmt.Sprintf("%s:%d", addr.IP, addr.Port)
}

func StrIpAddrToNetAddr(ipAddress string) (*wire.NetAddress, error) {
	parsedAddr, err := net.ResolveTCPAddr("tcp", ipAddress)
	if err != nil {
		return nil, err
	}
	netAddr := wire.NewNetAddress(parsedAddr, 1)
	return netAddr, err
}

func CheckError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
