package flow

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"

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
		sessionsStore := sessions.NewCookieStore([]byte(cookieSecret))
		session, _ := sessionsStore.Get(r, "session")
		value, ok := session.Values["email"].(string)

		assert.True(ok)
		assert.Equal("", value)

	case <-time.After(time.Second):
		t.Error("timeout")
	}
}
