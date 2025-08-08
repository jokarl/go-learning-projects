package cmd

import (
	"os"
	"strconv"

	"github.com/jokarl/go-learning-projects/todo-cli/util/slice"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "delete item(s)",
	Example: "todo delete 1 2 3",
	Aliases: []string{"d"},
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
		deletedIds, err := taskHandler.Delete(ids)
		if err != nil {
			cmd.PrintErrf("error deleting item(s): %v\n", err)
			os.Exit(1)
		}

		if len(deletedIds) == 0 {
			cmd.Println("No items deleted.")
		} else {
			if len(deletedIds) < len(ids) {
				cmd.Printf("Deleted %v, but could not find %v", deletedIds, slice.Diff(ids, deletedIds))
			} else {
				cmd.Printf("Deleted %v\n", deletedIds)
			}
		}
	},
}
