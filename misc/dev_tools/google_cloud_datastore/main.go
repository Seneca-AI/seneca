package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	st "seneca/api/type"
	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/client/logging"
	"seneca/internal/dao/userdao"
	"seneca/internal/util/data"
	"time"
)

func insertUserWithToken(sqlService *datastore.Service, email, pathToOauthToken string) {
	userDAO := userdao.NewSQLUserDAO(sqlService)

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
		Email:      email,
		OauthToken: bytes,
	}

	_, err = userDAO.InsertUniqueUser(user)
	if err != nil {
		log.Fatalf("Error inserting user err: %v", err)
	}
}

// For interfacing with cloud datastore.
func main() {
	projectID := "senecacam-sandbox"

	sqlService, err := datastore.New(context.TODO(), projectID)
	if err != nil {
		log.Fatalf("Error initializing datastore service %v", err)
	}

	gcsc, err := gcp.NewGoogleCloudStorageClient(context.Background(), projectID, time.Second*10, time.Minute)
	if err != nil {
		log.Fatalf("Error initializing cloud storage service %v", err)
	}

	logger := logging.NewLocalLogger(false)

	userID := "5685335367352320"
	if err := data.DeleteAllUserData(userID, false, sqlService, gcsc, logger); err != nil {
		log.Fatalf("DeleteAllUserDataInDB(%s, %t, _) returns err: %v", userID, false, err)
	}
}
