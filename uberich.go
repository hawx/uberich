package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/justinas/nosurf"
	"hawx.me/code/serve"
	"hawx.me/code/uberich/config"
	"hawx.me/code/uberich/cookies"
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
		log.Println("config:", err)
		return
	}

	hashKey, blockKey, err := conf.Keys()
	if err != nil {
		log.Println("config:", err)
	}
	store := cookies.New(conf.Domain, conf.Secure, hashKey, blockKey)

	http.Handle("/login", nosurf.New(web.Login(conf, store)))
	http.Handle("/change-password", nosurf.New(web.ChangePassword(conf, store)))
	http.Handle("/styles.css", web.Styles)

	serve.Serve(*port, *socket, context.ClearHandler(http.DefaultServeMux))
}
