package main

import (
	"encoding/base64"
	"flag"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"hawx.me/code/uberich/data"
	"hawx.me/code/uberich/web"
)

var (
	settingsPath = flag.String("settings", "./settings.toml", "")
	dbPath       = flag.String("database", "./uberich.db", "")
	port         = flag.String("port", "8080", "")
	socket       = flag.String("socket", "", "")
)

func main() {
	flag.Parse()

	var conf struct {
		Users []string `toml:"users"`

		Session struct {
			Auth string `toml:"auth"`
			Enc  string `toml:"enc"`
		} `toml:"session"`

		Mail struct {
			Addr string `toml:"addr"`
			User string `toml:"user"`
			Pass string `toml:"pass"`
			From string `toml:"from"`
		} `toml:"mail"`
	}

	if _, err := toml.DecodeFile(*settingsPath, &conf); err != nil {
		log.Println("toml: %v", err)
		return
	}

	var (
		authKey, _ = base64.StdEncoding.DecodeString(conf.Session.Auth)
		encKey, _  = base64.StdEncoding.DecodeString(conf.Session.Enc)
		mailer     = web.NewMailer(conf.Mail.User, conf.Mail.Pass, conf.Mail.Addr, conf.Mail.From)
		store      = sessions.NewCookieStore(authKey, encKey)
		db, _      = data.Open(*dbPath)
	)

	http.Handle("/register", nosurf.New(web.Register(conf.Users, db, mailer)))
	http.Handle("/confirm", web.Confirm(db))
	http.Handle("/login", nosurf.New(web.Login(db, store)))
	http.Handle("/styles.css", web.Styles)

	log.Println("Running on :8000")
	http.ListenAndServe(":8000", context.ClearHandler(http.DefaultServeMux))
}
