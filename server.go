package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/apex/gateway"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	yaml "gopkg.in/yaml.v2"
)

// Sever holds all dependencies of the webserver
type Server struct {
	listener string
	mode     string
	handler  http.Handler

	pingdom *PingdomProvider
	jira    *JiraConnector
}

func NewServer(listener string, c *Config) (*Server, error) {
	p, err := NewPingdomProvider(c.Pingdom.Token)
	if err != nil {
		return nil, err
	}
	j, err := NewJiraConnector(c.Jira.URL, c.Jira.Username, c.Jira.Password)
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		pingdom:  p,
		jira:     j,
	}

	r := mux.NewRouter().StrictSlash(true)
	routes := s.Routes()
	routes.Populate(r, c.PathPrefix)
	s.handler = alice.New(s.LoggerMiddleware).Then(r)
	return s, nil
}

func (s Server) run(mode string) error {
	switch strings.ToLower(mode) {
	case "local":
		if s.listener == "" {
			return fmt.Errorf("No listener defined")
		}
		fmt.Sprintf("Running locally at '%s'...\n", s.listener)
		return http.ListenAndServe(s.listener, s.handler)
	case "azurefunc":
		port, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT")
		if !ok {
			return fmt.Errorf("Environment FUNCTIONS_CUSTOMHANDLER_PORT not defined")
		}
		listener := fmt.Sprintf(":%s", port)
		fmt.Sprintf("Running as Azure Function at '%s'...\n", listener)
		return http.ListenAndServe(listener, s.handler)
	case "awslambda":
		fmt.Sprintf("Running as AWS Lambda...\n")
		return gateway.ListenAndServe(s.listener, s.handler)
	default:
		return fmt.Errorf("Unknown mode '%s'", mode)
	}
}

func (s Server) respond(res http.ResponseWriter, req *http.Request, code int, data interface{}) {
	var err error
	var errMesg []byte
	var out []byte

	f := req.Header.Get("Accept")
	if f == "text/yaml" {
		res.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		out, err = yaml.Marshal(data)
		errMesg = []byte("--- error: failed while rendering data to yaml")
	} else {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		out, err = json.MarshalIndent(data, "", "    ")
		errMesg = []byte("{ 'error': 'failed while rendering data to json' }")
	}

	if err != nil {
		out = errMesg
		code = http.StatusInternalServerError
	}
	res.WriteHeader(code)
	res.Write(out)
}

func (s Server) raw(res http.ResponseWriter, code int, data []byte) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(code)
}
