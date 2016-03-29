// Package flow implements the client flow of uberichs authentication protocol.
//
// It exposes two types:
// - a store which simply wraps a cookie jar; and
// - a client which provides handlers that implement sign-in/out and protection.
//
package flow

import (
	"crypto/hmac"
	"crypto/sha256"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
)

type Store interface {
	Set(w http.ResponseWriter, r *http.Request, email string)
	Get(r *http.Request) string
}

type emailStore struct {
	store sessions.Store
}

func NewStore(secret string) Store {
	return &emailStore{sessions.NewCookieStore([]byte(secret))}
}

func (s emailStore) Get(r *http.Request) string {
	session, _ := s.store.Get(r, "session")

	if v, ok := session.Values["email"].(string); ok {
		return v
	}

	return ""
}

func (s emailStore) Set(w http.ResponseWriter, r *http.Request, email string) {
	session, _ := s.store.Get(r, "session")
	session.Values["email"] = email
	session.Save(r, w)
}

func Client(application, uberichURL, secret string, store Store) *client {
	u, _ := url.Parse(uberichURL)

	return &client{
		application: application,
		uberichURL:  u,
		secret:      secret,
		store:       store,
	}
}

type client struct {
	application string
	uberichURL  *url.URL
	secret      string
	store       Store
}

func (c *client) wasHashedWithSecret(data []byte, verifyMAC []byte) bool {
	mac := hmac.New(sha256.New, []byte(c.secret))
	mac.Write(data)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(verifyMAC, expectedMAC)
}

// signInURI is the the path to this handler, redirectURI is the path to
// redirect on successful sign-ins.
func (c *client) SignIn(signInURI, redirectURI string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if email := r.FormValue("email"); email != "" {
			verifyMAC := r.FormValue("verify")
			if !c.wasHashedWithSecret([]byte(email), []byte(verifyMAC)) {
				log.Println("sign-in: response from unverified source")
				return
			}

			c.store.Set(w, r, email)
			http.Redirect(w, r, redirectURI, http.StatusFound)
			return
		}

		u, _ := c.uberichURL.Parse("login")
		q := u.Query()
		q.Add("redirect_uri", signInURI)
		q.Add("application", c.application)
		u.RawQuery = q.Encode()

		http.Redirect(w, r, u.String(), http.StatusFound)
	})
}

// redirectURI is the path to redirect to after successful sign-out.
func (c *client) SignOut(redirectURI string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.store.Set(w, r, "")
		http.Redirect(w, r, redirectURI, http.StatusFound)
	})
}

func (c *client) Protect(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if email := c.store.Get(r); email != "" {
			handler.ServeHTTP(w, r)
		}
	})
}
