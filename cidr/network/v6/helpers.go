package v6

import (
	"bytes"
	"math/big"
	"net/netip"
	"sort"
)

func sortPrefixesV6(ps []netip.Prefix) {
	sort.Slice(ps, func(i, j int) bool {
		ai := ps[i].Masked().Addr().As16()
		aj := ps[j].Masked().Addr().As16()
		if c := bytes.Compare(ai[:], aj[:]); c != 0 {
			return c < 0
		}
		return ps[i].Bits() < ps[j].Bits()
	})
}

func splitOnceV6(p netip.Prefix) (left, right netip.Prefix) {
	if !p.Addr().Is6() {
		panic("splitOnceV6: non-IPv6 prefix")
	}
	if p.Bits() >= 128 {
		return p, p
	}
	size := new(big.Int).Lsh(big.NewInt(1), uint(128-p.Bits())) // block size
	half := new(big.Int).Rsh(size, 1)

	base := addrToBig(p.Masked().Addr())
	left = netip.PrefixFrom(bigToAddr16(base), p.Bits()+1)

	rbase := new(big.Int).Add(base, half)
	right = netip.PrefixFrom(bigToAddr16(rbase), p.Bits()+1)
	return
}

func coalesceV6(ps []netip.Prefix) []netip.Prefix {
	if len(ps) == 0 {
		return nil
	}
	// Normalize and sort.
	for i := range ps {
		ps[i] = ps[i].Masked()
	}
	sortPrefixesV6(ps)

	out := make([]netip.Prefix, 0, len(ps))
	out = append(out, ps[0])

	for i := 1; i < len(ps); i++ {
		out = append(out, ps[i])
		// Try local upward merges repeatedly.
		for len(out) >= 2 {
			a := out[len(out)-2]
			b := out[len(out)-1]
			if m, ok := tryMergePairV6(a, b); ok {
				out = out[:len(out)-2]
				out = append(out, m)
			} else {
				break
			}
		}
	}
	// One pass is typically enough thanks to sorting + local merges.
	return out
}

func tryMergePairV6(a, b netip.Prefix) (netip.Prefix, bool) {
	if a.Bits() != b.Bits() || a.Bits() == 0 {
		return netip.Prefix{}, false
	}
	ab := addrToBig(a.Masked().Addr())
	bb := addrToBig(b.Masked().Addr())
	// Ensure ab <= bb
	if ab.Cmp(bb) > 0 {
		ab, bb = bb, ab
	}

	size := new(big.Int).Lsh(big.NewInt(1), uint(128-a.Bits()))
	// Buddies if adjacent and first is aligned to 2*size.
	diff := new(big.Int).Sub(bb, ab)
	if diff.Cmp(size) != 0 {
		return netip.Prefix{}, false
	}
	twoSize := new(big.Int).Lsh(size, 1)
	if new(big.Int).Mod(ab, twoSize).Sign() != 0 {
		return netip.Prefix{}, false
	}
	return netip.PrefixFrom(bigToAddr16(ab), a.Bits()-1), true
}

func addrToBig(a netip.Addr) *big.Int {
	h := a.As16()
	return new(big.Int).SetBytes(h[:]) // big-endian
}

func bigToAddr16(x *big.Int) netip.Addr {
	var b [16]byte
	xb := x.Bytes()
	if len(xb) > 16 {
		// Clamp (shouldn't happen in valid ranges).
		xb = xb[len(xb)-16:]
	}
	copy(b[16-len(xb):], xb)
	return netip.AddrFrom16(b)
}
