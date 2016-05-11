package web

import (
	"html/template"
	"log"
	"net/http"

	"github.com/justinas/nosurf"

	"hawx.me/code/uberich/config"
	"hawx.me/code/uberich/cookies"
)

const changePasswordPage = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Change Password</title>
    <link rel="stylesheet" href="/styles.css">
  </head>
  <body>
    <form method="post" action="/change-password">
      <fieldset>
        <label for="pass">New Password</label>
        <input type="password" id="pass" name="pass" />
      </fieldset>

      <fieldset>
        <label for="pass2">(Confirm)</label>
        <input type="password" id="pass2" name="pass2" />
      </fieldset>

      <input type="hidden" name="csrf_token" value="{{.Token}}" />

      <input type="submit" value="Change" />
    </form>
  </body>
</html>`

var changePasswordTmpl = template.Must(template.New("changePassword").Parse(changePasswordPage))

type changePasswordCtx struct {
	Token string
}

func ChangePassword(conf *config.Config, store cookies.Store, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email, err := store.Get(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		switch r.Method {
		case "GET":
			changePasswordTmpl.Execute(w, changePasswordCtx{
				Token: nosurf.Token(r),
			})

		case "POST":
			var (
				pass    = r.PostFormValue("pass")
				confirm = r.PostFormValue("pass2") == pass
			)

			if !confirm {
				http.Redirect(w, r, r.URL.Path, http.StatusFound)
				return
			}

			user := conf.GetUser(email)
			if user == nil {
				// This is impossible, really. Does it need checking?
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			if err := user.SetPassword(pass); err != nil {
				logger.Println("change-password:", err)
				return
			}

			conf.SetUser(user)

			if err := conf.Save(); err != nil {
				logger.Println("change-password:", err)
				return
			}

			store.Unset(w)

		default:
			w.WriteHeader(405)
		}
	})
}
