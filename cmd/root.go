package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   "etcdproxy-apiserver",
	Short: "etcd-gPRC Proxy API server prototype",
	Long:  `Set up an etcd-gRPC proxy to a namespace to access your data.`,
}

// Execute prepares and executes the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	addPersistentFlags()
	addCommands()
}

func addPersistentFlags() {
	//rootCmd.PersistentFlags().IntVarP(&logger.Level, "verbose", "v", 4, "Log level")
}

func addCommands() {
	rootCmd.AddCommand(startCmd)
}