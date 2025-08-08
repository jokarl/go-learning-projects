package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	description string
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&description, "desc", "d", "", "description of the item")

	requiredFlags := []string{"desc"}
	for _, r := range requiredFlags {
		if err := addCmd.MarkFlagRequired(r); err != nil {
			os.Exit(1)
		}
	}
}

var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "add an item",
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		d := cmd.Flags().Lookup("desc").Value.String()
		if err := taskHandler.Add(d); err != nil {
			cmd.PrintErrf("error adding item: %v", err)
			os.Exit(1)
		}
	},
}
