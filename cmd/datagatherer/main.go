package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"seneca/internal/datagatherer/rawvideohandler"
	"seneca/internal/util/cloud"
	"seneca/internal/util/logging"
)

const (
	port = "8080"
)

// TODO: make this configurable in different envs
func main() {
	// Initialize storage client and RawVideoHandler.
	ctx := context.Background()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		fmt.Fprintf(os.Stderr, "GOOGLE_CLOUD_PROJECT environment variable must be set.\n")
		os.Exit(1)
	}

	logger, err := logging.NewGCPLogger(ctx, "datagatherer", projectID)
	if err != nil {
		fmt.Printf("logging.NewGCPLogger() returns - err: %v", err)
		return
	}

	gcsc, err := cloud.NewGoogleCloudStorageClient(ctx, projectID, time.Second*10, time.Minute)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewGoogleCloudStorageClient() returns - err: %v", err))
		return
	}

	gcsd, err := cloud.NewGoogleCloudDatastoreClient(ctx, projectID, time.Second)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewGoogleCloudDatastoreClient() returns - err: %v", err))
		return
	}

	rawVideoHandler, err := rawvideohandler.NewRawVideoHandler(gcsc, gcsd, logger, "", projectID)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewRawVideoHandler() returns - err: %v", err))
		return
	}

	http.HandleFunc("/rawvideo", rawVideoHandler.HandleRawVideoPostRequest)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
