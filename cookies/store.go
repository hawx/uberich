package cookies

import (
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

type Store struct {
	domain string
	secure bool
	store  *securecookie.SecureCookie
}

func New(domain string, secure bool, hashKey, blockKey []byte) *Store {
	return &Store{
		domain: domain,
		secure: secure,
		store:  securecookie.New(hashKey, blockKey),
	}
}

func (s *Store) Set(w http.ResponseWriter, email string) error {
	encoded, err := s.store.Encode("uberich", email)
	if err == nil {
		cookie := &http.Cookie{
			Name:     "uberich",
			Value:    encoded,
			Path:     "/",
			Domain:   s.domain,
			Expires:  time.Now().UTC().Add(8 * 60 * time.Minute),
			HttpOnly: true,
			Secure:   s.secure,
		}
		http.SetCookie(w, cookie)
	}

	return err
}

func (s *Store) Unset(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    "uberich",
		Value:   "",
		Path:    "/",
		Domain:  s.domain,
		Expires: time.Now().UTC().Add(60 * time.Minute),
		Secure:  s.secure,
	})
}

func (s *Store) Get(r *http.Request) (string, error) {
	cookie, err := r.Cookie("uberich")
	if err != nil {
		return "", err
	}

	var value string
	if err = s.store.Decode("uberich", cookie.Value, &value); err != nil {
		return "", err
	}

	return value, nil
}
