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
	Start(wsEvents chan []byte)
	Stop()
}

func runServer(cmd *cobra.Command, args []string) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	done := make(chan interface{})
	wsEvents := make(chan []byte)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	clients, err := setup(ctx)
	if err != nil {
		logrus.Error("Failed to start server: ", err)
		os.Exit(1)
	}
	services := []Service{clients.TwitchClient}
	for _, service := range services {
		go service.Start(wsEvents)
	}

	go func() {
		for {
			select {
			case <-interrupt:
				for range interrupt {
					logrus.Warnln("Interrupt detected, flushing service")
					for _, service := range services {
						service.Stop()
					}
					ctxCancel()
					done <- "done"
				}
			}
		}
	}()

	for {
		select {
		case <-done:
			close(done)
			os.Exit(0)
		}
	}

}

func init() {
	rootCmd.AddCommand(serverCmd)
}
