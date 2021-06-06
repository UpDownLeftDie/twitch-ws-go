package oauth

import (
	"time"

	"golang.org/x/oauth2"
)

type Token struct {
	ClientID     string    `db:"client_id"`
	ServiceName  string    `db:"service_name"`
	AccessToken  string    `db:"access_token"`
	RefreshToken string    `db:"refresh_token"`
	ExpiresAt    time.Time `db:"expires_at"`
}

func (o Token) Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
	}
}
