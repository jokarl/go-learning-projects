package cmd

import (
	"github.com/jokarl/go-learning-projects/todo-cli/output"
	"github.com/spf13/cobra"
)

var (
	format string
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&format, "output", "o", "table", "output format (table, json)")
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list all items",
	Example: "todo list -o json",
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {
		t := taskHandler.GetAll()
		if format == "table" {
			if err := output.Table(t); err != nil {
				cmd.PrintErrf("error printing table: %v", err)
			}
		} else if format == "json" {
			if err := output.JSON(t); err != nil {
				cmd.PrintErrf("error printing json: %v", err)
			}
		}
	},
}
