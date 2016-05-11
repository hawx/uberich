package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
	"github.com/justinas/nosurf"
	"hawx.me/code/serve"
	"hawx.me/code/uberich/config"
	"hawx.me/code/uberich/cookies"
	"hawx.me/code/uberich/web"
)

func main() {
	var (
		settingsPath = flag.String("settings", "./settings.toml", "")
		port         = flag.String("port", "8080", "")
		socket       = flag.String("socket", "", "")
	)
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

	logger := log.New(os.Stdout, "", log.LstdFlags)

	http.Handle("/login", nosurf.New(web.Login(conf, store, logger)))
	http.Handle("/change-password", nosurf.New(web.ChangePassword(conf, store, logger)))
	http.Handle("/styles.css", web.Styles)

	serve.Serve(*port, *socket, context.ClearHandler(http.DefaultServeMux))
}
