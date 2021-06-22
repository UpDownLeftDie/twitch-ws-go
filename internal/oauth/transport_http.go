package oauth

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func MakeHTTPHandler(service Service, responseCode *int) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO use real csrf token
		url := service.AuthCodeURL("state-csrf-todo")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		err := service.AuthorizeCallback(r.FormValue("state"), r.FormValue("code"))
		if err != nil {
			logrus.WithError(err).Error("failed to process authorize callback")
			*responseCode = http.StatusInternalServerError
			fmt.Fprintf(w, "Failed! Try again.")
			return
		}

		*responseCode = http.StatusOK
		fmt.Fprintf(w, "Success! You can close this window.")
	})

	return mux
}
