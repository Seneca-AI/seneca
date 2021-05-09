package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	st "seneca/api/type"
	"seneca/internal/client/cloud/gcp"
	"time"
)

// For interfacing with cloud datastore.
func main() {
	cloudDatastore, err := gcp.NewGoogleCloudDatastoreClient(context.TODO(), "senecacam-sandbox", time.Minute)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	pathToOauthToken := "../google_drive/token.json"
	oauthFile, err := os.Open(pathToOauthToken)
	if err != nil {
		log.Fatalf("Error opening token file %q - err: %v", pathToOauthToken, err)
	}
	defer oauthFile.Close()

	bytes, err := ioutil.ReadAll(oauthFile)
	if err != nil {
		log.Fatalf("Error reading token file %q - err: %v", pathToOauthToken, err)
	}

	user := &st.User{
		Email:      "testuser002@gmail.com",
		OauthToken: bytes,
	}

	_, err = cloudDatastore.InsertUniqueUser(user)
	if err != nil {
		log.Fatalf("Error inserting user err: %v", err)
	}
}
