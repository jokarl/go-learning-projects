package cmd

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completeCmd)
}

var completeCmd = &cobra.Command{
	Use:     "complete",
	Short:   "complete item(s)",
	Example: "todo complete 1 2 3",
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		ids := make([]int, len(args))
		for i, line := range args {
			id, err := strconv.Atoi(line)
			if err != nil {
				cmd.PrintErrf("IDs must be integers: %s\n", line)
				os.Exit(1)
			}
			ids[i] = id
		}

		_, _ = taskHandler.Complete(ids)
	},
}
