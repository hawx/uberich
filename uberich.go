package main

import (
	"encoding/base64"
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
		log.Println("toml:", err)
		return
	}

	// Hash keys should be at least 32 bytes long
	hashKey, _ := base64.StdEncoding.DecodeString("dCTDOQl23gR+f4kLIPHTo6rqo7QmzVyhvJ9+nIMv3jo=")
	blockKey, _ := base64.StdEncoding.DecodeString("dCTDOQl23gR+f4kLIPHTo6rqo7QmzVyhvJ9+nIMv3jo=")
	store := cookies.New(hashKey, blockKey)

	http.Handle("/login", nosurf.New(web.Login(conf, store)))
	http.Handle("/change-password", nosurf.New(web.ChangePassword(conf, store)))
	http.Handle("/styles.css", web.Styles)

	serve.Serve(*port, *socket, context.ClearHandler(http.DefaultServeMux))
}
