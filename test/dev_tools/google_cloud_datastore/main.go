package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	st "seneca/api/type"
	"seneca/internal/client/cloud/gcpdatastore"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/userdao"
	"time"
)

func insertUserWithToken(sqlService *gcpdatastore.Service, email, pathToOauthToken string) {
	userDAO := userdao.NewSQLUserDao(sqlService)

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

func deleteAllRawVideosForUser(sqlService *gcpdatastore.Service, userID string) error {
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlService, time.Second*5)

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

func deleteAllRawLocationsForUser(sqlService *gcpdatastore.Service, userID string) error {
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

func deleteAllRawMotionsForUser(sqlService *gcpdatastore.Service, userID string) error {
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(sqlService)

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
	sqlService, err := gcpdatastore.New(context.TODO(), "senecacam-staging")
	if err != nil {
		log.Fatalf("Error initializing datastore service %v", err)
	}

	insertUserWithToken(sqlService, "itestuser000@senecacam.com", "token.json")
}
