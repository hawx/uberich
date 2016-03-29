package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/justinas/nosurf"
	"hawx.me/code/serve"
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
		mailer = web.NewMailer(conf.Mail.User, conf.Mail.Pass, conf.Mail.Addr, conf.Mail.From)
		db, _  = data.Open(*dbPath)
	)

	defer db.Close()

	http.Handle("/register", nosurf.New(web.Register(conf.Users, db, mailer)))
	http.Handle("/confirm", web.Confirm(db))
	http.Handle("/login", nosurf.New(web.Login(db)))
	http.Handle("/styles.css", web.Styles)

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
