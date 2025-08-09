package network

import (
	"strings"

	"github.com/jokarl/go-learning-projects/cidr/network/types"
	"github.com/jokarl/go-learning-projects/cidr/network/v4"
	"github.com/jokarl/go-learning-projects/cidr/network/v6"
)

// New creates a new ipv4Network instance from a string representation.
// It will automatically determine if it is a v4 or v6 ipv4Network based on the input format.
func New(cidr string) (types.Network, error) {
	if strings.Contains(cidr, ":") {
		return v6.NewNetwork(cidr)
	}
	return v4.NewNetwork(cidr)
}
