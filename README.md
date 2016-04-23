# uberich

A small-scale 3rd party authentication server.

Designed for registering and authenticating a whitelisted set of
users. Authenticating in the scope of this application means saying "yes that is
an email address I know about and this is the person it belongs to". It is
clearly inspired by Persona in this regard, but does not aim to be a replacement
for Persona.

I am not using this at the moment, and neither should you.


## Set-up

```bash
$ go get hawx.me/code/uberich/...
```

Use `uberich-admin` to add users and apps.

```bash
$ uberich-admin set-user someone@example.com secretPassword
$ uberich-admin set-app testApp http://test.example.com sharedSecret
$ uberich
...
```

Now `testApp` can integrate with uberich using the `flow` package.

```go
import (
  "github.com/gorilla/context"
  uberich "hawx.me/code/uberich/flow"
)

func main() {
  store := uberich.NewStore("cookieSecret")
  uberich := uberich.Client("testApp", "http://uberich.example.com", "sharedSecret", store)

  http.Handle("/secret-data", uberich.Protect(SecretHandler))
  http.Handle("/sign-in", uberich.SignIn("http://test.example.com/sign-in", "/secret-data"))
  http.Handle("/sign-out", uberich.SignOut("/")

  http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}
```


## Flow

The authentication flow is for an application (`https://app`) using an uberich
authentication server (`https://uberich`):

1. User visits `https://app` and requests secret data.

2. `https://app` redirects the user to `https://uberich/login`, passing the
   `application` and `redirect_uri` query parameters.

3. User logs in using their registered details.

4. `https://uberich` redirects to `redirect_uri` with the `email` and `verify`
   query parameters.

5. `https://app` checks the `verify` parameter contains `email` hashed with the
   shared secret, then sets a cookie with the User's email address for later
   reference.
