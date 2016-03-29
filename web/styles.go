package web

import (
	"fmt"
	"net/http"
)

const styles = `
body {
    margin: 3rem;
    font-family: monospace;
}

label {
    width: 5rem;
    text-align: right;
    margin-right: 1rem;
}

fieldset {
    display: block;
    border: none;
}

input[type=text], input[type=password] {
    font-family: monospace;
    border: none;
    border-bottom: 1px dotted;
    margin: 0 0 1rem;
    width: 15rem;
}

input[type=submit] {
    font-family: monospace;
    width: 5rem;
    margin-left: 15rem;
    border: 1px solid;
    background: none;
}
`

var Styles = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/css")
	fmt.Fprint(w, styles)
})
