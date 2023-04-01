package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	authenticator := NewAuthenticator(
		"common",
		os.Getenv("MICROSOFT_APP_CLIENT_ID"),
		"https://login.microsoftonline.com/%s/oauth2/v2.0/devicecode",
		"https://login.microsoftonline.com/%s/oauth2/v2.0/token",
		[]string{"Calendars.ReadWrite"},
	)

	accessToken, err := authenticator.Authenticate(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error authenticating: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(accessToken)
}
