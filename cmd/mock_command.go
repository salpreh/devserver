package main

import (
	server "com.github/salpreh/devserver/pkg"
	"github.com/spf13/cobra"
)

func CreateMockCmd() *cobra.Command {
	var port int
	mockCmd := &cobra.Command{
		Use:   "mock",
		Short: "Start a mock server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config := args[0]
			server.CreateMockServer(port, config)
		},
	}

	mockCmd.Flags().IntVarP(&port, "port", "p", 9000, "Server port")

	return mockCmd
}
