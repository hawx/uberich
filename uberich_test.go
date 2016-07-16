package uberich

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"hawx.me/code/assert"
)

func TestSignOut(t *testing.T) {
	cookieSecret := "Cookie Secret"

	client := NewClient("", "", "", "", NewStore(cookieSecret))

	redirectCh := make(chan *http.Request, 1)
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCh <- r
	}))
	defer redirectServer.Close()

	signOut := httptest.NewServer(client.SignOut(redirectServer.URL))
	defer signOut.Close()

	jar, _ := cookiejar.New(&cookiejar.Options{})
	httpClient := http.Client{Jar: jar}

	req, _ := http.NewRequest("GET", signOut.URL, nil)
	resp, _ := httpClient.Do(req)

	assert := assert.New(t)

	assert.Equal(200, resp.StatusCode)

	select {
	case r := <-redirectCh:
		assert.Equal("", client.CurrentUser(r))

	case <-time.After(time.Second):
		t.Error("timeout")
	}
}

func TestSignIn(t *testing.T) {
	uberichCh := make(chan *http.Request, 1)
	uberich := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uberichCh <- r
	}))
	defer uberich.Close()

	cookieSecret := "Cookie Secret"
	appName := "my-app"
	appURI := "http://app_uri"
	somePath := "/cool/a/path"

	client := NewClient(appName, appURI, uberich.URL, "", NewStore(cookieSecret))

	signIn := httptest.NewServer(client.SignIn(""))
	defer signIn.Close()

	httpClient := http.Client{}
	req, _ := http.NewRequest("GET", signIn.URL+somePath, nil)
	resp, _ := httpClient.Do(req)

	assert := assert.New(t)

	assert.Equal(200, resp.StatusCode)

	select {
	case r := <-uberichCh:
		assert.Equal("/login", r.URL.Path)
		assert.Equal(appName, r.URL.Query().Get("application"))
		assert.Equal(appURI+somePath, r.URL.Query().Get("redirect_uri"))

	case <-time.After(time.Second):
		t.Error("timeout")
	}
}

func TestSignInWhenSignedIn(t *testing.T) {
	redirectCh := make(chan *http.Request, 1)
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCh <- r
	}))
	defer redirect.Close()

	cookieSecret := "Cookie Secret"
	appName := "my-app"
	appURI := "http://app_uri"
	secret := "rjiwjre my secret"
	email := "someguy@someplace.something"

	client := NewClient(appName, appURI, "", secret, NewStore(cookieSecret))

	signIn := httptest.NewServer(client.SignIn(redirect.URL))
	defer signIn.Close()

	jar, _ := cookiejar.New(&cookiejar.Options{})
	httpClient := http.Client{Jar: jar}

	query := url.Values{}
	query.Add("email", email)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(email))
	query.Add("verify", base64.URLEncoding.EncodeToString(mac.Sum(nil)))

	req, _ := http.NewRequest("GET", signIn.URL+"?"+query.Encode(), nil)
	resp, _ := httpClient.Do(req)

	assert := assert.New(t)

	assert.Equal(200, resp.StatusCode)

	select {
	case r := <-redirectCh:
		assert.Equal("/", r.URL.Path)

		user := client.CurrentUser(r)
		assert.Equal(email, user)

	case <-time.After(time.Second):
		t.Error("timeout")
	}
}
