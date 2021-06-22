package oauth

import (
	"database/sql"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"golang.org/x/oauth2"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Repository interface {
	GetOauthToken() (Token, error)
	UpsertOauthToken(token *oauth2.Token, clientID string) error
}

type repository struct {
	db          *sqlx.DB
	oauthConfig *oauth2.Config
}

func NewRepository(db *sqlx.DB, oauthConfig *oauth2.Config) repository {
	return repository{
		db:          db,
		oauthConfig: oauthConfig,
	}
}

func (r repository) GetOauthToken() (Token, error) {
	var oauthToken Token
	err := r.db.Get(
		&oauthToken,
		`SELECT
			client_id, access_token, refresh_token, scope, token_type, expires_at
		FROM
			oauth_token
		WHERE
			client_id = $1
		LIMIT 1`,
		r.oauthConfig.ClientID,
	)
	if err == sql.ErrNoRows {
		return Token{}, nil
	}
	if err != nil {
		return Token{}, err
	}
	oauthToken.ExpiresAt, err = time.Parse(time.RFC3339, oauthToken.ExpiresAtRaw)
	if err != nil {
		return oauthToken, err
	}
	return oauthToken, nil
}

func (r repository) UpsertOauthToken(token *oauth2.Token, clientID string) error {
	oauthToken := Token{
		ClientID:     clientID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
	}

	expiresInFloat := token.Extra("expires_in").(float64)
	if expiresInFloat < 0 {
		logrus.Error("invalid oauth code exchange, no expires_in")
		return errors.New("missing expires_in")
	}
	expiresAt := time.Unix(time.Now().Unix()+int64(expiresInFloat), 0)
	oauthToken.ExpiresAt = expiresAt

	scopesRaw := token.Extra("scope").([]interface{})
	var scopes []string
	for _, scope := range scopesRaw {
		scopes = append(scopes, scope.(string))
	}
	oauthToken.Scope = "[\"" + strings.Join(scopes, "\", \"") + "\"]"

	_, err := r.db.Exec(
		`INSERT INTO
			oauth_token (client_id, access_token, refresh_token, scope, token_type, expires_at)
		VALUES
			($1, $2, $3, $4, $5, $6)
		ON CONFLICT (client_id)
		DO UPDATE SET
			access_token = coalesce($2, access_token),
			refresh_token = coalesce($3, refresh_token),
			scope = coalesce($4, scope),
			token_type = coalesce($5, token_type),
			expires_at = coalesce($6, expires_at),
			updated_at = CURRENT_TIMESTAMP`, // ! change this if using different db driver TODO: detect user config?
		oauthToken.ClientID,
		oauthToken.AccessToken,
		oauthToken.RefreshToken,
		oauthToken.Scope,
		oauthToken.TokenType,
		oauthToken.ExpiresAt.Format(time.RFC3339),
	)
	if err != nil {
		return errors.Wrap(err, "unable to upsert oauth token")
	}

	return nil
}
