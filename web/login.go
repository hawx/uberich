package web

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
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
        <label for="user">Username</label>
        <input type="text" id="user" name="user" />
      </fieldset>

      <fieldset>
        <label for="pass">Password</label>
        <input type="password" id="pass" name="pass" />
      </fieldset>

      <input type="hidden" name="csrf_token" value="{{.Token}}" />
      <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}" />

      <input type="submit" value="Login" />
    </form>
  </body>
</html>`

var loginTmpl = template.Must(template.New("login").Parse(loginPage))

type loginCtx struct {
	Token       string
	RedirectURI string
}

func Login(db data.Database, store *sessions.CookieStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			loginTmpl.Execute(w, loginCtx{
				Token:       nosurf.Token(r),
				RedirectURI: r.FormValue("redirect_uri"),
			})
			return
		}

		if r.Method == "POST" {
			user := r.PostFormValue("user")
			pass := r.PostFormValue("pass")

			record, ok := db.Get(user)
			if !ok {
				log.Println("login: no such user")
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(record.Hash), []byte(pass)); err != nil {
				log.Println("login: password incorrect")
				return
			}

			session, _ := store.Get(r, "user")
			session.Values["email"] = user
			session.Save(r, w)

			redirectURI, _ := url.Parse(r.PostFormValue("redirect_uri"))
			query := redirectURI.Query()
			query.Add("email", user)
			redirectURI.RawQuery = query.Encode()

			http.Redirect(w, r, redirectURI.String(), http.StatusFound)
			return
		}

		w.WriteHeader(405)
	})
}
