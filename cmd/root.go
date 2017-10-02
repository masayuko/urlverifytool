package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// RootCmd reporesents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "urlverifytool",
	Short: "A manipulation tool for verifying URL data",
	Long:  `urlverifytool is a manipulation tool for verifying URL data.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize()
}
