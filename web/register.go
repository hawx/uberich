package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/justinas/nosurf"
	"github.com/pborman/uuid"
	"hawx.me/code/uberich/data"
)

const registerPage = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Register</title>
    <link rel="stylesheet" href="/styles.css" />
  </head>
  <body>
    <form method="post" action="/register">
      <fieldset>
        <label for="user">Username</label>
        <input type="text" id="user" name="user" />
      </fieldset>

      <fieldset>
        <label for="pass">Password</label>
        <input type="password" id="pass" name="pass" />
      </fieldset>

      <fieldset>
        <label for="pass2">(Confirm)</label>
        <input type="password" id="pass2" name="pass2" />
      </fieldset>

      <input type="hidden" name="csrf_token" value="{{.Token}}" />

      <input type="submit" value="Register" />
    </form>
  </body>
</html>`

var registerTmpl = template.Must(template.New("register").Parse(registerPage))

type registerCtx struct {
	Token string
}

// Register handles the initial step of adding new users. If the user is allowed
// to register they are sent an email containing a token that is checked by
// Confirm.
func Register(whitelist []string, db data.Database, mailer *Mailer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			registerTmpl.Execute(w, registerCtx{
				Token: nosurf.Token(r),
			})
			return
		}

		if r.Method == "POST" {
			user := r.PostFormValue("user")
			pass := r.PostFormValue("pass")

			if pass != r.PostFormValue("pass2") {
				// Flash password did not match
				log.Println("Passwords do not match")
				http.Redirect(w, r, "/register", http.StatusFound)
				return
			}

			whitelisted := false
			for _, name := range whitelist {
				if name == user {
					whitelisted = true
				}
			}

			if !whitelisted {
				// Flash user not allowed
				log.Println("register: not whitelisted", user)
				return
			}

			if _, err := db.GetUser(user); err != nil {
				// Flash already registered
				log.Println("register:", err)
				return
			}

			var (
				hash, _ = bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
				token   = uuid.NewRandom().String()
			)

			db.SetUser(data.User{
				Email:    user,
				Hash:     string(hash),
				Token:    token,
				Expires:  time.Now().Add(5 * time.Minute),
				Verified: false,
			})

			url, _ := url.Parse("http://localhost:8000/confirm")
			query := url.Query()
			query.Add("email", user)
			query.Add("token", token)
			url.RawQuery = query.Encode()

			err := mailer.Send(
				user,
				"Confirm registration",
				fmt.Sprintf("To verify your registration please go to %s", url.String()),
			)
			if err != nil {
				log.Println("register:", err)
			}

			http.Redirect(w, r, "http://localhost:8000/confirm", http.StatusFound)
			return
		}

		w.WriteHeader(405)
	})
}
