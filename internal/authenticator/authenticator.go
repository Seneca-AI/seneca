package authenticator

import (
	"fmt"
	"net/http"
	"seneca/api/constants"
)

const authHeaderKey = "Authorization"

func AuthorizeHTTPRequest(w http.ResponseWriter, r *http.Request) error {
	auth := r.Header.Get(authHeaderKey)
	if auth != constants.SenecaAPIKey {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Invalid auth token %q\n", auth)
		return fmt.Errorf("invalid token")
	}
	return nil
}

func AddRequestAuth(req *http.Request) *http.Request {
	req.Header = http.Header{
		authHeaderKey: []string{constants.SenecaAPIKey},
	}
	return req
}
