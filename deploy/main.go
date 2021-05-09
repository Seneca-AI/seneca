// Package cloudfunction defines all of the Google Cloud Functions we run for seneca.
package cloudfunction

import (
	"fmt"
	"net/http"
)

func RunSyncer(w http.ResponseWriter, r *http.Request) {
	url := "http://singleserver.internal.seneca.com.:6060/syncer"
	resp, err := http.Get(url)
	if err != nil {
		defer resp.Body.Close()
		fmt.Printf("GET %q returns err: %v", url, err)
		w.WriteHeader(400)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Response from %q resp: %v", url, resp)
	w.WriteHeader(200)
}
