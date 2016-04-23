package web

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/justinas/nosurf"
	"hawx.me/code/uberich/config"
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

// Login handles requests for a user to verify their identity. It displays and
// handles a standard login form.
func Login(conf *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			loginTmpl.Execute(w, loginCtx{
				Application: r.FormValue("application"),
				Token:       nosurf.Token(r),
				RedirectURI: r.FormValue("redirect_uri"),
				WasProblem:  r.FormValue("problem") != "",
			})
			return
		}

		if r.Method == "POST" {
			var (
				email          = r.PostFormValue("email")
				pass           = r.PostFormValue("pass")
				application    = r.PostFormValue("application")
				redirectURI, _ = url.Parse(r.PostFormValue("redirect_uri"))
			)

			redirectHere := func() {
				here := r.URL
				query := url.Values{}
				query.Add("application", application)
				query.Add("redirect_uri", redirectURI.String())
				query.Add("problem", "yes")
				here.RawQuery = query.Encode()

				http.Redirect(w, r, here.String(), http.StatusFound)
			}

			user := conf.GetUser(email)
			if user == nil {
				log.Println("login: no such user", email)
				redirectHere()
				return
			}

			app := conf.GetApp(application)
			if app == nil {
				log.Println("login: no such app", application)
				http.Error(w, "no such app", http.StatusInternalServerError)
				return
			}

			if !app.CanRedirectTo(redirectURI.String()) {
				log.Println("login: cannot redirect to specified URI:", redirectURI.String())
				http.Error(w, "cannot redirect to URI", http.StatusInternalServerError)
				return
			}

			if !user.IsPassword(pass) {
				log.Println("login: password incorrect", email)
				redirectHere()
				return
			}

			query := redirectURI.Query()
			query.Add("email", email)
			query.Add("verify", string(app.HashWithSecret([]byte(email))))
			redirectURI.RawQuery = query.Encode()

			http.Redirect(w, r, redirectURI.String(), http.StatusFound)
			return
		}

		w.WriteHeader(405)
	})
}
