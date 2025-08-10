package network

import (
	"math/big"

	"github.com/jokarl/go-learning-projects/cidr/network/types"
	"github.com/jokarl/go-learning-projects/cidr/output"
)

type outputFormat struct {
	BaseAddress      string            `json:"baseAddress" tabs:"Base address"`
	BroadcastAddress *string           `json:"broadcastAddress,omitempty" tabs:"Broadcast address,omitempty"`
	Netmask          string            `json:"netmask" tabs:"Netmask"`
	UsableAddresses  usableRangeOutput `json:"usableAddresses" tabs:"Usable addresses"`
	TotalAddresses   *big.Int          `json:"totalAddresses" tabs:"Total addresses"`
}

type usableRangeOutput struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

func (u usableRangeOutput) String() string {
	return u.First + " - " + u.Last
}

// PrintNetwork formats network information using the provided formatter
func PrintNetwork(n types.Network, f output.Formatter) error {
	o := outputFormat{
		BaseAddress: n.BaseAddress().String(),
		UsableAddresses: usableRangeOutput{
			First: n.FirstUsableAddress().String(),
			Last:  n.LastUsableAddress().String(),
		},
		Netmask:        n.Netmask().String(),
		TotalAddresses: n.Count(),
	}

	// Only set broadcast address for IPv4 networks
	if broadcastAddr := n.BroadcastAddress(); broadcastAddr != nil {
		broadcastStr := broadcastAddr.String()
		o.BroadcastAddress = &broadcastStr
	}

	return f.Print(o)
}
