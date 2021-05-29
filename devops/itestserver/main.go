package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"seneca/internal/authenticator"
)

const (
	port     = "5050"
	endpoint = "/integration_test"
)

var lock = false

func main() {
	http.HandleFunc(endpoint, runIntegrationTest)
	fmt.Printf("Starting server at port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}

func runIntegrationTest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received integration test request")

	if err := authenticator.AuthorizeHTTPRequest(w, r); err != nil {
		return
	}

	if lock {
		w.WriteHeader(400)
		fmt.Fprintf(w, "An integration test is already running.")
		return
	}
	lock = true
	defer func() { lock = false }()

	defer cleanUpRepos()
	output, err := exec.Command("/bin/sh", "itest.sh").Output()
	fmt.Println(string(output))
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error: %v\n", err)
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "SUCCESS")
}

func cleanUpRepos() {
	os.RemoveAll("seneca/")
	os.RemoveAll("common/")
}
