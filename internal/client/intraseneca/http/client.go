package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	st "seneca/api/type"

	"github.com/golang/protobuf/proto"
)

type Client struct {
	hostName   string
	hostPort   string
	authKey    string
	authToken  string
	httpClient *http.Client
}

func New(hostName, hostPort, authKey, authToken string) *Client {
	return &Client{
		hostName:   hostName,
		hostPort:   hostPort,
		authKey:    authKey,
		authToken:  authToken,
		httpClient: &http.Client{},
	}
}

func (ct *Client) ListTrips(req *st.TripListRequest) (*st.TripListResponse, error) {
	buffer := &bytes.Buffer{}
	proto.MarshalText(buffer, req)

	client := &http.Client{}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%s/users/%s/trips", ct.hostName, ct.hostPort, req.UserId), buffer)
	if err != nil {
		return nil, fmt.Errorf("error initializing HTTP get request: %w", err)
	}
	httpReq.Header.Add(ct.authKey, ct.authToken)

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
