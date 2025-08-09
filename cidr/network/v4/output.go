package v4

import (
	"math/big"

	"github.com/jokarl/go-learning-projects/cidr/output"
)

type outputFormat struct {
	BaseAddress      string            `json:"baseAddress" tabs:"Base address"`
	BroadcastAddress string            `json:"broadcastAddress" tabs:"Broadcast address"`
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

func (n *network) Print(f output.Formatter) error {
	o := outputFormat{
		BaseAddress:      n.BaseAddress().String(),
		BroadcastAddress: n.BroadcastAddress().String(),
		UsableAddresses: usableRangeOutput{
			First: n.FirstUsableAddress().String(),
			Last:  n.LastUsableAddress().String(),
		},
		Netmask:        n.Netmask().String(),
		TotalAddresses: n.Count(),
	}
	return f.Print(o)
}
