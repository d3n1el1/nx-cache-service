package auth

import (
	"crypto/subtle"
	"errors"

	"nxCacheService/internal/env"
)

type AllowedAction string

const (
	Read  AllowedAction = "READ"
	Write AllowedAction = "WRITE"
)

var ErrMissingCiToken = errors.New("auth: CI_TOKEN must be set")

type StaticTokenAuthenticator struct {
	ciToken       string
	readOnlyToken string
}

func NewStaticTokenAuthenticator() (StaticTokenAuthenticator, error) {
	authenticator := StaticTokenAuthenticator{
		ciToken:       env.CiToken.GetValue(),
		readOnlyToken: env.ReadOnlyToken.GetValue(),
	}

	if len(authenticator.ciToken) == 0 {
		return StaticTokenAuthenticator{}, ErrMissingCiToken
	}

	return authenticator, nil
}

func (a StaticTokenAuthenticator) HasReadOnlyToken() bool {
	return len(a.readOnlyToken) > 0
}

func (a StaticTokenAuthenticator) IsKnown(token string) bool {
	return matches(a.ciToken, token) || matches(a.readOnlyToken, token)
}

func (a StaticTokenAuthenticator) IsAllowed(token string, action AllowedAction) bool {
	if action == Read {
		return matches(a.ciToken, token) || matches(a.readOnlyToken, token)
	}

	if action == Write {
		return matches(a.ciToken, token)
	}

	return false
}

func matches(configured, provided string) bool {
	if len(configured) == 0 || len(provided) == 0 {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(configured), []byte(provided)) == 1
}
