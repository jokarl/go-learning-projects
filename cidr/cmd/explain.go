package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jokarl/go-learning-projects/cidr/network"
	"github.com/jokarl/go-learning-projects/cidr/output"
	"github.com/spf13/cobra"
)

var format string

func init() {
	rootCmd.AddCommand(explainCmd)

	explainCmd.Flags().StringVarP(
		&format,
		"out",
		"o",
		output.DefaultFormat, // default to "tab"
		fmt.Sprintf("Output format (%s)", strings.Join(output.Formats(), ", ")),
	)
}

var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain CIDR notation",
	Long: `Explain CIDR notation by providing the base address of the network.
It is possible to pass any number of CIDR notated networks, and mixing v4 and v6 addresses.`,
	Aliases: []string{"e"},
	Example: `cidr explain 10.0.0.0/16
cidr explain 2001:db8::/32`,
	Run: func(cmd *cobra.Command, args []string) {
		f, err := output.GetFormatter(format)
		if err != nil {
			cmd.PrintErrf("Unknown output format: %s\n", format)
			os.Exit(1)
		}

		for _, arg := range args {
			n, err := network.New(arg)
			if err != nil {
				cmd.PrintErrf("Error: %s\n", err)
				os.Exit(1)
			}
			if err := network.PrintNetwork(n, f); err != nil {
				cmd.PrintErrf("Error printing network %s: %s\n", arg, err)
				os.Exit(1)
			}
		}
	},
}
