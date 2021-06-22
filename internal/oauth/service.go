package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

type Service interface {
	AuthCodeURL(csrfToken string) string
	AuthorizeCallback(csrfToken, code string) error
}

type service struct {
	oauthConfig    *oauth2.Config
	connectBaseURL string
	repo           Repository
}

func NewService(oauthConfig *oauth2.Config, connectBaseURL string, repo Repository) service {
	return service{
		oauthConfig:    oauthConfig,
		connectBaseURL: connectBaseURL,
		repo:           repo,
	}
}

func (s service) AuthCodeURL(csrfToken string) string {
	return s.oauthConfig.AuthCodeURL(csrfToken)
}

func (s service) AuthorizeCallback(csrfToken, code string) error {

	token, err := s.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return err
	}

	err = s.repo.UpsertOauthToken(token, s.oauthConfig.ClientID)
	if err != nil {
		return err
	}

	return nil
}
