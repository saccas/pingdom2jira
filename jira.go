package main

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira"
)

type JiraConnector struct {
	client *jira.Client
}

func NewJiraConnector(url, username, password string) (*JiraConnector, error) {
	tp := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	c, err := jira.NewClient(tp.Client(), url)
	if err != nil {
		return nil, err
	}
	j := &JiraConnector{
		client: c,
	}
	return j, nil
}

func (j JiraConnector) CreateTicket(project, accountID, title, desc, issueType string) error {
	i := jira.Issue{
		Fields: &jira.IssueFields{
			Assignee: &jira.User{
				AccountID: accountID,
			},
			Description: desc,
			Type: jira.IssueType{
				Name: issueType,
			},
			Project: jira.Project{
				Key: project,
			},
			Summary: title,
		},
	}

	_, _, err := j.client.Issue.Create(&i)
	if err != nil {
		return err
	}

	return nil
}

func (j JiraConnector) GetUserAccountID(email string) (string, error) {
	users, _, err := j.client.User.Find(email)
	if err != nil {
		return "", err
	}

	if len(users) != 1 {
		return "", fmt.Errorf("Email '%s' does match %d users, must match exactly one", email, len(users))
	}

	id := users[0].AccountID

	return id, nil
}
