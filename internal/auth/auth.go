package auth

import "nxCacheService/internal/env"

type AllowedAction string

const (
	Read  AllowedAction = "READ"
	Write AllowedAction = "WRITE"
)

type StaticTokenAuthenticator struct {
}

func (StaticTokenAuthenticator) IsAllowed(token string, action AllowedAction) bool {
	if action == Read {
		return env.CiToken.GetValue() == token || env.ReadOnlyToken.GetValue() == token
	}

	if action == Write {
		return env.CiToken.GetValue() == token
	}

	return false
}
