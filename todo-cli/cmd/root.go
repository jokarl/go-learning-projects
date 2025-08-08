package cmd

import (
	"os"

	"github.com/jokarl/go-learning-projects/todo-cli/tasks"
	"github.com/spf13/cobra"
)

// taskHandler is a global variable to hold the task handler instance
var taskHandler *tasks.Handler

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "todo-cli",
	Short:            "A todo list for the terminal",
	TraverseChildren: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	var err error
	taskHandler, err = tasks.NewHandler(&tasks.HandlerOptions{
		Path: "tasks.csv",
	})

	if err != nil {
		os.Exit(1)
	}
}
