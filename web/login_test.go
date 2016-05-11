package web

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"hawx.me/code/assert"
	"hawx.me/code/uberich/config"
)

type fakeStore string

func (s fakeStore) Set(_ http.ResponseWriter, email string) error { return nil }
func (s fakeStore) Unset(_ http.ResponseWriter)                   {}
func (s fakeStore) Get(_ *http.Request) (string, error)           { return string(s), nil }

type emptyStore struct{ *fakeStore }

func (s emptyStore) Get(_ *http.Request) (string, error) { return "", errors.New("") }

var discardLogger = log.New(ioutil.Discard, "", 0)

func conf(app *config.App) *config.Config {
	return &config.Config{Apps: []*config.App{app}}
}

func httpGet(reqURL string, params map[string]string) (*http.Response, error) {
	u, _ := url.Parse(reqURL)
	q := u.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	return http.Get(u.String())
}

func chanServer() (<-chan *http.Request, *httptest.Server) {
	ch := make(chan *http.Request, 1)
	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ch <- r
		}),
	)

	return ch, s
}

func TestLogin(t *testing.T) {
	success, successServer := chanServer()
	defer successServer.Close()

	email := "me@example.com"
	testApp := &config.App{
		Name:   "testing",
		URI:    successServer.URL,
		Secret: "i have secrets",
	}
	hmac := "n\xf3\x18D\xa7\r\xc6b\x87y\xd7\xca\a\x80T\xd9s\xb9`b\n\xa2ly\xd8\xedN\xddca\xb8<"

	loginServer := httptest.NewServer(Login(conf(testApp), fakeStore(email), discardLogger))
	defer loginServer.Close()

	resp, err := httpGet(loginServer.URL, map[string]string{
		"application":  testApp.Name,
		"redirect_uri": testApp.URI,
	})

	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)

	select {
	case r := <-success:
		assert.Equal("GET", r.Method)
		assert.Equal("/", r.URL.Path)
		assert.Equal(email, r.URL.Query().Get("email"))
		assert.Equal(hmac, r.URL.Query().Get("verify"))

	case <-time.After(time.Second):
		t.Error("time out")
	}
}

func TestLoginWhenNoCookie(t *testing.T) {
	success, successServer := chanServer()
	defer successServer.Close()

	testApp := &config.App{
		Name:   "testing",
		URI:    successServer.URL,
		Secret: "i have secrets",
	}

	loginServer := httptest.NewServer(Login(conf(testApp), emptyStore{}, discardLogger))
	defer loginServer.Close()

	resp, err := httpGet(loginServer.URL, map[string]string{
		"application":  testApp.Name,
		"redirect_uri": testApp.URI,
	})

	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)

	select {
	case <-success:
		t.Error("was redirected")
	case <-time.After(time.Second):
	}
}

func TestLoginWhenNoSuchApp(t *testing.T) {
	email := "me@example.com"
	testApp := &config.App{
		Name:   "testing",
		URI:    "http://localhost",
		Secret: "i have secrets",
	}

	loginServer := httptest.NewServer(Login(conf(testApp), fakeStore(email), discardLogger))
	defer loginServer.Close()

	resp, err := httpGet(loginServer.URL, map[string]string{
		"application":  testApp.Name + "2",
		"redirect_uri": testApp.URI,
	})

	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 500)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("no such app\n", string(body))
}
