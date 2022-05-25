package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/schema"
)

func (s Server) HookHandler(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		s.respond(res, req, http.StatusInternalServerError, fmt.Sprintf("could not read request body: %s", err.Error()))
		return
	}

	payload := &PingdomWebhook{}
	err = json.Unmarshal(body, payload)
	if err != nil {
		s.respond(res, req, http.StatusInternalServerError, fmt.Sprintf("could not parse request body: %s", err.Error()))
		return
	}

	switch strings.ToLower(payload.CurrentState) {
	case "down":
		// Get Params
		params := &HookQueryParams{}
		err = params.Parse(req)
		if err != nil {
			s.respond(res, req, http.StatusInternalServerError, err.Error())
			return
		}

		err = params.Validate()
		if err != nil {
			s.respond(res, req, http.StatusInternalServerError, err.Error())
			return
		}

		// Reproduct check
		data, code, _, err := s.pingdom.Reproducer(*payload)
		if err != nil {
			s.respond(res, req, http.StatusInternalServerError, err.Error())
			return
		}

		message := fmt.Sprintf("PINGDOM has reported an error on a check named '%s'.\nThe body of the failed request (HTTP Code %d) was:\n---\n%s", payload.CheckName, code, string(data))

		id, err := s.jira.GetUserAccountID(params.JiraAssigneeEmail)
		if err != nil {
			s.respond(res, req, http.StatusInternalServerError, err.Error())
			return
		}

		err = s.jira.CreateTicket(params.JiraProject,
			id,
			fmt.Sprintf("PINGDOM Check Error: %s", payload.CheckName),
			message,
			params.JiraType,
		)
		if err != nil {
			s.respond(res, req, http.StatusInternalServerError, err.Error())
			return
		}
		s.respond(res, req, http.StatusOK, "sent")
		return
	default:
		s.respond(res, req, http.StatusOK, fmt.Sprintf("no action taken for status '%s'", strings.ToLower(payload.CurrentState)))
		return
	}
}

type HookQueryParams struct {
	JiraAssigneeEmail string `schema:"jira_assignee_email"`
	JiraProject       string `schema:"jira_project"`
	JiraType          string `schema:"jira_type"`
}

func (r *HookQueryParams) Default() {}

func (r *HookQueryParams) Parse(req *http.Request) error {
	r.Default()
	return schema.NewDecoder().Decode(r, req.URL.Query())
}

func (r *HookQueryParams) Validate() error {
	if r.JiraType == "" {
		return fmt.Errorf("jira_type must be set")
	}
	if r.JiraAssigneeEmail == "" {
		return fmt.Errorf("jira_assignee_email must be set")
	}
	if r.JiraProject == "" {
		return fmt.Errorf("jira_project must be set")
	}
	return nil
}
