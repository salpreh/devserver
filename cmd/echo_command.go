package main

import (
	server "com.github/salpreh/devserver/pkg"
	"github.com/spf13/cobra"
)

const defaultPort = 9000

func CreateEchoCmd() *cobra.Command {
	var port int
	var echoCmd = &cobra.Command{
		Use:   "",
		Short: "Start echo server",
		Run: func(cmd *cobra.Command, args []string) {
			server.CreateEchoServer(port)
		},
	}

	echoCmd.Flags().IntVarP(&port, "port", "p", defaultPort, "Server port")

	return echoCmd
}
