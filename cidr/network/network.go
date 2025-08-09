package network

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/jokarl/go-learning-projects/cidr/network/types"
	"github.com/jokarl/go-learning-projects/cidr/network/v4"
	"github.com/jokarl/go-learning-projects/cidr/network/v6"
)

// New creates a new ipv4Network instance from a string representation.
// It will automatically determine if it is a v4 or v6 ipv4Network based on the input format.
func New(cidr string) (types.Network, error) {
	cidr = strings.TrimSpace(cidr)
	p, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, err
	}
	if p.Addr().Is4() {
		return v4.NewNetwork(cidr)
	}
	if p.Addr().Is6() {
		return v6.NewNetwork(cidr)
	}
	return nil, fmt.Errorf("unsupported address type: %s", cidr)
}
