package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type DeviceFlowResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type TokenResponse struct {
	Scopes       string `json:"scope"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type Authenticator struct {
	TenantID      string
	ClientID      string
	DeviceFlowURL string
	TokenURL      string
	Scopes        []string
	logDebug      func(string, ...any)
	logInfo       func(string, ...any)
}

func NewAuthenticator(tenantID string, clientID string, deviceFlowURL string,
	tokenURL string, scopes []string) *Authenticator {

	return &Authenticator{TenantID: tenantID, ClientID: clientID, DeviceFlowURL: deviceFlowURL,
		TokenURL: tokenURL, Scopes: scopes, logInfo: logTo(os.Stderr), logDebug: noOpLog()}
}

func (a *Authenticator) Authenticate(ctx context.Context) (string, error) {
	deviceFlowResp, err := a.requestDeviceFlow(ctx)
	if err != nil {
		return "", err
	}

	a.logInfo("To sign in, use a web browser to open the page %s and enter the code %s\n",
		deviceFlowResp.VerificationURL, deviceFlowResp.UserCode)

	tokenResp, err := a.pollForToken(ctx, deviceFlowResp)
	if err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func (a *Authenticator) requestDeviceFlow(_ context.Context) (*DeviceFlowResponse, error) {
	deviceCodeURL := fmt.Sprintf(a.DeviceFlowURL, a.TenantID)
	resp, err := http.PostForm(deviceCodeURL, url.Values{"client_id": {a.ClientID}, "scope": {strings.Join(a.Scopes, " ")}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed with status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var deviceFlowResp DeviceFlowResponse
	err = json.Unmarshal(body, &deviceFlowResp)
	if err != nil {
		return nil, err
	}

	return &deviceFlowResp, nil
}

func (a *Authenticator) pollForToken(ctx context.Context, deviceFlowResp *DeviceFlowResponse) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf(a.TokenURL, a.TenantID)
	interval := time.Duration(deviceFlowResp.Interval) * time.Second
	expiresAt := time.Now().Add(time.Duration(deviceFlowResp.ExpiresIn) * time.Second)

	for {
		if time.Now().After(expiresAt) {
			return nil, fmt.Errorf("authentication request expired")
		}

		time.Sleep(interval)

		a.logDebug("Polling for token...\n")
		resp, err := http.PostForm(tokenURL, url.Values{
			"client_id":   {a.ClientID},
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"device_code": {deviceFlowResp.DeviceCode},
			"scope":       {strings.Join(a.Scopes, " ")},
		})
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		a.logDebug("post form response code was %d\n", resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var tokenResp TokenResponse
			err = json.Unmarshal(body, &tokenResp)
			if err != nil {
				return nil, err
			}

			a.logDebug("Got token response: %+v\n", tokenResp)
			return &tokenResp, nil
		}

		if resp.StatusCode != http.StatusBadRequest {
			return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
		}
	}
}

func noOpLog() func(string, ...any) {
	return func(str string, args ...any) {
		// no-op
	}
}

func logTo(file *os.File) func(string, ...any) {
	return func(str string, args ...any) {
		_, _ = fmt.Fprintf(file, str, args...)
	}
}
