package btcP2P

import (
	"fmt"
	"net"

	"github.com/conformal/btcwire"
)

// a slice of net addresses is ordered based on the timestamp
type NetAddressSlice []*btcwire.NetAddress

func (s NetAddressSlice) Len() int {
	return len(s)
}
func (s NetAddressSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s NetAddressSlice) Less(i, j int) bool {
	return s[i].Timestamp.Before(s[j].Timestamp)
}

func AddrToString(addr *btcwire.NetAddress) string {
	return fmt.Sprintf("%s:%d", addr.IP, addr.Port)
}

func StrIpAddrToNetAddr(ipAddress string) (*btcwire.NetAddress, error) {
	parsedAddr, err := net.ResolveTCPAddr("tcp", ipAddress)
	if err != nil {
		return nil, err
	}
	netAddr, err := btcwire.NewNetAddress(parsedAddr, 1)
	if err != nil {
		return nil, err
	}
	return netAddr, err
}

func CheckError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
