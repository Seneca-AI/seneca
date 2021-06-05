package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	st "seneca/api/type"
	"seneca/internal/client/intraseneca"

	"github.com/golang/protobuf/proto"
)

const (
	objectInVideoEndpoint = "/objects_in_video"
)

type Client struct {
	serverConfig           *intraseneca.ServerConfig
	senecaServerHTTPClient *http.Client
	mlServerHTTPClient     *http.Client
}

func New(config *intraseneca.ServerConfig) *Client {
	return &Client{
		serverConfig: config,
		senecaServerHTTPClient: &http.Client{
			Timeout: config.SenecaServerTimeout,
		},
		mlServerHTTPClient: &http.Client{
			Timeout: config.MLServerTimeout,
		},
	}
}

func (ct *Client) ListTrips(req *st.TripListRequest) (*st.TripListResponse, error) {
	buffer := &bytes.Buffer{}
	proto.MarshalText(buffer, req)

	client := &http.Client{}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%s/users/%s/trips", ct.serverConfig.SenecaServerHostName, ct.serverConfig.SenecaServerHostPort, req.UserId), buffer)
	if err != nil {
		return nil, fmt.Errorf("error initializing HTTP get request: %w", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	defer resp.Body.Close()

	respProto := &st.TripListResponse{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading bytes: %w", err)
	}
	if err := proto.UnmarshalText(string(bodyBytes), respProto); err != nil {
		responseMessage := string(bodyBytes)
		return nil, fmt.Errorf("error unmarshalling response: %w - string message: %q", err, responseMessage)
	}

	return respProto, nil
}

func (ct *Client) ProcessObjectsInVideo(req *st.ObjectsInFrameRequest) (*st.ObjectsInFrameResponse, error) {
	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s/%s", ct.serverConfig.MLServerHostName, ct.serverConfig.MLServerHostPort, objectInVideoEndpoint), bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error initializing HTTP get request: %w", err)
	}

	resp, err := ct.mlServerHTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respProto := &st.ObjectsInFrameResponse{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading bytes: %w", err)
	}

	responseMessage := string(bodyBytes)
	if err := proto.UnmarshalText(responseMessage, respProto); err != nil {
		log.Fatalf("Error unmarshalling response: %v- string message: %q", err, responseMessage)
	}

	return respProto, nil
}
