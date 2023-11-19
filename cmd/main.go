package main

import (
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := createRootCmd()
	rootCmd.AddCommand(CreateEchoCmd(), CreateMockCmd(), CreateContractCmd())

	e := rootCmd.Execute()
	if e != nil {
		panic(e)
	}
}

func createRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "devserver",
		Short: "Devserver cli",
		Long:  "Devserver allows you to create echo and mock server for development",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
}
