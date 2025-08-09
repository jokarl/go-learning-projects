package v6

import (
	"fmt"
	"math/big"
	"net/netip"

	"github.com/jokarl/go-learning-projects/cidr/network/types"
)

type network struct {
	prefix netip.Prefix
}

// NewNetwork creates a new IPv6 network from a CIDR string.
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

func (n *network) BroadcastAddress() *netip.Addr {
	return nil // IPv6 does not have a broadcast address
}

func (n *network) Netmask() netip.Addr {
	bits := n.prefix.Bits()
	if bits == 0 {
		return netip.IPv6Unspecified()
	}
	if bits == 128 {
		var all [16]byte
		for i := range all {
			all[i] = 0xFF
		}
		return netip.AddrFrom16(all)
	}

	var mask [16]byte
	full := bits >> 3 // number of full 0xFF bytes (e.g. bits / 2^3)
	rem := bits & 7   // remaining bits in the next byte (e.g. bits % 2^3)

	for i := 0; i < full; i++ {
		mask[i] = 0xFF
	}
	if rem != 0 {
		mask[full] = 0xFF << (8 - rem)
	}
	return netip.AddrFrom16(mask)
}

func (n *network) FirstUsableAddress() netip.Addr {
	return n.BaseAddress().Next()
}

func (n *network) LastUsableAddress() netip.Addr {
	addr := n.prefix.Addr()
	bits := n.prefix.Bits()
	addrBytes := addr.As16()

	hostBits := addr.BitLen() - bits

	// Set all host bits to 1
	for i := 15; i >= 0 && hostBits > 0; i-- {
		if hostBits >= 8 {
			addrBytes[i] = 0xFF
			hostBits -= 8
		} else {
			// Set only the remaining bits
			addrBytes[i] |= (1 << hostBits) - 1
			hostBits = 0
		}
	}

	lastAddr := netip.AddrFrom16(addrBytes)
	return lastAddr.Prev() // Subtract 1 to get the last usable address
}

func (n *network) Count() *big.Int {
	hostBits := n.prefix.Addr().BitLen() - n.prefix.Bits()
	return big.NewInt(0).Lsh(big.NewInt(1), uint(hostBits))
}

func (n *network) Contains(addrs []string) map[string]bool {
	r := make(map[string]bool, len(addrs))
	for _, addr := range addrs {
		a, err := netip.ParseAddr(addr)
		if err != nil || !a.Is6() {
			fmt.Println("Warning: Invalid IPv6 address:", addr)
			r[addr] = false
			continue
		}
		r[addr] = n.prefix.Contains(a)
	}
	return r
}
