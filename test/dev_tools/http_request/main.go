// Package main is simple used to test making http calls in go.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	st "seneca/api/type"
	"seneca/test/testutil"

	"github.com/golang/protobuf/proto"
)

func main() {

	postBody, _ := proto.Marshal(
		&st.FollowingDistanceForVideoRequest{
			RequestId:             testutil.TestUserID,
			SimpleStorageVideoUrl: "gs://luca-sample-footage/20210416_170736_NF.mp4",
		},
	)

	resp, err := http.Post("http://127.0.0.1:5000/following_distance", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		log.Fatalf("error making request: %v\n", err)
	}
	defer resp.Body.Close()

	lcrsp := &st.FollowingDistanceForVideoResponse{}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading bytes: %v\n", err)
	}

	if err := proto.UnmarshalText(string(bodyBytes), lcrsp); err != nil {
		log.Fatalf("error umarshalling: %v\n", err)
	}

	fmt.Printf("%v\n", lcrsp)
}
