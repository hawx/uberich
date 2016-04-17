package web

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"hawx.me/code/uberich/data"
)

const confirmPage = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Confirm</title>
    <link rel="stylesheet" href="/styles.css" />
  </head>
  <body>
    <form method="post" action="/confirm">
      <fieldset>
        <label for="email">Username</label>
        <input type="text" id="email" name="email" />
      </fieldset>

      <fieldset>
        <label for="token">Token</label>
        <input type="text" id="token" name="token" />
      </fieldset>

      <input type="submit" value="Confirm" />
    </form>
  </body>
</html>`

var confirmTmpl = template.Must(template.New("register").Parse(registerPage))

// Confirm handles the final step of registering users. It is visited via a link
// sent to the user's email which contains a token. If the token matches that
// set on the user, verified is set to true.
func Confirm(db data.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			email = r.FormValue("email")
			token = r.FormValue("token")
		)

		if email == "" || token == "" {
			confirmTmpl.Execute(w, nil)
			return
		}

		record, err := db.GetUser(email)
		if err != nil {
			log.Println("confirm:", err)
			http.NotFound(w, r)
			return
		}

		if time.Now().After(record.Expires) {
			log.Println("confirm: expired for", email)
			http.NotFound(w, r)
			return
		}

		if token != record.Token {
			log.Println("confirm: incorrect token for", email)
			http.NotFound(w, r)
			return
		}

		record.Verified = true
		db.SetUser(record)
	})
}
