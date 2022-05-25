package main

import (
	"flag"
	"fmt"
	"os"
)

// App holds the parsed flag values.
type App struct {
	address    string
	mode       string
	configFile string
}

func main() {
	var app App
	flag.StringVar(&app.address, "a", ":8080", "ip:port to listen on (runs as lambda if empty)")
	flag.StringVar(&app.mode, "m", "local", "mode, can me either 'local', 'azurefunc' or 'awslambda'")
	flag.StringVar(&app.configFile, "c", "./config.yaml", "name of the config file")
	flag.Parse()

	if app.version {
		fmt.Println(versionInfo())
		return
	}

	c, err := NewConfig(app.configFile)
	exitOnErr(err)

	s, err := NewServer(app.address, c)
	exitOnErr(err)

	err = s.run(app.mode)
	exitOnErr(err)
}

func exitOnErr(errs ...error) {
	errNotNil := false
	for _, err := range errs {
		if err == nil {
			continue
		}
		errNotNil = true
		fmt.Fprintf(os.Stderr, "ERROR: %s", err.Error())
	}
	if errNotNil {
		fmt.Print("\n")
		os.Exit(-1)
	}
}
