package types

import (
	"math/big"
	"net/netip"
)

// Network represents an IP network with methods for calculating network properties.
type Network interface {
	// BaseAddress returns the network address (first address in the network).
	BaseAddress() netip.Addr

	// BroadcastAddress returns the broadcast address (last address in the network).
	// If the network is IPv6, this is not applicable and will return nil.
	// If the network is IPv4, it will return the broadcast address.
	BroadcastAddress() *netip.Addr

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

	// Divide divides the network into smaller subnets.
	// The `int` parameter specifies the number of subnets to create.
	// If bool is true, it will use Variable Length Subnet Masking (VLSM).
	Divide(int, bool) ([]netip.Prefix, error)

	// Embed embeds an IPv4 address into an IPv6 address.
	// This will return an error if the network is not an IPv6 network.
	// See https://www.rfc-editor.org/rfc/rfc6052.html#section-2.2
	Embed(string) (netip.Addr, error)
}
