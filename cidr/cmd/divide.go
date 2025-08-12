package cmd

import (
	"context"
	"os"
	"strconv"

	"github.com/jokarl/go-learning-projects/cidr/network"
	"github.com/jokarl/go-learning-projects/cidr/network/types"
	"github.com/jokarl/go-learning-projects/cidr/output"
	"github.com/spf13/cobra"
)

var vlsm bool

func init() {
	rootCmd.AddCommand(divideCmd)
	divideCmd.Flags().BoolP("vlsm", "v", false, "Use Variable Length Subnet Masking (VLSM) to divide the CIDR into subnets of different sizes")
}

var divideCmd = &cobra.Command{
	Use:     "divide",
	Short:   "Divide a CIDR into smaller subnets",
	Aliases: []string{"d"},
	Example: "cidr divide 10.0.0.0/16 4",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.PrintErrln("Usage: cidr divide <CIDR> [<subnet count>]")
			os.Exit(1)
		}

		// Validate CIDR format
		n, err := network.New(args[0])
		if err != nil {
			cmd.PrintErrf("Invalid CIDR: %s\n", err)
			os.Exit(1)
		}

		c, err := strconv.Atoi(args[1])
		if err != nil {
			cmd.PrintErrf("Invalid subnet count: %s\n", err)
			os.Exit(1)
		}

		if c <= 0 {
			cmd.PrintErrf("count must be > 0")
		}

		vlsm, _ = cmd.Flags().GetBool("vlsm")
		if c&(c-1) != 0 && !vlsm {
			cmd.Println(output.Yellow, "Warning: count is not a power of two; extra subnets will be unused. Use --vlsm.", output.Reset)
		}

		cmd.SetContext(context.WithValue(cmd.Context(), "validatedCount", c))
		cmd.SetContext(context.WithValue(cmd.Context(), "validatedNetwork", n))
		cmd.SetContext(context.WithValue(cmd.Context(), "vlsm", vlsm))
	},
	Run: func(cmd *cobra.Command, args []string) {
		n := cmd.Context().Value("validatedNetwork").(types.Network)
		count := cmd.Context().Value("validatedCount").(int)
		subnets, err := n.Divide(count, cmd.Context().Value("vlsm").(bool))
		if err != nil {
			cmd.PrintErrf("Could not divide: %s\n", err)
			os.Exit(1)
		}

		for _, subnet := range subnets {
			cmd.Println(subnet.String())
		}
	},
}
