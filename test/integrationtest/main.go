package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"seneca/env"
	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/client/logging"
	"seneca/internal/dao/userdao"
	"seneca/test/integrationtest/apiserver"
	"seneca/test/integrationtest/syncer"
	"time"
)

const testUserEmail = "itestuser000@senecacam.com"

func main() {
	startTime := time.Now()
	failureMessage := ""

	if err := env.ValidateEnvironmentVariables(); err != nil {
		log.Fatalf("Failed to ValidateEnvironmentVariables: %s", err.Error())
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	ctx := context.TODO()
	// Initialize clients.
	logger, err := logging.NewGCPLogger(ctx, "itestserver", projectID)
	if err != nil {
		log.Fatalf("logging.NewGCPLogger() returns - err: %v\n", err)
		return
	}

	gcsc, err := gcp.NewGoogleCloudStorageClient(ctx, projectID, time.Second*10, time.Minute)
	if err != nil {
		failureMessage = fmt.Sprintf("cloud.NewGoogleCloudStorageClient() returns - err: %v", err)
		logger.Critical(failureMessage)
		log.Fatal(failureMessage)
	}

	sqlService, err := datastore.New(ctx, projectID)
	if err != nil {
		failureMessage = fmt.Sprintf("Error initializing sql service - err: %v", err)
		logger.Critical(failureMessage)
		log.Fatal(failureMessage)
	}

	userDAO := userdao.NewSQLUserDAO(sqlService)

	if err := syncer.E2ESyncer(projectID, testUserEmail, sqlService, gcsc, userDAO, logger); err != nil {
		failureMessage = fmt.Sprintf("E2ESyncerIntegrationTest failed: %v", err)
		logger.Error(failureMessage)
	}

	if err := apiserver.E2EAPIServer(projectID, testUserEmail, sqlService, userDAO, logger); err != nil {
		failureMessage = fmt.Sprintf("E2EAPIServerIntegrationTest failed: %v", err)
		logger.Error(failureMessage)
	}

	logger.Log(fmt.Sprintf("Integration tests completed in %v\n", time.Since(startTime)))
	// Sleep for 2 seconds so the program doesn't end before the logger gets the opportunity to write.
	time.Sleep(2 * time.Second)

	if failureMessage != "" {
		log.Fatal(failureMessage)
	}
}
