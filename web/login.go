package web

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/justinas/nosurf"
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

// Login handles requests for a user to verify their identity. It displays and
// handles a standard login form.
func Login(conf *config.Config, store *cookies.Store) http.Handler {
	getApp := func(w http.ResponseWriter, r *http.Request, name, redirectURI string) *config.App {
		app := conf.GetApp(name)

		if app == nil {
			log.Println("login: no such app", name)
			http.Error(w, "no such app", http.StatusInternalServerError)

			return app
		}

		if !app.CanRedirectTo(redirectURI) {
			log.Println("login: cannot redirect to specified URI:", redirectURI)
			http.Error(w, "no such app", http.StatusInternalServerError)

			return app
		}

		return app
	}

	redirectSuccess := func(w http.ResponseWriter, r *http.Request, app *config.App, email string, redirectURI *url.URL) {
		query := redirectURI.Query()
		query.Add("email", email)
		query.Add("verify", string(app.HashWithSecret([]byte(email))))
		redirectURI.RawQuery = query.Encode()

		http.Redirect(w, r, redirectURI.String(), http.StatusFound)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			var (
				application    = r.FormValue("application")
				redirectURI, _ = url.Parse(r.FormValue("redirect_uri"))
				wasProblem     = r.FormValue("problem")
			)

			app := getApp(w, r, application, redirectURI.String())
			if app == nil {
				return
			}

			if email, err := store.Get(r); err == nil {
				redirectSuccess(w, r, app, email, redirectURI)
				return
			}

			loginTmpl.Execute(w, loginCtx{
				Application: application,
				Token:       nosurf.Token(r),
				RedirectURI: redirectURI.String(),
				WasProblem:  wasProblem != "",
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

			redirectHere := func(ok bool) {
				here := r.URL
				query := url.Values{}
				query.Add("application", application)
				query.Add("redirect_uri", redirectURI.String())
				if !ok {
					query.Add("problem", "yes")
				}
				here.RawQuery = query.Encode()

				http.Redirect(w, r, here.String(), http.StatusFound)
			}

			user := conf.GetUser(email)
			if user == nil {
				log.Println("login: no such user", email)
				redirectHere(false)
				return
			}

			app := getApp(w, r, application, redirectURI.String())
			if app == nil {
				return
			}

			if !user.IsPassword(pass) {
				log.Println("login: password incorrect", email)
				redirectHere(false)
				return
			}

			if err := store.Set(w, email); err != nil {
				log.Println("login: could not set cookie:", err)
				redirectHere(false)
				return
			}

			redirectHere(true)
			return
		}

		w.WriteHeader(405)
	})
}
