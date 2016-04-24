package cookies

import (
	"net/http"

	"github.com/gorilla/securecookie"
)

type Store struct {
	store *securecookie.SecureCookie
}

func New(hashKey []byte, blockKey []byte) *Store {
	return &Store{
		store: securecookie.New(hashKey, blockKey),
	}
}

func (s *Store) Set(w http.ResponseWriter, email string) error {
	encoded, err := s.store.Encode("uberich", email)
	if err == nil {
		cookie := &http.Cookie{
			Name:     "uberich",
			Value:    encoded,
			Path:     "/",
			MaxAge:   3600 * 8, // 8 hours
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	}

	return err
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
