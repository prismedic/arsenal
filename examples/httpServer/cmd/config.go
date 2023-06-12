package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"examples/httpServer/internal/pkg/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.PrintConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\t[FATAL]\tfail to output config: %v\n", time.Now().Format(time.RFC3339), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
