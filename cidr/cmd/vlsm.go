package cmd

import (
	"fmt"
	"os"

	"github.com/jokarl/go-learning-projects/cidr/network"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(vlsmCmd)
}

var vlsmCmd = &cobra.Command{
	Use:     "vlsm",
	Short:   "VLSM takes a CIDR and divides it into smaller subnets based on the number of hosts required.",
	Aliases: []string{"v"},
	Example: "cidr vlsm 10.0.0.0/16 120 60 30 10",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.PrintErrln("Usage: cidr vlsm <CIDR> <host count> <host count 2>...")
			os.Exit(1)
		}

		n, err := network.New(args[0])
		if err != nil {
			cmd.PrintErrf("Error: %s\n", err)
			os.Exit(1)
		}

		hostNum := args[1:]
		integers := make([]int, len(hostNum))
		for i, h := range hostNum {
			var num int
			if _, err := fmt.Sscanf(h, "%d", &num); err != nil {
				cmd.PrintErrf("Invalid host count: %s\n", h)
				os.Exit(1)
			}
			integers[i] = num
		}

		allocated, leftover, err := n.VLSM(integers)
		if err != nil {
			cmd.PrintErrf("Error: %s\n", err)
			os.Exit(1)
		}

		if len(allocated) == 0 {
			cmd.PrintErrln("No subnets allocated. Please check the host counts and CIDR.")
			os.Exit(1)
		}

		if len(allocated) > 0 {
			cmd.Println("Allocated subnets:")
			for _, a := range allocated {
				cmd.Printf("  %s\n", a)
			}
		}

		if len(leftover) > 0 {
			cmd.Printf("Leftover subnets:\n")
			for _, l := range leftover {
				cmd.Printf("  %s\n", l)
			}
		}
	},
}
