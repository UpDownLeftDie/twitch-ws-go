package oauth

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func MakeHTTPHandler(service Service) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := service.AuthCodeURL("state-csrf-todo")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		err := service.AuthorizeCallback(r.FormValue("state"), r.FormValue("code"))
		if err != nil {
			logrus.WithError(err).Error("failed to process authorize callback")
			fmt.Fprintf(w, "Failed!")
			return
		}

		fmt.Fprintf(w, "Success!")
	})

	return mux
}
