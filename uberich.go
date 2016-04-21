package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/justinas/nosurf"
	"hawx.me/code/serve"
	"hawx.me/code/uberich/config"
	"hawx.me/code/uberich/web"
)

var (
	settingsPath = flag.String("settings", "./settings.toml", "")
	port         = flag.String("port", "8080", "")
	socket       = flag.String("socket", "", "")
)

func main() {
	flag.Parse()

	conf, err := config.Read(*settingsPath)
	if err != nil {
		log.Println("toml:", err)
		return
	}

	http.Handle("/login", nosurf.New(web.Login(conf)))
	http.Handle("/change-password", nosurf.New(web.ChangePassword(conf)))
	http.Handle("/styles.css", web.Styles)

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
