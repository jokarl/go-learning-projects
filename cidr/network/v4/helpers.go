package v4

import (
	"encoding/binary"
	"net/netip"
	"sort"
)

func sortPrefixesV4(ps []netip.Prefix) {
	sort.Slice(ps, func(i, j int) bool {
		ai := u32(ps[i].Masked().Addr())
		aj := u32(ps[j].Masked().Addr())
		if ai != aj {
			return ai < aj
		}
		return ps[i].Bits() < ps[j].Bits()
	})
}

func splitOnceV4(p netip.Prefix) (left, right netip.Prefix) {
	if !p.Addr().Is4() {
		panic("splitOnceV4: non-IPv4 prefix")
	}
	if p.Bits() >= p.Addr().BitLen() {
		return p, p
	}
	size := uint32(1) << uint(32-p.Bits())
	half := size >> 1
	base := u32(p.Masked().Addr())
	left = netip.PrefixFrom(addr4(base), p.Bits()+1)
	right = netip.PrefixFrom(addr4(base+half), p.Bits()+1)
	return
}

func coalesceV4(ps []netip.Prefix) []netip.Prefix {
	if len(ps) == 0 {
		return nil
	}
	// Normalize to masked, sort by addr then bits.
	for i := range ps {
		ps[i] = ps[i].Masked()
	}
	sortPrefixesV4(ps)

	out := make([]netip.Prefix, 0, len(ps))
	out = append(out, ps[0])

	for i := 1; i < len(ps); i++ {
		out = append(out, ps[i])
		// Attempt repeated upward merges. Each pass may enable a higher-level merge.
		for len(out) >= 2 {
			a := out[len(out)-2]
			b := out[len(out)-1]
			m, ok := tryMergePairV4(a, b)
			if !ok {
				break
			}
			// Replace last two with merged one.
			out = out[:len(out)-2]
			out = append(out, m)
			// After a merge, we may be able to merge again with the previous element.
			// So continue inner loop.
		}
	}
	// If new merge opportunities exist due to ordering, do another pass.
	// Usually one pass is enough because we keep it sorted and merge locally.
	// But to be safe, detect if another pass changes length.
	coalesced := out
	if len(coalesced) < len(ps) {
		return coalesceV4(coalesced)
	}
	return coalesced
}

func tryMergePairV4(a, b netip.Prefix) (netip.Prefix, bool) {
	if a.Bits() != b.Bits() {
		return netip.Prefix{}, false
	}
	if a.Bits() == 0 {
		return netip.Prefix{}, false
	}
	size := uint32(1) << uint(32-a.Bits())
	uA := u32(a.Masked().Addr())
	uB := u32(b.Masked().Addr())
	if uA > uB {
		uA, uB = uB, uA
	}
	// Buddies if they are adjacent and the first is aligned to 2*size.
	if uB-uA != size {
		return netip.Prefix{}, false
	}
	if (uA%(2*size) != 0) || (uB%(2*size) != size) {
		return netip.Prefix{}, false
	}
	return netip.PrefixFrom(addr4(uA), a.Bits()-1), true
}

func u32(a netip.Addr) uint32 {
	v4 := a.As4()
	return binary.BigEndian.Uint32(v4[:])
}

func addr4(u uint32) netip.Addr {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], u)
	return netip.AddrFrom4(b)
}
