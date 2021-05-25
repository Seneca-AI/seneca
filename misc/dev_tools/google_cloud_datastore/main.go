package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	st "seneca/api/type"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/client/logging"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/userdao"
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

func deleteAllRawVideosForUser(sqlService *datastore.Service, userID string) error {
	logger := logging.NewLocalLogger(false)
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlService, logger, time.Second*5)

	rawVideoIDs, err := rawVideoDAO.ListUserRawVideoIDs(userID)
	if err != nil {
		return err
	}

	for _, rid := range rawVideoIDs {
		if err := rawVideoDAO.DeleteRawVideoByID(rid); err != nil {
			return err
		}
	}

	return nil
}

func deleteAllRawLocationsForUser(sqlService *datastore.Service, userID string) error {
	rawLocationDAO := rawlocationdao.NewSQLRawLocationDAO(sqlService)

	rawLocationIDs, err := rawLocationDAO.ListUserRawLocationIDs(userID)
	if err != nil {
		return err
	}

	for _, rlid := range rawLocationIDs {
		if err := rawLocationDAO.DeleteRawLocationByID(rlid); err != nil {
			return err
		}
	}

	return nil
}

func deleteAllRawMotionsForUser(sqlService *datastore.Service, userID string) error {
	logger := logging.NewLocalLogger(false)
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(sqlService, logger)

	rawMotionIDs, err := rawMotionDAO.ListUserRawMotionIDs(userID)
	if err != nil {
		return err
	}

	fmt.Printf("Got %d rawMotions\n", len(rawMotionIDs))
	for _, rlid := range rawMotionIDs {
		if err := rawMotionDAO.DeleteRawMotionByID(rlid); err != nil {
			return err
		}
	}

	return nil
}

// For interfacing with cloud datastore.
func main() {
	sqlService, err := datastore.New(context.TODO(), "senecacam-staging")
	if err != nil {
		log.Fatalf("Error initializing datastore service %v", err)
	}

	insertUserWithToken(sqlService, "itestuser000@senecacam.com", "token.json")
}
