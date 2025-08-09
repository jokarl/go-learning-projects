package cmd

import (
	"os"

	"github.com/jokarl/go-learning-projects/cidr/network"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(containsCmd)
}

var containsCmd = &cobra.Command{
	Use:     "contains",
	Short:   "Check if a network contains specific addresses",
	Aliases: []string{"in"},
	Example: "cidr contains 10.0.0.0/16 10.0.0.1 10.0.0.2",
	Run: func(cmd *cobra.Command, args []string) {
		n, err := network.New(args[0])
		if err != nil {
			cmd.PrintErrf("Error: %s\n", err)
			os.Exit(1)
		}
		r := n.Contains(args[1:])
		for i, _ := range r {
			if r[i] {
				cmd.Printf("%s is contained in %s\n", i, args[0])
			} else {
				cmd.Printf("%s is NOT contained in %s\n", i, args[0])
			}
		}
	},
}
