package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/logging"
	"seneca/internal/datagatherer/cutvideohandler"
	"seneca/internal/datagatherer/rawvideohandler"
	"seneca/internal/util/mp4"
)

const (
	port                 = "8080"
	rawVideoEndpointPath = "rawvideo"
	cutVideoEndpointPath = "cutvideo"
)

// TODO: make this configurable in different envs.
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

	gcsc, err := gcp.NewGoogleCloudStorageClient(ctx, projectID, time.Second*10, time.Minute)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewGoogleCloudStorageClient() returns - err: %v", err))
		return
	}

	gcsd, err := gcp.NewGoogleCloudDatastoreClient(ctx, projectID, time.Second)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewGoogleCloudDatastoreClient() returns - err: %v", err))
		return
	}

	mp4Tool, err := mp4.NewMP4Tool()
	if err != nil {
		logger.Critical(fmt.Sprintf("mp4.NewMP4Tool() returns - err: %v", err))
		return
	}

	rawVideoHandler, err := rawvideohandler.NewRawVideoHandler(gcsc, gcsd, mp4Tool, logger, projectID)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewRawVideoHandler() returns - err: %v", err))
		return
	}

	cutVideoHandler, err := cutvideohandler.NewCutVideoHandler(gcsc, gcsd, mp4Tool, logger, projectID)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewCutVideoHandler() returns - err: %v", err))
		return
	}

	http.HandleFunc(fmt.Sprintf("/%s", rawVideoEndpointPath), rawVideoHandler.HandleRawVideoPostRequest)
	http.HandleFunc(fmt.Sprintf("/%s", cutVideoEndpointPath), cutVideoHandler.HandleRawVideoPostRequest)

	fmt.Printf("Starting server at port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
