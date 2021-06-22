package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "http server to allow Oauth Authorization",
	Long:  ``,
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	done := make(chan interface{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	clients := setup(ctx, done, interrupt)

	go func() {
		for range interrupt {
			logrus.Warnln("Interrupt detected, flushing service")
			clients.TwitchClient.TwitchWsConn.Close()
			ctxCancel()
		}
	}()
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
