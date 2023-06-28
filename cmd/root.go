package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/updownleftdie/twitch-ws-go/v3/shared"
)

var info = shared.LogrusInfo{
	Version:    "None",
	Buildstamp: "None",
	Githash:    "None",
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "adapter",
	Short: "adapter",
	Long:  `An interface between APIs & Websockets`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	shared.SetupLogrus(info)

	return nil
}

//func convertTwitchScopesToTopics(scopes []string, twitchId string) []string {
//	var topics []string
//	var twitchTopics = map[string][]string{
//		"bits:read":                  {"channel-bits-events-v2.%s", "channel-bits-badge-unlocks.%s"},
//		"channel:read:redemptions":   {"channel-points-channel-v1.%s"},
//		"channel:read:subscriptions": {"channel-subscribe-events-v1.%s"},
//		"channel:moderate":           {"chat_moderator_actions.%s.%s", "automod-queue.%s.%s"},
//		"chat:read":                  {"user-moderation-notifications.%s.%s"},
//		"whispers:read":              {"whispers.<user ID>"},
//	}
//
//	for i := range scopes {
//		for j := range twitchTopics[scopes[i]] {
//			topic := fmt.Sprintf(twitchTopics[scopes[i]][j], twitchId)
//			topics = append(topics, topic)
//		}
//	}
//	return topics
//}
