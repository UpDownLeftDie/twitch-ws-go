package cmd

import (
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/updownleftdie/twitch-ws-go/v2/shared"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "TODO set this description",
	Long:  ``,
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	//ctx, ctxCancel := context.WithCancel(context.Background())
	done := make(chan interface{})
	interrupt := make(chan os.Signal, 1)
	defer close(interrupt)
	signal.Notify(interrupt, os.Interrupt)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Cmd:              exec.Command("./examples/twitch/plugin-twitch"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("twitch")
	if err != nil {
		log.Fatal(err)
	}
	twitchPlugin := raw.(shared.CustomPlugin)
	twitchPlugin.Impl.Start()


	//for _, plugin := range plugins {
	//	go plugin.Start(db, eventChan)
	//}

	go func() {
		for {
			select {
			case <-interrupt:
				for range interrupt {
					logrus.Warnln("Interrupt detected, stopping plugins")
					twitchPlugin.Impl.Stop()
					//for _, plugin := range plugins {
					//	plugin.Stop()
					//}
					//ctxCancel(
				}
				done <- "done"
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
