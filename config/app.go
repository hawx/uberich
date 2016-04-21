package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"strings"
)

type App struct {
	Name   string `toml:"name"`
	URI    string `toml:"uri"`
	Secret string `toml:"secret"`
}

// CanRedirectTo checks whether the Application can issue a HTTP redirect to the
// given URI by checking if it shares the same root URI.
func (a App) CanRedirectTo(uri string) bool {
	return strings.HasPrefix(uri, a.URI)
}

// HashWithSecret returns a HMAC using the secret for the Application.
func (a App) HashWithSecret(data []byte) []byte {
	mac := hmac.New(sha256.New, []byte(a.Secret))
	mac.Write(data)
	return mac.Sum(nil)
}
