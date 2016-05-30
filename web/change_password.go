package web

import (
	"html/template"
	"log"
	"net/http"

	"github.com/justinas/nosurf"

	"hawx.me/code/mux"
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

type changePasswordHandler struct {
	conf   *config.Config
	store  cookies.Store
	logger *log.Logger
}

func (h *changePasswordHandler) Get(w http.ResponseWriter, r *http.Request) {
	if _, err := h.store.Get(r); err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	changePasswordTmpl.Execute(w, changePasswordCtx{
		Token: nosurf.Token(r),
	})
}

func (h *changePasswordHandler) Post(w http.ResponseWriter, r *http.Request) {
	email, err := h.store.Get(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	var (
		pass    = r.PostFormValue("pass")
		confirm = r.PostFormValue("pass2") == pass
	)

	if !confirm {
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
		return
	}

	user := h.conf.GetUser(email)
	if user == nil {
		// This is impossible, really. Does it need checking?
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := user.SetPassword(pass); err != nil {
		h.logger.Println("change-password:", err)
		return
	}

	h.conf.SetUser(user)

	if err := h.conf.Save(); err != nil {
		h.logger.Println("change-password:", err)
		return
	}

	h.store.Unset(w)
}

func ChangePassword(conf *config.Config, store cookies.Store, logger *log.Logger) http.Handler {
	handler := &changePasswordHandler{conf, store, logger}

	return mux.Method{
		"GET":  http.HandlerFunc(handler.Get),
		"POST": http.HandlerFunc(handler.Post),
	}
}
