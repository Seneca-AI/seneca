package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

// devtools offers tools to help developers do certain things that would not be
// done in the actual application.  For now, all functions are private.  If you want
// to run a function just put it in the main function and run "go run ."

// generateDriveToken generates a token with the given scope using the path to the given oauth credentials and saves it to the given output path.
func generateDriveToken(pathToOAuthCredentials, pathToOutputToken, scope string) error {
	b, err := ioutil.ReadFile(pathToOAuthCredentials)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile(%s) returns err: %w", pathToOAuthCredentials, err)
	}

	oauthConfig, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		return fmt.Errorf("google.ConfigFromJSON(%s, drive.DriveMetadataScope) returns err: %w", pathToOAuthCredentials, err)
	}

	authURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return fmt.Errorf("unable to read authorization code %v", err)
	}

	tok, err := oauthConfig.Exchange(context.TODO(), authCode)
	if err != nil {
		return fmt.Errorf("unable to retrieve token from web %v", err)
	}

	fmt.Printf("Saving credential file to: %s\n", pathToOutputToken)
	f, err := os.OpenFile(pathToOutputToken, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(tok)
	return nil
}

func main() {
	if err := generateDriveToken("oauth.json", "token.json", drive.DriveScope); err != nil {
		log.Fatal(err)
	}
}
