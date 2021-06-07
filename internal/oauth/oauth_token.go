package oauth

import (
	"time"

	"golang.org/x/oauth2"
)

type Token struct {
	ClientID     string    `db:"client_id"`
	AccessToken  string    `db:"access_token"`
	RefreshToken string    `db:"refresh_token"`
	Scope        string    `db:"scope"`
	TokenType    string    `db:"token_type"`
	ExpiresAt    time.Time `db:"_"`
	ExpiresAtRaw string    `db:"expires_at"`
}

func (o Token) Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
	}
}
