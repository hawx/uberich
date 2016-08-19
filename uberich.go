// Package uberich implements the client flow of uberich's authentication protocol.
package uberich

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

func NewClient(appName, appURL, uberichURL, secret string, store Store) *Client {
	appU, _ := url.Parse(appURL)
	uberichU, _ := url.Parse(uberichURL)

	return &Client{
		appName:    appName,
		appURL:     appU,
		uberichURL: uberichU,
		secret:     secret,
		store:      store,
	}
}

type Client struct {
	appName    string
	appURL     *url.URL
	uberichURL *url.URL
	secret     string
	store      Store
}

func (c *Client) wasHashedWithSecret(data []byte, verifyMAC []byte) bool {
	mac := hmac.New(sha256.New, []byte(c.secret))
	mac.Write(data)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(verifyMAC, expectedMAC)
}

// SignIn returns a handler that prompts the user to sign-in with uberich, on
// success they will be redirected to redirectURI.
func (c *Client) SignIn(redirectURI string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if email := r.FormValue("email"); email != "" {
			verifyMAC, err := base64.URLEncoding.DecodeString(r.FormValue("verify"))
			if err != nil {
				log.Println(err)
			}

			if !c.wasHashedWithSecret([]byte(email), verifyMAC) {
				log.Println("sign-in: response from unverified source")
				return
			}

			c.store.Set(w, r, email)
			http.Redirect(w, r, redirectURI, http.StatusFound)
			return
		}

		path := r.URL.Path
		if len(path) > 0 {
			path = path[1:]
		}

		redirectURI, _ := c.appURL.Parse(path)

		u, _ := c.uberichURL.Parse("login")
		q := u.Query()
		q.Add("redirect_uri", redirectURI.String())
		q.Add("application", c.appName)
		u.RawQuery = q.Encode()

		http.Redirect(w, r, u.String(), http.StatusFound)
	})
}

// SignOut returns a handler that removes the session cookie for the currently
// signed-in user. It then redirects to redirectURI.
func (c *Client) SignOut(redirectURI string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.store.Set(w, r, "")
		http.Redirect(w, r, redirectURI, http.StatusFound)
	})
}

// Protect takes two handlers, the first will be used if an entry exists in the
// store. Otherwise the second handler is used.
func (c *Client) Protect(handler, errHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if email := c.store.Get(r); email != "" {
			handler.ServeHTTP(w, r)
		} else {
			errHandler.ServeHTTP(w, r)
		}
	})
}

func (c *Client) CurrentUser(r *http.Request) string {
	return c.store.Get(r)
}
