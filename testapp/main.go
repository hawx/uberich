package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/context"
	uberich "hawx.me/code/uberich/flow"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
  <head>
    <title>Index</title>
  </head>
  <body>
    <a href="/sign-in">Sign-in</a>
    <a href="/sign-out">Sign-out</a>
    <a href="/secret">SECRETS</a>
  </body>
</html>`)
}

func Secret(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
  <head>
    <title>Secrets</title>
  </head>
  <body>
    <h1>You're in!</h1>
  </body>
</html>`)
}

func main() {
	store := uberich.NewStore("COOKIE-SECRETSZ")

	uberich := uberich.Client("testapp", "http://localhost:8080", "thisissecret", store)

	http.HandleFunc("/", Index)
	http.Handle("/secret", uberich.Protect(http.HandlerFunc(Secret)))
	http.Handle("/sign-in", uberich.SignIn("http://localhost:3001/sign-in", "/secret"))
	http.Handle("/sign-out", uberich.SignOut("/"))

	http.ListenAndServe(":3001", context.ClearHandler(http.DefaultServeMux))
}
