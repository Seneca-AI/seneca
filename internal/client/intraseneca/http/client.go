package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/internal/authenticator"
	"seneca/internal/client/intraseneca"

	"github.com/golang/protobuf/proto"
)

const (
	objectInVideoEndpoint = "/objects_in_frame"
)

type Client struct {
	serverConfig           *intraseneca.ServerConfig
	senecaServerHTTPClient *http.Client
	mlServerHTTPClient     *http.Client
}

func New(config *intraseneca.ServerConfig) (*Client, error) {
	var senecaServerHTTPClient *http.Client
	if config.SenecaServerHostName != "" {
		senecaServerHTTPClient = &http.Client{
			Timeout: config.SenecaServerTimeout,
		}
	}

	var mlServerHTTPClient *http.Client
	if config.MLServerHostName != "" {
		mlServerHTTPClient = &http.Client{
			Timeout: config.MLServerTimeout,
		}
	}

	// Send a hearbeat message to each of the hosts.
	if senecaServerHTTPClient != nil {
		if err := sendHeartBeat(config.SenecaServerHostName, config.SenecaServerHostPort, senecaServerHTTPClient); err != nil {
			return nil, fmt.Errorf("SenecaServer failed heartbeat: %w", err)
		}
	}
	if mlServerHTTPClient != nil {
		if err := sendHeartBeat(config.MLServerHostName, config.MLServerHostPort, mlServerHTTPClient); err != nil {
			return nil, fmt.Errorf("MLServer failed heartbeat: %w", err)
		}
	}

	return &Client{
		serverConfig:           config,
		senecaServerHTTPClient: senecaServerHTTPClient,
		mlServerHTTPClient:     mlServerHTTPClient,
	}, nil
}

func (ct *Client) ListTrips(req *st.TripListRequest) (*st.TripListResponse, error) {
	buffer := &bytes.Buffer{}
	proto.MarshalText(buffer, req)

	client := &http.Client{}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%s/users/%s/trips", ct.serverConfig.SenecaServerHostName, ct.serverConfig.SenecaServerHostPort, req.UserId), buffer)
	if err != nil {
		return nil, fmt.Errorf("error initializing HTTP get request: %w", err)
	}

	httpReq = authenticator.AddRequestAuth(httpReq)

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d with message %q", resp.StatusCode, string(bodyBytes))
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

	httpReq = authenticator.AddRequestAuth(httpReq)

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

func sendHeartBeat(hostname, port string, httpClient *http.Client) error {
	// TODO(lucaloncar): define http/https protocol as a type
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s/%s", hostname, port, constants.HeartbeatEndpoint), nil)
	if err != nil {
		return fmt.Errorf("error initializing request: %w", err)
	}

	request = authenticator.AddRequestAuth(request)

	resp, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("server responded with status %d and message was unparseable", resp.StatusCode)
		}
		return fmt.Errorf("server responded with status %d and message %q", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
