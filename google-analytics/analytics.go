package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const gaHost = "https://www.google-analytics.com"

// Event is an event that happened in an ofc installation.
type Event struct {
	Action   string `json:"action"`
	Category string `json:"category"`
	User     string `json:"user"`
}

// Client is a google analytics client.
type Client struct {
	client     *http.Client
	events     chan Event
	clientID   string
	trackingID string
	appName    string
	appVersion string
}

// NewClient creates a new google analytics client.
func NewClient(clientID, trackingID, appName, appVersion string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		events:     make(chan Event, 10),
		clientID:   clientID,
		trackingID: trackingID,
		appName:    appName,
		appVersion: appVersion,
	}
}

// ReportHandler handles HTTP requests and reports
// the events that are received to google analytics.
func (c *Client) ReportHandler(w http.ResponseWriter, r *http.Request) {
	var (
		msg  string
		code = http.StatusOK
	)
	defer func() {
		w.WriteHeader(code)
		w.Write([]byte(msg))
	}()
	if r.Method != http.MethodPost {
		msg = "Only supported method is POST"
		code = http.StatusMethodNotAllowed
		return
	}
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		msg = "Events must to be sent as json payloads"
		code = http.StatusBadRequest
		return
	}

	var e Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		msg = fmt.Sprintf("Invalid event payload: %s", err)
		code = http.StatusBadRequest
		return
	}

	if err := c.Report(e); err != nil {
		msg = fmt.Sprintf("Unable to report analytics event: %s", err)
		code = http.StatusInternalServerError
		return
	}

	msg = "Event successfully reported"
}

// Report an analytics event to google.
func (c *Client) Report(event Event) error {
	c.events <- event
	return nil
}

// ListenAndSend listens for events and sends
// them to the google analytics api.
func (c *Client) ListenAndSend(ctx context.Context) {
	for e := range c.events {
		req, err := http.NewRequest(http.MethodPost, gaHost+"/collect", c.TransformAndEncode(e))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(ctx)
		res, err := c.client.Do(req)
		if err != nil {
			log.Printf("Request failed: %s", err)
			continue
		}
		res.Body.Close()
		log.Printf("Google status code: %d", res.StatusCode)
	}
}

// TransformAndEncode transforms the incoming event
// into a google analytics event and encodes it
// as url values.
func (c *Client) TransformAndEncode(e Event) io.Reader {
	form := url.Values{}
	form.Add("v", "1")
	form.Add("t", "event")
	form.Add("cid", c.clientID)
	form.Add("tid", c.trackingID)
	form.Add("an", c.appName)
	form.Add("av", c.appVersion)
	form.Add("aip", "0")
	form.Add("cd1", e.User)
	form.Add("ec", e.Category)
	form.Add("ea", e.Action)
	return strings.NewReader(form.Encode())
}
