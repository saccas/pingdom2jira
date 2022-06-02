package main

import (
	"flag"
	"fmt"
	"os"
)

const configFileEnv = "CONFIG_FILE"

// App holds the parsed flag values.
type App struct {
	address    string
	mode       string
	configFile string
	version    bool
}

func main() {
	var app App
	flag.StringVar(&app.address, "a", ":8080", "ip:port to listen on (runs as lambda if empty)")
	flag.StringVar(&app.configFile, "c", "./config.yaml", fmt.Sprintf("name of the config file (alternatively the config file be specified via environment variable '%s'))", configFileEnv))
	flag.StringVar(&app.mode, "m", "local", "mode, can me either 'local', 'azurefunc' or 'awslambda'")
	flag.BoolVar(&app.version, "v", false, "print version information")
	flag.Parse()

	envConfig := os.Getenv(configFileEnv)
	if envConfig != "" {
		app.configFile = envConfig
	}

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
