package service

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Repository interface {
	GetOauthTokens() ([]OauthToken, error)
	UpsertOauthToken(oauthToken OauthToken) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) repository {
	return repository{
		db: db,
	}
}

func (r repository) GetOauthTokens() ([]OauthToken, error) {
	var oauthTokens []OauthToken
	err := r.db.Select(
		&oauthTokens,
		`SELECT
			client_id, access_token, refresh_token, expires_at
		FROM
			oauth_token`,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return oauthTokens, nil
}

func (r repository) UpsertOauthToken(oauthToken OauthToken) error {
	_, err := r.db.Exec(
		`INSERT INTO
			oauth_token (client_id, access_token, refresh_token, expires_at)
		VALUES
			($1, $2, $3, $4)
		ON CONFLICT (client_id)
		DO UPDATE SET
			access_token = $2,
			refresh_token = $3,
			expires_at = $4,
			updated_at = CURRENT_TIMESTAMP`, // ! change this if using different db driver TODO: detect user config?
		oauthToken.ClientID,
		oauthToken.AccessToken,
		oauthToken.RefreshToken,
		oauthToken.ExpiresAt,
	)
	if err != nil {
		return errors.Wrap(err, "unable to upsert oauth token")
	}

	return nil
}
