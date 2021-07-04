package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	st "seneca/api/type"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/dao/rawframedao"

	"github.com/golang/protobuf/proto"
)

const (
	mlServerHostname = "10.128.0.28"
	mlServerPort     = "5000"
	mlServerEndpoint = "objects_in_video"
)

func main() {
	sqlService, err := datastore.New(context.TODO(), "senecacam-sandbox")
	if err != nil {
		log.Fatalf("Error datastore.New(): %v", err)
	}

	rawFrameDAO := rawframedao.NewSQLRawFrameDAO(sqlService)

	rawFrame, err := rawFrameDAO.GetRawFrameByID("4540225682407424")
	if err != nil {
		log.Fatalf("Error GetRawFrameByID(): %v", err)
	}

	objectsInVideoRequest := &st.ObjectsInFrameRequest{
		RawFrame: rawFrame,
	}

	data, err := proto.Marshal(objectsInVideoRequest)
	if err != nil {
		log.Fatalf("Error marshalling request: %v", err)
	}

	client := &http.Client{}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s/%s", mlServerHostname, mlServerPort, mlServerEndpoint), bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error initializing HTTP get request: %v", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	defer resp.Body.Close()

	respProto := &st.ObjectsInFrameResponse{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading bytes: %v", err)
	}

	responseMessage := string(bodyBytes)

	if err := proto.UnmarshalText(responseMessage, respProto); err != nil {
		log.Fatalf("Error unmarshalling response: %v- string message: %q", err, responseMessage)
	}

	fmt.Printf("Successful response: %v\n", respProto)
}
