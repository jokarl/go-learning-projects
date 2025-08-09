package types

import (
	"math/big"
	"net/netip"

	"github.com/jokarl/go-learning-projects/cidr/output"
)

// Network represents an IP network with methods for calculating network properties.
type Network interface {
	// BaseAddress returns the network address (first address in the network).
	BaseAddress() netip.Addr

	// BroadcastAddress returns the broadcast address (last address in the network).
	BroadcastAddress() netip.Addr

	// Netmask returns the subnet mask for this network.
	Netmask() netip.Addr

	// FirstUsableAddress returns the first host address in the network.
	FirstUsableAddress() netip.Addr

	// LastUsableAddress returns the last host address in the network.
	LastUsableAddress() netip.Addr

	// Count returns the total number of addresses in this network.
	Count() *big.Int

	// Contains checks if the network contains the specified IP addresses.
	Contains([]string) map[string]bool

	// Print formats the network information using the provided formatter.
	Print(formatter output.Formatter) error
}
