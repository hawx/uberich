package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"hawx.me/code/serve"
	"hawx.me/code/uberich/config"
	"hawx.me/code/uberich/web"
)

const help = `Usage: uberich [options]

  Uberich is a login server. It uses a simple config file for storage of
  settings and understands the concept of users and apps.

  A user is registered by email address and has a password.
  An app has a name, a URI and a secret.

  See the correspoding hawx.me/code/uberich/flow package that implements
  helpers for integrating clients with uberich.

 OPTIONS

   --settings PATH    # Read settings from path (default: './settings.toml')
   --port PORT        # Serve on given port (default: '8080')
   --socket PATH      # Serve at given socket, instead

 SETTINGS

   The settings file is written in TOML and must contain at least the following

     # the domain uberich is running at
     domain = "my.example.com"

     # whether to use secure cookies, only set false in local development
     secure = true

     # 32 or 64 byte, in standard base64, used to authenticate the cookie
     # using HMAC
     hashKey = "..."

     # Encryption key for the cookie, the length corresponds to the algorithm
     # used: for AES, used by default, valid lengths are 16, 24, or 32 bytes
     # to select AES-128, AES-192, or AES-256. Given in standard base64.
     blockKey = "..."

   To add users and apps see uberich/cmd/uberich-admin.
`

func main() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, help) }

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

	handler, err := web.New(conf)
	if err != nil {
		log.Println("config:", err)
		return
	}

	serve.Serve(*port, *socket, handler)
}
