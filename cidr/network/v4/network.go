package v4

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"net/netip"
	"sort"

	"github.com/jokarl/go-learning-projects/cidr/network/types"
	"github.com/jokarl/go-learning-projects/cidr/output"
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

func (n *network) BroadcastAddress() *netip.Addr {
	p := n.prefix.Masked()
	octets := p.Addr().As4()
	ip := binary.BigEndian.Uint32(octets[:])

	m4 := n.Netmask().As4()
	mask := binary.BigEndian.Uint32(m4[:])
	broadcast := ip | ^mask

	var out [4]byte
	binary.BigEndian.PutUint32(out[:], broadcast)
	addr := netip.AddrFrom4(out)
	return &addr
}

func (n *network) Netmask() netip.Addr {
	bits := n.prefix.Bits()
	if bits == 0 {
		return netip.IPv4Unspecified()
	}
	if bits == n.prefix.Addr().BitLen() {
		var all [4]byte
		for i := range all {
			all[i] = 0xFF
		}
		return netip.AddrFrom4(all)
	}

	mask := ^uint32(0) << (n.prefix.Addr().BitLen() - bits)

	var b [4]byte
	binary.BigEndian.PutUint32(b[:], mask)
	return netip.AddrFrom4(b)
}

func (n *network) FirstUsableAddress() netip.Addr {
	return n.BaseAddress().Next()
}

func (n *network) LastUsableAddress() netip.Addr {
	return n.BroadcastAddress().Prev()
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
			fmt.Println(output.Yellow, "Warning: Invalid IPv4 address:", addr, output.Reset)
			r[addr] = false
			continue
		}
		r[addr] = n.prefix.Contains(a)
	}
	return r
}

func (n *network) Embed(_ string) (netip.Addr, error) {
	return netip.Addr{}, fmt.Errorf("embedding not supported for IPv4 networks")
}

func (n *network) Divide(c int, vlsm bool) ([]netip.Prefix, error) {
	if vlsm {
		return n.divideVLSM(c)
	}

	borrowHostBits := nextPow2(c)
	newPrefix := n.prefix.Bits() + borrowHostBits
	if newPrefix > n.prefix.Addr().BitLen() {
		return nil, fmt.Errorf("prefix would exceed %d bits", n.prefix.Addr().BitLen())
	}

	a4 := n.prefix.Masked().Addr().As4()
	base := uint32(a4[0])<<24 | uint32(a4[1])<<16 | uint32(a4[2])<<8 | uint32(a4[3])
	numAddr := uint32(1) << (n.prefix.Addr().BitLen() - newPrefix)

	out := make([]netip.Prefix, c)
	for i := 0; i < c; i++ {
		mask := base + uint32(i)*numAddr
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], mask)
		out[i] = netip.PrefixFrom(netip.AddrFrom4(b), newPrefix)
	}

	return out, nil
}

func (n *network) divideVLSM(count int) ([]netip.Prefix, error) {
	orig := n.prefix.Masked()
	hostBits := n.prefix.Addr().BitLen() - orig.Bits()
	if count > (1 << hostBits) {
		return nil, fmt.Errorf("insufficient address space for %d subnets", count)
	}

	blocks := []netip.Prefix{orig}
	largestIdx := func() int {
		idx := 0
		for i := 1; i < len(blocks); i++ {
			if blocks[i].Bits() < blocks[idx].Bits() {
				idx = i
			}
		}
		return idx
	}

	for len(blocks) < count {
		i := largestIdx()
		p := blocks[i]
		if p.Bits() == n.prefix.Addr().BitLen() {
			return nil, fmt.Errorf("cannot split /32 further")
		}
		childLen := p.Bits() + 1

		base := p.Masked().Addr()
		var a = base.As4()
		size := uint32(1) << (32 - childLen)

		c1 := netip.PrefixFrom(netip.AddrFrom4(a), childLen)

		baseU := binary.BigEndian.Uint32(a[:])
		baseU2 := baseU + size
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], baseU2)
		c2 := netip.PrefixFrom(netip.AddrFrom4(b), childLen)

		blocks[i] = c1
		blocks = append(blocks, c2)
	}

	sort.Slice(blocks, func(i, j int) bool {
		bi := blocks[i].Addr().As4()
		ai := binary.BigEndian.Uint32(bi[:])
		bj := blocks[j].Addr().As4()
		aj := binary.BigEndian.Uint32(bj[:])
		if ai != aj {
			return ai < aj
		}
		return blocks[i].Bits() < blocks[j].Bits()
	})

	return blocks, nil
}

func nextPow2(c int) int {
	// In mathematics, the binary logarithm is the power to
	// which the number 2 must be raised to obtain the value c.
	// E.g. trying to divide into 8 subnets requires 3 bits (2^3 = 8).
	// E.g. trying to divide into 9 subnets requires 4 bits because:
	// 2^3 = 8 (3 borrowed bits) can fit at most 8 subnets, so we need to borrow one more bit:
	// 2^4 = 16 (4 borrowed bits) can fit 9 subnets, but also 16 subnets.
	return int(math.Ceil(math.Log2(float64(c))))
}
