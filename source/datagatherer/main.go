package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"seneca/source/datagatherer/rawvideohandler"
	"seneca/source/util/gcp_util"
)

const (
	port = "8080"
)

func main() {
	// Initialize storage client and RawVideoHandler.
	ctx := context.Background()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		fmt.Fprintf(os.Stderr, "GOOGLE_CLOUD_PROJECT environment variable must be set.\n")
		os.Exit(1)
	}

	gcsc, err := gcp_util.NewGoogleCloudStorageClient(ctx, projectID, time.Second*10, time.Minute)
	if err != nil {
		fmt.Printf("NewGoogleCloudStorageClient() returns - err: %v", err)
		return
	}

	gcsd, err := gcp_util.NewGoogleCloudDatastoreClient(ctx, projectID)
	if err != nil {
		fmt.Printf("NewGoogleCloudDatastoreClient() returns - err: %v", err)
		return
	}

	rawVideoHandler, err := rawvideohandler.NewRawVideoHandler(gcsc, gcsd, "", projectID)
	if err != nil {
		fmt.Printf("NewRawVideoHandler() returns - err: %v", err)
		return
	}

	http.HandleFunc("/rawvideo", rawVideoHandler.HandleRawVideoPostRequest)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
