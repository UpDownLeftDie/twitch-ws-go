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

type Service interface {
	Start()
	Stop(done chan interface{})
}

func runServer(cmd *cobra.Command, args []string) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	done := make(chan interface{})
	defer close(done)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	clients := setup(ctx, done, interrupt)
	services := []Service{clients.TwitchClient}
	for _, service := range services {
		go service.Start()
	}

	go func() {
		for {
			select {
			case <-interrupt:
				for range interrupt {
					logrus.Warnln("Interrupt detected, flushing service")
					for _, service := range services {
						service.Stop(done)
					}
					ctxCancel()
				}
			}
		}
	}()
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
