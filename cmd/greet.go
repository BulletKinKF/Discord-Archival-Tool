package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var greetCmd = &cobra.Command{
	Use:   "greet",
	Short: "Print a greeting",
	Long:  "Print a friendly greeting to the terminal",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from Cobra 🐍")
	},
}

var name string

func init() {
	greetCmd.Flags().StringVarP(&name, "name", "n", "world", "name to greet")
	rootCmd.AddCommand(greetCmd)
}
