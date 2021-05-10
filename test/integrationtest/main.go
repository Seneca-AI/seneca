package main

import (
	"log"
	"seneca/env"
	"seneca/test/integrationtest/syncer"
)

func main() {
	if err := env.ValidateEnvironmentVariables(); err != nil {
		log.Fatalf("Failed to ValidateEnvironmentVariables: %s", err.Error())
	}

	if err := syncer.E2ESyncer("senecacam-staging"); err != nil {
		log.Fatalf("SyncerTest failed: %s", err.Error())
	}
}
