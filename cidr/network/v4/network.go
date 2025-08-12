package v4

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"net/netip"
	"sort"

	"github.com/jokarl/go-learning-projects/cidr/math"
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
	ip := u32(p.Addr())

	mask := u32(n.Netmask())
	broadcast := ip | ^mask

	addr := addr4(broadcast)
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

	borrowHostBits := math.NextPow2(c)
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
	orig := n.prefix.Masked()                          // 10.0.0.0/16
	hostBits := n.prefix.Addr().BitLen() - orig.Bits() // 32 - 16 = 16
	if count > (1 << uint(hostBits)) {                 // count = 5, 1 << 16 = 65536, so 5 > 65536
		return nil, fmt.Errorf("insufficient address space for %d subnets", count)
	}

	blocks := []netip.Prefix{orig} // [10.0.0.0/16]
	largestIdx := func() int {     // On first run, returns 0 with prefix 10.0.0.0/16
		idx := 0
		for i := 1; i < len(blocks); i++ {
			if blocks[i].Bits() < blocks[idx].Bits() {
				idx = i
			}
		}
		return idx
	}

	for len(blocks) < count {
		i := largestIdx() // 0
		p := blocks[i]    // 10.0.0/16
		if p.Bits() == n.prefix.Addr().BitLen() {
			return nil, fmt.Errorf("cannot split /32 further")
		}

		childLen := p.Bits() + 1  // 16 --> 17
		base := p.Masked().Addr() // 10.0.0.0
		var a = base.As4()

		c1 := netip.PrefixFrom(netip.AddrFrom4(a), childLen) // 10.0.0.0/17

		baseU := binary.BigEndian.Uint32(a[:])                // 167772160 (10.0.0.0 as uint32)
		size := uint32(1) << uint(p.Addr().BitLen()-childLen) // 2^(32-17) = 2^15 = 32768
		baseU2 := baseU + size                                // 167772160 + 32768 = 167804928 (10.0.128.0)
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], baseU2) // Convert to byte array with 4 bytes (10.0.128.0)
		c2 := netip.PrefixFrom(netip.AddrFrom4(b), childLen)

		blocks[i] = c1              // 10.0.0.0/17
		blocks = append(blocks, c2) // 10.0.128.0/17
		// Next round will split the largest block again
		// in this case 10.0.0.0/17 because both have the same amount of bits
		// but appears first in the slice.
	}

	// Sort the blocks by address and bits
	// Larger blocks (with more bits) come first, then smaller blocks
	sort.Slice(blocks, func(i, j int) bool {
		ai := blocks[i].Addr().As4()
		aj := blocks[j].Addr().As4()
		if cmp := bytes.Compare(ai[:], aj[:]); cmp != 0 {
			return cmp < 0
		}
		return blocks[i].Bits() < blocks[j].Bits()
	})

	return blocks, nil
}

func (n *network) VLSM(hostCounts []int) (allocated, leftover []netip.Prefix, err error) {
	if !n.prefix.Addr().Is4() {
		return nil, nil, fmt.Errorf("VLSM: IPv4 only")
	}

	// Copy + sort host counts (largest first) so big blocks get placed early.
	counts := append([]int(nil), hostCounts...)
	sort.Slice(counts, func(i, j int) bool { return counts[i] > counts[j] })

	free := []netip.Prefix{n.prefix.Masked()}

	for _, hosts := range counts {
		if hosts <= 0 {
			return nil, nil, fmt.Errorf("host count must be > 0 (got %d)", hosts)
		}

		// Classic VLSM: hosts + 2 for network+broadcast
		required := hosts + 2
		needBits := math.NextPow2(required) // 2^n >= required
		wantLen := 32 - needBits
		if wantLen < 0 {
			return nil, nil, fmt.Errorf("host count %d too large for IPv4", hosts)
		}

		// Best-fit: pick the smallest free block that can produce wantLen (max Bits() subject to Bits() <= wantLen).
		idx := -1
		bestBits := -1
		for i, b := range free {
			if b.Bits() <= wantLen && b.Bits() > bestBits {
				idx = i
				bestBits = b.Bits()
			}
		}
		if idx == -1 {
			return nil, nil, fmt.Errorf("insufficient address space for %d hosts", hosts)
		}

		// Pop chosen block.
		block := free[idx]
		free = append(free[:idx], free[idx+1:]...)

		// Split only as needed; keep right siblings as free space.
		cur := block
		for cur.Bits() < wantLen {
			l, r := splitOnceV4(cur)
			// Take left path for allocation; keep the right as free.
			free = append(free, r)
			cur = l
		}
		allocated = append(allocated, cur)
	}

	// Clean up + coalesce remaining free space for a nice, compact remainder view.
	leftover = coalesceV4(free)

	// Sort outputs: by address, then by prefix length (shorter prefix first).
	sortPrefixesV4(allocated)
	sortPrefixesV4(leftover)

	return allocated, leftover, nil
}
