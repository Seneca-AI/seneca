package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"seneca/env"
	"seneca/internal/client/logging"
	"seneca/test/integrationtest/testenv"
	"seneca/test/integrationtest/tests"
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

	testEnv, err := testenv.New(projectID, logger)
	if err != nil {
		log.Fatalf("testenv.New() returns - err: %v\n", err)
		return
	}

	if err := tests.E2ESyncer(testUserEmail, testEnv); err != nil {
		failureMessage = fmt.Sprintf("E2ESyncerIntegrationTest failed: %v", err)
		logger.Error(failureMessage)
	}

	if err := tests.E2ERunner(testUserEmail, testEnv); err != nil {
		failureMessage = fmt.Sprintf("E2ERunnerIntegrationTest failed: %v", err)
		logger.Error(failureMessage)
	}

	if err := tests.E2EAPIServer(testUserEmail, testEnv); err != nil {
		failureMessage = fmt.Sprintf("E2EAPIServerIntegrationTest failed: %v", err)
		logger.Error(failureMessage)
	}

	if err := tests.E2ESource(testUserEmail, testEnv); err != nil {
		failureMessage = fmt.Sprintf("E2ESourcePlumbed failed: %v", err)
		logger.Error(failureMessage)
	}

	logger.Log(fmt.Sprintf("Integration tests completed in %v\n", time.Since(startTime)))
	// Sleep for 2 seconds so the program doesn't end before the logger gets the opportunity to write.
	time.Sleep(2 * time.Second)

	if failureMessage != "" {
		log.Fatal(failureMessage)
	}
}
