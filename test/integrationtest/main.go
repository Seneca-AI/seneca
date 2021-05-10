package main

import (
	"fmt"
	"log"
	"seneca/test/integrationtest/syncer"
)

func main() {
	if err := syncer.E2ESyncer("senecacam-staging"); err != nil {
		log.Fatal(fmt.Sprintf("SyncerTest failed: %s", err.Error()))
	}
}
