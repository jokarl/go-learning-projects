package cmd

import (
	"os"

	"github.com/jokarl/go-learning-projects/cidr/network"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(embedCmd)
}

var embedCmd = &cobra.Command{
	Use:     "embed",
	Short:   "Embed a v4 address in a v6 address",
	Aliases: []string{"in"},
	Example: "cidr embed 2001:db8::/32 192.0.2.33",
	Run: func(cmd *cobra.Command, args []string) {
		n, err := network.New(args[0])
		if err != nil {
			cmd.PrintErrf("Error: %s\n", err)
			os.Exit(1)
		}
		v4 := args[1:]
		for _, ip := range v4 {
			addr, err := n.Embed(ip)
			if err != nil {
				cmd.PrintErrf("Error embedding %s in %s: %s\n", ip, args[0], err)
			}
			cmd.Println(addr.String())
		}
	},
}
