package oauth

import (
	"database/sql"
	"golang.org/x/oauth2"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Repository interface {
	GetOauthToken() (Token, error)
	UpsertOauthToken(oauthToken Token) error
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

func (r repository) UpsertOauthToken(oauthToken Token) error {
	_, err := r.db.Exec(
		`INSERT INTO
			oauth_token (client_id, access_token, refresh_token, scope, token_type, expires_at)
		VALUES
			($1, $2, $3, $4, $5, $6)
		ON CONFLICT (client_id)
		DO UPDATE SET
			access_token = $2,
			refresh_token = $3,
		    scope = $4,
		    token_type = $5,
			expires_at = $6,
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
