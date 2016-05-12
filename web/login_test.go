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

type fakeStore struct{ s string }

func (s *fakeStore) Set(_ http.ResponseWriter, email string) error {
	s.s = email
	return nil
}

func (s *fakeStore) Unset(_ http.ResponseWriter) {}

func (s *fakeStore) Get(_ *http.Request) (string, error) {
	if s.s == "" {
		return "", errors.New("")
	}
	return s.s, nil
}

func emptyStore() *fakeStore {
	return &fakeStore{""}
}

var discardLogger = log.New(ioutil.Discard, "", 0)

func conf(app *config.App) *config.Config {
	return &config.Config{Apps: []*config.App{app}}
}

func addUser(conf *config.Config, email, pass string) {
	user := &config.User{Email: email}
	user.SetPassword(pass)
	conf.SetUser(user)
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

func httpPost(reqURL string, params map[string]string) (*http.Response, error) {
	q := url.Values{}
	for k, v := range params {
		q.Add(k, v)
	}

	return http.PostForm(reqURL, q)
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

	loginServer := httptest.NewServer(Login(conf(testApp), &fakeStore{email}, discardLogger))
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

	loginServer := httptest.NewServer(Login(conf(testApp), emptyStore(), discardLogger))
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

	loginServer := httptest.NewServer(Login(conf(testApp), &fakeStore{email}, discardLogger))
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

func TestLoginWhenPost(t *testing.T) {
	success, successServer := chanServer()
	defer successServer.Close()

	email := "me@example.com"
	password := "hello"
	testApp := &config.App{
		Name:   "testing",
		URI:    successServer.URL,
		Secret: "i have secrets",
	}
	hmac := "n\xf3\x18D\xa7\r\xc6b\x87y\xd7\xca\a\x80T\xd9s\xb9`b\n\xa2ly\xd8\xedN\xddca\xb8<"

	conf := conf(testApp)
	addUser(conf, email, password)

	loginServer := httptest.NewServer(Login(conf, emptyStore(), discardLogger))
	defer loginServer.Close()

	resp, err := httpPost(loginServer.URL, map[string]string{
		"email":        email,
		"pass":         password,
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

func TestLoginWhenPostWithBadCredentials(t *testing.T) {
	testCases := []struct {
		Email, Pass             string
		AppendEmail, AppendPass string
	}{
		{"me@example.com", "secretpassword", "2", ""},
		{"me@example.com", "secretpassword", "", "2"},
	}

	success, successServer := chanServer()
	defer successServer.Close()

	testApp := &config.App{
		Name:   "testing",
		URI:    successServer.URL,
		Secret: "i have secrets",
	}

	for _, testCase := range testCases {
		email := testCase.Email
		password := testCase.Pass

		conf := conf(testApp)
		addUser(conf, email, password)

		loginServer := httptest.NewServer(Login(conf, emptyStore(), discardLogger))
		defer loginServer.Close()

		resp, err := httpPost(loginServer.URL, map[string]string{
			"email":        email + testCase.AppendEmail,
			"pass":         password + testCase.AppendPass,
			"application":  testApp.Name,
			"redirect_uri": testApp.URI,
		})

		assert := assert.New(t)
		assert.Nil(err)
		assert.Equal(resp.StatusCode, 200)

		select {
		case <-success:
			t.Error("was redirected", testCase)
		case <-time.After(time.Second):
		}
	}
}
