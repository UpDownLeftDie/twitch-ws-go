package twitch

import (
	"fmt"
	"net/http"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

func NewOAuthConfig(clientID, clientSecret string) *oauth2.Config {
	fmt.Println(viper.GetString("CALLBACK_URL"))
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  viper.GetString("CALLBACK_URL"),
		Scopes:       []string{viper.GetString("twitch.scopes")},
		Endpoint:     twitch.Endpoint,
	}
	return oauthConfig
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) client {
	return client{
		baseURL:    "https://id.twitch.tv",
		httpClient: httpClient,
	}
}
