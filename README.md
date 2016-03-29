# uberich

A small-scale 3rd party authentication server.

Designed for registering and authenticating a whitelisted set of
users. Authenticating in the scope of this application means saying "yes that is
an email address I know about and this is the person it belongs to". It is
clearly inspired by Persona in this regard, but does not aim to be a replacement
for Persona.

I am not using this at the moment, and neither should you.


## Flow

The authentication flow is for an application (`https://app`) using an uberich
authentication server (`https://uberich`):

1. User visits `https://app` and requests secret data.
2. `https://app` redirects the user to `https://uberich/login`.
3. User logs in using their registered details.
4. `https://uberich` redirects to the URI given by `https://app` with the email
   in the query string.
5. `https://app` sets a cookie with the User's email address for later
   reference.

### Things missing

- Hard for `https://app` to know the redirect was from
  `https://uberich`. Applications should be registered with uberich so that some
  shared data can be used to HMAC the email. Then `https://app` can be sent this
  with the email and calculate the HMAC itself to confirm.
