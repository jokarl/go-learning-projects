package v4

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"net/netip"

	"github.com/jokarl/go-learning-projects/cidr/network/types"
)

type network struct {
	prefix netip.Prefix
}

// NewNetwork creates a new IPv4 network from a CIDR string.
func NewNetwork(s string) (types.Network, error) {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR format: %w", err)
	}

	return &network{
		prefix: p,
	}, nil
}

func (n *network) BaseAddress() netip.Addr {
	return n.prefix.Masked().Addr()
}

func (n *network) BroadcastAddress() netip.Addr {
	p := n.prefix.Masked()
	octets := p.Addr().As4()
	ip := binary.BigEndian.Uint32(octets[:])

	m4 := n.Netmask().As4()
	mask := binary.BigEndian.Uint32(m4[:])
	broadcast := ip | ^mask

	var out [4]byte
	binary.BigEndian.PutUint32(out[:], broadcast)
	return netip.AddrFrom4(out)
}

func (n *network) Netmask() netip.Addr {
	mask := ^uint32(0) << (n.prefix.Addr().BitLen() - n.prefix.Bits())

	var b [4]byte
	binary.BigEndian.PutUint32(b[:], mask)
	return netip.AddrFrom4(b)
}

func (n *network) FirstUsableAddress() netip.Addr {
	return n.BaseAddress().Next()
}

func (n *network) LastUsableAddress() netip.Addr {
	broadcast := n.BroadcastAddress()
	return broadcast.Prev()
}

func (n *network) Count() *big.Int {
	hostBits := n.prefix.Addr().BitLen() - n.prefix.Bits()
	return big.NewInt(0).Lsh(big.NewInt(1), uint(hostBits))
}

func (n *network) Contains(addrs []string) map[string]bool {
	r := make(map[string]bool, len(addrs))
	for _, addr := range addrs {
		a, err := netip.ParseAddr(addr)
		if err != nil || !a.Is4() {
			fmt.Println("Warning: Invalid IPv4 address", addr, "-")
		}
		r[addr] = n.prefix.Contains(a)
	}
	return r
}
