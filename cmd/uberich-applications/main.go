package main

import (
	"flag"
	"fmt"

	"hawx.me/code/uberich/data"
)

var (
	dbPath = flag.String("database", "./uberich.db", "")
)

const usage = `Usage: uberich-applications COMMAND

  Administration tool for uberich.

    list
        Lists all registered applications

    set NAME ROOTURI SECRET
        Adds/updates an application

    remove NAME
        Removes an application
`

func main() {
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Print(usage)
		return
	}

	db, err := data.Open(*dbPath)
	if err != nil {
		fmt.Println("db:", err)
		return
	}
	defer db.Close()

	switch flag.Arg(0) {
	case "list":
		apps, err := db.ListApplications()
		if err != nil {
			fmt.Println("list:", err)
			return
		}

		for _, app := range apps {
			printApplication(app)
		}
	case "set":
		if len(flag.Args()) < 4 {
			fmt.Println("set: missing required arguments")
			return
		}

		app := data.Application{
			Name:    flag.Arg(1),
			RootURI: flag.Arg(2),
			Secret:  flag.Arg(3),
		}

		err := db.SetApplication(app)
		if err != nil {
			fmt.Println("set:", err)
			return
		}
		printApplication(app)

	case "remove":
		if len(flag.Args()) < 2 {
			fmt.Println("remove: missing required argument")
			return
		}

		err := db.RemoveApplication(flag.Arg(1))
		if err != nil {
			fmt.Println("remove:", err)
			return
		}

	default:
		fmt.Print(usage)
	}
}

func printApplication(app data.Application) {
	fmt.Printf("%s\trootURI=%s\tsecret=%s\n", app.Name, app.RootURI, app.Secret)
}
