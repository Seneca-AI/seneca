package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	st "seneca/api/type"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/dao/userdao"
	"seneca/internal/util/data"
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
	sqlService, err := datastore.New(context.TODO(), "senecacam-sandbox")
	if err != nil {
		log.Fatalf("Error initializing datastore service %v", err)
	}

	userID := "5697673499770880"
	if err := data.DeleteAllUserDataInDB(userID, false, sqlService); err != nil {
		log.Fatalf("DeleteAllUserDataInDB(%s, %t, _) returns err: %v", userID, false, err)
	}
}
