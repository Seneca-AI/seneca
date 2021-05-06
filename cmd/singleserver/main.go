// Package main in singleserver starts the entire Seneca application on
// a single server utilizing channels to mimic HTTP request routing.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"seneca/env"
	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/logging"
	"seneca/internal/datagatherer/rawvideohandler"
	"seneca/internal/syncer"
	"seneca/internal/util/mp4"
	"time"
)

const (
	endpoint = "/syncer"
	port     = "6060"
)

func main() {
	if err := env.ValidateEnvironmentVariables(); err != nil {
		log.Fatalf("Error in ValidateEnvironmentVariables: %v", err)
	}

	ctx := context.TODO()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		fmt.Fprintf(os.Stderr, "GOOGLE_CLOUD_PROJECT environment variable must be set.\n")
		os.Exit(1)
	}

	logger, err := logging.NewGCPLogger(ctx, "singleserver", projectID)
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

	mp4Tool, err := mp4.NewMP4Tool(logger)
	if err != nil {
		logger.Critical(fmt.Sprintf("mp4.NewMP4Tool() returns - err: %v", err))
		return
	}

	rawVideoHandler, err := rawvideohandler.NewRawVideoHandler(gcsc, gcsd, mp4Tool, logger, projectID)
	if err != nil {
		logger.Critical(fmt.Sprintf("cloud.NewRawVideoHandler() returns - err: %v", err))
		return
	}

	gDriveFactory := &googledrive.UserClientFactory{}

	syncer := syncer.New(rawVideoHandler, gDriveFactory, gcsd, logger)
	handler := &HTTPHandler{
		syncer: syncer,
	}

	http.HandleFunc(endpoint, handler.RunSyncer)

	fmt.Printf("Starting server at port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}

type HTTPHandler struct {
	syncer *syncer.Syncer
}

func (handler *HTTPHandler) RunSyncer(w http.ResponseWriter, r *http.Request) {
	handler.syncer.ScanAllUsers()
}
