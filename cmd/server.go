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
	clientPlugins, err := plugin.Discover("plugin-*", "plugins/**")
	if err != nil {
		log.Fatal("Error getting plugins: ", err)
	}
	if len(clientPlugins) < 1 {
		log.Fatal("Didn't find any plugins in 'plugins/' dir that start with 'plugin-'")
	}
	log.Printf("Found plugins: %s", clientPlugins)

	type LoadedPlugin struct {
		*plugin.Client
		Plugin *shared.CustomPlugin
	}
	var loadedPlugins []LoadedPlugin
	for _, clientPlugin := range clientPlugins {
		log.Printf("Loading plugin: %s", clientPlugin)
		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig:  shared.Handshake,
			Cmd:              exec.Command(clientPlugin),
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
			Logger:           logger,
		})
		//defer client.Kill()

		rpcClient, err := client.Client()
		if err != nil {
			log.Fatal("Error creating rpcClient: ", err)
		}

		// Request the plugin
		raw, err := rpcClient.Dispense("twitch")
		if err != nil {
			log.Fatal("Error on rpcClient.Dispense(): ", err)
		}
		newPlugin := raw.(shared.CustomPlugin)
		loadedPlugins = append(loadedPlugins, LoadedPlugin{client, &newPlugin})
		loadedPlugins[len(loadedPlugins)-1].Plugin.Impl.Start()
	}

	go func() {
		for {
			select {
			case <-interrupt:
				for range interrupt {
					logrus.Warnln("Interrupt detected, stopping plugins")
					for _, loadedPlugin := range loadedPlugins {

						loadedPlugin.Plugin.Impl.Stop()
						loadedPlugin.Client.Kill()
					}
					//ctxCancel()
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
