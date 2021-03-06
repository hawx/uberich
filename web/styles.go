package web

import (
	"fmt"
	"net/http"
)

const styles = `
body {
    margin: 3rem;
    font: 16px/1.3 monospace;
    width: 27rem;
}

hr {
    border: none;
    border-top: 1px solid #eee;
    width: 15rem;
    margin: 4rem auto;
}

h1 {
    font-size: 1em;
    font-weight: bold;
    text-decoration: underline;
}

ul {
    list-style: circle;
}

label {
    display: inline-block;
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
    margin-left: 7rem;
    border: 1px solid;
    background: none;
}

.problem {
    color: rgb(164, 34, 34);
    border: 1px dashed;
    padding: 1rem;
    margin: 0 0 2rem;
    display: inline-block;
}

@media screen and (max-width: 27rem) {
    body {
        margin: 3rem 1rem;
    }

    label {
        display: block;
        text-align: left;
        margin: .5rem 0;
    }
}
`

var Styles = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/css")
	fmt.Fprint(w, styles)
})
