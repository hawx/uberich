package web

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
	"github.com/justinas/nosurf"
	"hawx.me/code/uberich/config"
	"hawx.me/code/uberich/cookies"
)

func New(conf *config.Config) (http.Handler, error) {
	mux := http.NewServeMux()

	hashKey, blockKey, err := conf.Keys()
	if err != nil {
		return mux, err
	}

	store := cookies.New(conf.Domain, conf.Secure, hashKey, blockKey)

	logger := log.New(os.Stdout, "", log.LstdFlags)

	mux.Handle("/login", nosurf.New(Login(conf, store, logger)))
	mux.Handle("/change-password", nosurf.New(ChangePassword(conf, store, logger)))
	mux.Handle("/styles.css", Styles)

	return context.ClearHandler(mux), nil
}
