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

func redirectWithParams(w http.ResponseWriter, r *http.Request, u *url.URL, params map[string]string) {
	q := u.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

// Login handles requests for a user to verify their identity. It displays and
// handles a standard login form.
func Login(conf *config.Config, store cookies.Store, logger *log.Logger) http.Handler {
	getApp := func(w http.ResponseWriter, name, redirectURI string) *config.App {
		app := conf.GetApp(name)

		if app == nil {
			logger.Println("login: no such app", name)
			http.Error(w, "no such app", http.StatusInternalServerError)

		} else if !app.CanRedirectTo(redirectURI) {
			logger.Println("login: cannot redirect to specified URI:", redirectURI)
			http.Error(w, "no such app", http.StatusInternalServerError)
		}

		return app
	}

	redirectSuccess := func(w http.ResponseWriter, r *http.Request, app *config.App, email string, redirectURI *url.URL) {
		redirectWithParams(w, r, redirectURI, map[string]string{
			"email":  email,
			"verify": string(app.HashWithSecret([]byte(email))),
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			var (
				application      = r.FormValue("application")
				redirectURI, err = url.Parse(r.FormValue("redirect_uri"))
				wasProblem       = r.FormValue("problem")
			)

			if err != nil {
				logger.Println("login: could not parse redirect_uri")
				http.Error(w, "could not parse redirect_uri", http.StatusInternalServerError)
				return
			}

			app := getApp(w, application, redirectURI.String())
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

		case "POST":
			var (
				email          = r.PostFormValue("email")
				pass           = r.PostFormValue("pass")
				application    = r.PostFormValue("application")
				redirectURI, _ = url.Parse(r.PostFormValue("redirect_uri"))
			)

			redirectHere := func(ok bool) {
				params := map[string]string{
					"application":  application,
					"redirect_uri": redirectURI.String(),
				}
				if !ok {
					params["problem"] = "yes"
				}

				redirectWithParams(w, r, r.URL, params)
			}

			app := getApp(w, application, redirectURI.String())
			if app == nil {
				logger.Println("login: no such app", app)
				redirectHere(false)
				return
			}

			user := conf.GetUser(email)
			if user == nil {
				logger.Println("login: no such user", email)
				redirectHere(false)
				return
			}

			if !user.IsPassword(pass) {
				logger.Println("login: password incorrect", email)
				redirectHere(false)
				return
			}

			if err := store.Set(w, email); err != nil {
				logger.Println("login: could not set cookie:", err)
				redirectHere(false)
				return
			}

			redirectHere(true)

		default:
			w.WriteHeader(405)
		}
	})
}
