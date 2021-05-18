package services

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

var (
	oauthConfTwitch = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "",
		Scopes:       []string{},
		Endpoint:     twitch.Endpoint,
	}
	oauthStateStringTwitch = ""
)

/*
InitializeOAuthTwitch Function
*/
func InitializeOAuthTwitch() {
	oauthConfTwitch.ClientID = viper.GetString("twitch.clientID")
	oauthConfTwitch.ClientSecret = viper.GetString("twitch.clientSecret")
	oauthConfTwitch.Scopes = []string{viper.GetString("twitch.scopes")}
	oauthConfTwitch.RedirectURL = viper.GetString("redirectURL")
	oauthStateStringTwitch = viper.GetString("oauthStateString")
}

/*
HandleGoogleLogin Function
*/
func HandleTwitchLogin(w http.ResponseWriter, r *http.Request) {
	HandleLogin(w, r, oauthConfTwitch, oauthStateStringTwitch)
}

/*
CallBackFromTwitch Function
*/
func CallBackFromTwitch(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Callback-twitch..")

	state := r.FormValue("state")
	logger.Log.Info(state)
	if state != oauthStateStringTwitch {
		logger.Log.Info("invalid oauth state, expected " + oauthStateStringTwitch + ", got " + state + "\n")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	logger.Log.Info(code)

	if code == "" {
		logger.Log.Warn("Code not found..")
		w.Write([]byte("Code Not Found to provide AccessToken..\n"))
		reason := r.FormValue("error_reason")
		if reason == "user_denied" {
			w.Write([]byte("User has denied Permission.."))
		}
		// User has denied access..
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	} else {
		token, err := oauthConfTwitch.Exchange(oauth2.NoContext, code)
		if err != nil {
			logger.Log.Error("oauthConfTwitch.Exchange() failed with " + err.Error() + "\n")
			return
		}
		logger.Log.Info("TOKEN>> AccessToken>> " + token.AccessToken)
		logger.Log.Info("TOKEN>> Expiration Time>> " + token.Expiry.String())
		logger.Log.Info("TOKEN>> RefreshToken>> " + token.RefreshToken)

		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
		if err != nil {
			logger.Log.Error("Get: " + err.Error() + "\n")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Log.Error("ReadAll: " + err.Error() + "\n")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		logger.Log.Info("parseResponseBody: " + string(response) + "\n")

		w.Write([]byte("Hello, I'm protected\n"))
		w.Write([]byte(string(response)))
		return
	}
}
