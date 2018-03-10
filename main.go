package main

import (
	"log"
	"os"

	"github.com/dynport/dgtk/cli"
)

var logger = log.New(os.Stderr, "", 0)

func main() {
	router := cli.NewRouter()
	router.Register("tasks/list", &tasksList{}, "List Running Tasks")
	router.Register("hosts/list", &hostsList{}, "List Running Instances")
	switch err := router.RunWithArgs(); err {
	case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
		// ignore
		return
	default:
		logger.Fatal(err)
	}
}
