package web

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/justinas/nosurf"
	"hawx.me/code/mux"
	"hawx.me/code/uberich/config"
	"hawx.me/code/uberich/cookies"
)

const loginPage = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Login</title>
    <link rel="stylesheet" href="/styles.css" />
  </head>
  <body>
    {{ if .WasProblem }}
      <p class="problem">Try again!</p>
    {{ end }}

    <form method="post" action="/login">
      <fieldset>
        <label for="email">Email</label>
        <input type="text" id="email" name="email" />
      </fieldset>

      <fieldset>
        <label for="pass">Password</label>
        <input type="password" id="pass" name="pass" />
      </fieldset>

      <input type="hidden" name="application" value="{{.Application}}" />
      <input type="hidden" name="csrf_token" value="{{.Token}}" />
      <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}" />

      <input type="submit" value="Login" />
    </form>
  </body>
</html>`

var loginTmpl = template.Must(template.New("login").Parse(loginPage))

type loginCtx struct {
	Application string
	Token       string
	RedirectURI string
	WasProblem  bool
}

func redirectWithParams(w http.ResponseWriter, r *http.Request, u *url.URL, params map[string]string) {
	q := u.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

type loginHandler struct {
	conf   *config.Config
	store  cookies.Store
	logger *log.Logger
}

func (h *loginHandler) getApp(w http.ResponseWriter, name, redirectURI string) *config.App {
	app := h.conf.GetApp(name)

	if app == nil {
		h.logger.Println("login: no such app", name)
		http.Error(w, "no such app", http.StatusInternalServerError)

	} else if !app.CanRedirectTo(redirectURI) {
		h.logger.Println("login: cannot redirect to specified URI:", redirectURI)
		http.Error(w, "no such app", http.StatusInternalServerError)
	}

	return app
}

func (h *loginHandler) Get(w http.ResponseWriter, r *http.Request) {
	var (
		application      = r.FormValue("application")
		redirectURI, err = url.Parse(r.FormValue("redirect_uri"))
		wasProblem       = r.FormValue("problem")
	)

	if err != nil {
		h.logger.Println("login: could not parse redirect_uri")
		http.Error(w, "could not parse redirect_uri", http.StatusInternalServerError)
		return
	}

	app := h.getApp(w, application, redirectURI.String())
	if app == nil {
		return
	}

	if email, err := h.store.Get(r); err == nil {
		redirectWithParams(w, r, redirectURI, map[string]string{
			"email":  email,
			"verify": string(app.HashWithSecret([]byte(email))),
		})

		return
	}

	loginTmpl.Execute(w, loginCtx{
		Application: application,
		Token:       nosurf.Token(r),
		RedirectURI: redirectURI.String(),
		WasProblem:  wasProblem != "",
	})
}

func (h *loginHandler) Post(w http.ResponseWriter, r *http.Request) {
	var (
		email            = r.PostFormValue("email")
		pass             = r.PostFormValue("pass")
		application      = r.PostFormValue("application")
		redirectURI, err = url.Parse(r.PostFormValue("redirect_uri"))
	)

	if err != nil {
		h.logger.Println("login: could not parse redirect_uri")
		http.Error(w, "could not parse redirect_uri", http.StatusInternalServerError)
		return
	}

	redirectHere := func() {
		redirectWithParams(w, r, r.URL, map[string]string{
			"application":  application,
			"redirect_uri": redirectURI.String(),
			"problem":      "yes",
		})
	}

	app := h.getApp(w, application, redirectURI.String())
	if app == nil {
		h.logger.Println("login: no such app", app)
		redirectHere()
		return
	}

	user := h.conf.GetUser(email)
	if user == nil {
		h.logger.Println("login: no such user", email)
		redirectHere()
		return
	}

	if !user.IsPassword(pass) {
		h.logger.Println("login: password incorrect", email)
		redirectHere()
		return
	}

	if err := h.store.Set(w, email); err != nil {
		h.logger.Println("login: could not set cookie:", err)
		redirectHere()
		return
	}

	redirectWithParams(w, r, r.URL, map[string]string{
		"application":  application,
		"redirect_uri": redirectURI.String(),
	})
}

// Login handles requests for a user to verify their identity. It displays and
// handles a standard login form.
func Login(conf *config.Config, store cookies.Store, logger *log.Logger) http.Handler {
	handler := &loginHandler{conf, store, logger}

	return mux.Method{
		"GET":  http.HandlerFunc(handler.Get),
		"POST": http.HandlerFunc(handler.Post),
	}
}
