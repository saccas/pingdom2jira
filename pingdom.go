package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/algolia/go-pingdom/pingdom"
)

type PingdomWebhook struct {
	CheckID     int    `json:"check_id"`
	CheckName   string `json:"check_name"`
	CheckType   string `json:"check_type"`
	CheckParams struct {
		BasicAuth  bool   `json:"basic_auth"`
		Encryption bool   `json:"encryption"`
		FullURL    string `json:"full_url"`
		Header     string `json:"header"`
		Hostname   string `json:"hostname"`
		Ipv6       bool   `json:"ipv6"`
		Port       int    `json:"port"`
		URL        string `json:"url"`
	} `json:"check_params"`
	Tags                  []string `json:"tags"`
	PreviousState         string   `json:"previous_state"`
	CurrentState          string   `json:"current_state"`
	ImportanceLevel       string   `json:"importance_level"`
	StateChangedTimestamp int      `json:"state_changed_timestamp"`
	StateChangedUtcTime   string   `json:"state_changed_utc_time"`
	LongDescription       string   `json:"long_description"`
	Description           string   `json:"description"`
	FirstProbe            struct {
		IP       string `json:"ip"`
		Ipv6     string `json:"ipv6"`
		Location string `json:"location"`
	} `json:"first_probe"`
	SecondProbe struct {
		IP       string `json:"ip"`
		Ipv6     string `json:"ipv6"`
		Location string `json:"location"`
		Version  int    `json:"version"`
	} `json:"second_probe"`
}

type PingdomProvider struct {
	client *pingdom.Client
}

func NewPingdomProvider(token string) (*PingdomProvider, error) {
	c, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
		APIToken: token,
	})
	return &PingdomProvider{client: c}, err
}

func (pp PingdomProvider) Reproducer(wh PingdomWebhook) ([]byte, int, bool, error) {
	details, err := pp.client.Checks.Read(wh.CheckID)
	if err != nil {
		return []byte{}, 0, false, err
	}

	switch strings.ToLower(details.Type.Name) {
	case "http":
		method := http.MethodGet
		if details.Type.HTTP.PostData != "" {
			method = http.MethodPost
		}

		proto := "http"
		if details.Type.HTTP.Encryption {
			proto = "https"
		}

		url := fmt.Sprintf("%s://%s:%d%s", proto, details.Hostname, details.Type.HTTP.Port, details.Type.HTTP.Url)

		req, err := http.NewRequest(method, url, strings.NewReader(details.Type.HTTP.PostData))
		if err != nil {
			return []byte{}, 0, false, err
		}

		if details.Type.HTTP.Password != "" || details.Type.HTTP.Username != "" {
			req.SetBasicAuth(details.Type.HTTP.Username, details.Type.HTTP.Password)
		}

		for k, v := range details.Type.HTTP.RequestHeaders {
			req.Header.Add(k, v)
		}

		client := &http.Client{
			Timeout: time.Second * 10,
		}
		res, err := client.Do(req)
		if err != nil {
			return []byte{}, 0, false, err
		}

		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()

		success := true
		if details.Type.HTTP.ShouldContain != "" &&
			!bytes.Contains(body, []byte(details.Type.HTTP.ShouldContain)) {
			success = false
		}
		if details.Type.HTTP.ShouldNotContain != "" &&
			bytes.Contains(body, []byte(details.Type.HTTP.ShouldNotContain)) {
			success = false
		}
		if res.StatusCode < 200 || res.StatusCode > 299 {
			success = false
		}
		return body, res.StatusCode, success, nil

	default:
		return []byte{}, 0, false, fmt.Errorf("check type '%s' cannot be reproduced", strings.ToLower(details.Type.Name))
	}

}
