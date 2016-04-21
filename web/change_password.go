package web

import (
	"net/http"

	"hawx.me/code/uberich/config"
)

func ChangePassword(conf *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
