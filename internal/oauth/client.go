package oauth

import (
	"net/http"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func NewOAuthConfig(clientID string, clientSecret string, scopes []string, endpoint oauth2.Endpoint) *oauth2.Config {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  viper.GetString("CALLBACK_URL"),
		Scopes:       scopes,
		Endpoint:     endpoint,
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
