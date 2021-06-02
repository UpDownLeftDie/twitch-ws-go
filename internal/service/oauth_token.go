package service

import (
	"time"

	"golang.org/x/oauth2"
)

type OauthToken struct {
	ClientID     string    `db:"client_id"`
	ServiceName  string    `db:"service_name"`
	AccessToken  string    `db:"access_token"`
	RefreshToken string    `db:"refresh_token"`
	ExpiresAt    time.Time `db:"expires_at"`
}

func (o OauthToken) Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
	}
}
