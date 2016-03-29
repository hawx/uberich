package web

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/justinas/nosurf"
	"golang.org/x/crypto/bcrypt"
	"hawx.me/code/uberich/data"
)

const loginPage = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Login</title>
    <link rel="stylesheet" href="/styles.css" />
  </head>
  <body>
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
}

func Login(db data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			loginTmpl.Execute(w, loginCtx{
				Application: r.FormValue("application"),
				Token:       nosurf.Token(r),
				RedirectURI: r.FormValue("redirect_uri"),
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

			user, err := db.GetUser(email)
			if err != nil {
				log.Println("login:", email, err)
				return
			}
			if !user.Verified {
				log.Println("login: user not verified")
				return
			}

			app, err := db.GetApplication(application)
			if err != nil {
				log.Println("login:", err)
				return
			}

			if !app.CanRedirectTo(redirectURI.String()) {
				log.Println("login: cannot redirect to specified URI:", redirectURI.String())
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(pass)); err != nil {
				log.Println("login: password incorrect")
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
