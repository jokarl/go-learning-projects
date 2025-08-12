package cmd

import (
	"os"

	"github.com/jokarl/go-learning-projects/cidr/network"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(countCmd)
}

var countCmd = &cobra.Command{
	Use:     "count",
	Short:   "Count addresses in a CIDR network",
	Aliases: []string{"c", "num"},
	Example: "cidr count 10.0.0.0/16",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.PrintErrln("Usage: cidr count <CIDR>")
			os.Exit(1)
		}

		for _, arg := range args {
			n, err := network.New(arg)
			if err != nil {
				cmd.PrintErrf("Error: %s\n", err)
				os.Exit(1)
			}
			cmd.Println(n.Count())
		}
	},
}
