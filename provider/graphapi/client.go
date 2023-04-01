package graphapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Attendee struct {
	Type         string       `json:"type"`
	EmailAddress EmailAddress `json:"emailAddress"`
}

type EmailAddress struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Event struct {
	ID        string     `json:"id,omitempty"`
	Subject   string     `json:"subject,omitempty"`
	StartTime DateTime   `json:"start,omitempty"`
	EndTime   DateTime   `json:"end,omitempty"`
	Location  Location   `json:"location,omitempty"`
	Attendees []Attendee `json:"attendees,omitempty"`
	ShowAs    string     `json:"showAs,omitempty"`
}

type DateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type Location struct {
	DisplayName string `json:"displayName"`
}

type GraphAPIClient struct {
	client   *http.Client
	token    string
	endpoint string
}

func NewGraphAPIClient(token string) *GraphAPIClient {
	return &GraphAPIClient{
		client:   &http.Client{},
		token:    token,
		endpoint: "https://graph.microsoft.com/v1.0",
	}
}

func (g *GraphAPIClient) sendRequest(ctx context.Context, method, url string, payload []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Content-Type", "application/json")

	return g.client.Do(req)
}

// ReadEvents reads events (Scopes: Calendars.Read, Calendars.Read.Shared)
func (g *GraphAPIClient) ReadEvents(ctx context.Context) ([]Event, error) {
	url := g.endpoint + "/me/events"
	resp, err := g.sendRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var eventsResp struct {
		Events []Event `json:"value"`
	}
	err = json.Unmarshal(body, &eventsResp)
	if err != nil {
		return nil, err
	}

	return eventsResp.Events, nil
}

// CreateEvent creates an event (Scopes: Calendars.ReadWrite, Calendars.ReadWrite.Shared)
func (g *GraphAPIClient) CreateEvent(ctx context.Context, event Event) (*Event, error) {
	url := g.endpoint + "/me/events"
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	resp, err := g.sendRequest(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var createdEvent Event
	err = json.Unmarshal(body, &createdEvent)
	if err != nil {
		return nil, err
	}

	return &createdEvent, nil
}

// UpdateEvent updates an event (Scopes: Calendars.ReadWrite, Calendars.ReadWrite.Shared)
func (g *GraphAPIClient) UpdateEvent(ctx context.Context, eventID string, updatedEvent Event) error {
	url := g.endpoint + "/me/events/" + eventID
	payload, err := json.Marshal(updatedEvent)
	if err != nil {
		return err
	}

	resp, err := g.sendRequest(ctx, http.MethodPatch, url, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	return nil
}

// DeleteEvent delets an event (Scopes: Calendars.ReadWrite, Calendars.ReadWrite.Shared)
func (g *GraphAPIClient) DeleteEvent(ctx context.Context, eventID string) error {
	url := g.endpoint + "/me/events/" + eventID
	resp, err := g.sendRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	return nil
}
