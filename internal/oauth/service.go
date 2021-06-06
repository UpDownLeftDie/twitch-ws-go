package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
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

	oauthToken := Token{
		ClientID:     s.oauthConfig.ClientID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	expiresInFloat := token.Extra("expires_in").(float64)
	if expiresInFloat < 0 {
		logrus.Error("invalid oauth code exchange, no expires_in")
		return errors.New("missing expires_in")
	}
	expiresAt := time.Unix(time.Now().Unix()+int64(expiresInFloat), 0)

	oauthToken.ExpiresAt = expiresAt

	err = s.repo.UpsertOauthToken(oauthToken)
	if err != nil {
		return err
	}

	return nil
}
