package main

import (
	"flag"
	"fmt"

	"hawx.me/code/uberich/config"
)

var (
	settingsPath = flag.String("settings", "./settings.toml", "")
)

const usage = `Usage: uberich-admin COMMAND

  Administration tool for uberich.

  Commands:

    list-apps
    set-app NAME ROOTURI SECRET
    remove-app NAME

    list-users
    set-user EMAIL PASSWORD
    remove-user EMAIL
`

func main() {
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Print(usage)
		return
	}

	conf, err := config.Read(*settingsPath)
	if err != nil {
		fmt.Println("config:", err)
		return
	}

	switch flag.Arg(0) {
	case "list-apps":
		for _, app := range conf.Apps {
			fmt.Printf("%s uri=%s secret=%s\n", app.Name, app.URI, app.Secret)
		}

	case "set-app":
		if len(flag.Args()) < 4 {
			fmt.Println("set-app: missing required arguments")
			return
		}

		app := &config.App{
			Name:   flag.Arg(1),
			URI:    flag.Arg(2),
			Secret: flag.Arg(3),
		}

		conf.SetApp(app)

		if err := conf.Save(); err != nil {
			fmt.Println("set-app:", err)
			return
		}

		fmt.Printf("%s uri=%s secret=%s\n", app.Name, app.URI, app.Secret)

	case "remove-app":
		if len(flag.Args()) < 2 {
			fmt.Println("remove: missing required argument")
			return
		}

		conf.RemoveApp(flag.Arg(1))

		if err := conf.Save(); err != nil {
			fmt.Println("remove-app:", err)
			return
		}
	case "list-users":
		for _, user := range conf.Users {
			fmt.Printf("%s\n", user.Email)
		}

	case "set-user":
		if len(flag.Args()) < 3 {
			fmt.Println("set-user: missing required arguments")
			return
		}

		user := &config.User{
			Email: flag.Arg(1),
		}
		user.SetPassword(flag.Arg(2))

		conf.SetUser(user)

		if err := conf.Save(); err != nil {
			fmt.Println("set-user:", err)
			return
		}

		fmt.Printf("%s\n", user.Email)

	case "remove-user":
		if len(flag.Args()) < 2 {
			fmt.Println("remove-user: missing required argument")
			return
		}

		conf.RemoveUser(flag.Arg(1))

		if err := conf.Save(); err != nil {
			fmt.Println("remove-user:", err)
			return
		}

	default:
		fmt.Print(usage)
	}
}
