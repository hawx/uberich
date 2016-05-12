package cookies

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

type Store interface {
	Set(w http.ResponseWriter, email string) error
	Unset(w http.ResponseWriter)
	Get(r *http.Request) (email string, err error)
}

type store struct {
	domain string
	secure bool
	cookie *securecookie.SecureCookie
}

func New(domain string, secure bool, hashKey, blockKey []byte) Store {
	return &store{
		domain: domain,
		secure: secure,
		cookie: securecookie.New(hashKey, blockKey),
	}
}

func (s *store) Set(w http.ResponseWriter, email string) error {
	encoded, err := s.cookie.Encode("uberich", email)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "uberich",
			Value:    encoded,
			Path:     "/",
			Domain:   s.domain,
			Expires:  time.Now().UTC().Add(8 * 60 * time.Minute),
			HttpOnly: true,
			Secure:   s.secure,
		})
	}

	return err
}

func (s *store) Unset(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    "uberich",
		Value:   "",
		Path:    "/",
		Domain:  s.domain,
		Expires: time.Now().UTC().Add(60 * time.Minute),
		Secure:  s.secure,
	})
}

func (s *store) Get(r *http.Request) (string, error) {
	cookie, err := r.Cookie("uberich")
	if err != nil {
		return "", err
	}

	var value string
	if err = s.cookie.Decode("uberich", cookie.Value, &value); err != nil {
		return "", err
	}

	if value == "" {
		return "", errors.New("invalid user")
	}

	return value, nil
}
