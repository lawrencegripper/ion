package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of frontapi",
	Long:  `All software has versions. This is frontapi's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("frontapi version 0.1")
	},
}
