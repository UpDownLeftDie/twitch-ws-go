package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/service"
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

	serviceService := setup(ctx)
	mux := service.MakeHTTPHandler(serviceService)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("PORT")),
		Handler: mux,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			logrus.Warnln("Interrupt detected, flushing service")
			server.Shutdown(context.TODO())
			ctxCancel()
		}
	}()

	server.ListenAndServe()
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
