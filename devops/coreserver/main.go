package main

// TODO(lucaloncar): make all of the code in this directory better

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"seneca/api/constants"
	"seneca/devops/coreserver/helpers"
	"seneca/internal/authenticator"
	"seneca/internal/client/logging"
)

const (
	projectID = "senecacam-core"
	port      = "5050"

	integrationTestEndpoint        = "integration_test"
	integrationTestServerPort      = "5050"
	integrationTestServerIPAddress = "35.188.190.79"

	pushEndpoint = "push_main"
)

type httpHandler struct {
	medic  *helpers.Medic
	pusher *helpers.Pusher
	logger logging.LoggingInterface
}

func main() {
	logger, err := logging.NewGCPLogger(context.TODO(), "coreserver", projectID)
	if err != nil {
		log.Fatalf("logging.NewGCPLogger() returns err: %v", err)
	}

	medic, err := helpers.NewMedic(projectID, logger)
	if err != nil {
		logger.Critical(fmt.Sprintf("helpers.NewMedic() returns err: %v", err))
		return
	}

	pusher, err := helpers.NewPusher(projectID)
	if err != nil {
		logger.Critical(fmt.Sprintf("helpers.NewPusher() returns err: %v", err))
		return
	}

	hndlr := &httpHandler{
		medic:  medic,
		pusher: pusher,
		logger: logger,
	}

	http.HandleFunc("/", hndlr.handleHTTPRequest)
	fmt.Printf("Starting server at port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}

func (hndl *httpHandler) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received HTTP request to %q\n", r.URL.Path)

	if err := authenticator.AuthorizeHTTPRequest(w, r); err != nil {
		return
	}

	if r.URL.Path == fmt.Sprintf("/%s", integrationTestEndpoint) {
		handleIntegrationTestRequest(w, r)
	} else if r.URL.Path == fmt.Sprintf("/%s", constants.HeartbeatEndpoint) {
		if err := hndl.medic.Run(); err != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	} else if r.URL.Path == fmt.Sprintf("/%s", pushEndpoint) {
		if err := hndl.pusher.Run(); err != nil {
			hndl.logger.Error(fmt.Sprintf("Error running pusher: %v", err))
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	} else {
		fmt.Fprintf(w, "Unsupported path")
		w.WriteHeader(400)
	}
}

func handleIntegrationTestRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received integration test request")

	httpClient := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s/%s", integrationTestServerIPAddress, integrationTestServerPort, integrationTestEndpoint), nil)
	if err != nil {
		fmt.Fprintf(w, "Error initializing request to itest server: %v", err)
		w.WriteHeader(500)
		return
	}

	req = authenticator.AddRequestAuth(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(w, "Error sending request to itest server: %v", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	io.Copy(w, resp.Body)
	resp.Body.Close()
}
