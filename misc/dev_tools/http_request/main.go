// Package main is simple used to test making http calls in go.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	st "seneca/api/type"
	"seneca/internal/util"
	"time"

	"github.com/golang/protobuf/proto"
)

func main() {
	tripStart := time.Date(2021, 05, 16, 0, 0, 0, 0, time.UTC)
	tripEnd := time.Date(2021, 05, 16, 1, 0, 0, 0, time.UTC)

	// Create 50 event requests.
	for i := 0; i < 50; i++ {
		eventCreateRequest := &st.EventCreateRequest{
			UserId: "123",
			Event: &st.EventInternal{
				UserId:      "123",
				EventType:   st.EventType(1 + rand.Intn(3)),
				Severity:    float64(rand.Intn(10)),
				TimestampMs: util.TimeToMilliseconds(tripStart.Add(time.Duration(rand.Intn(60) * int(time.Minute)))),
			},
		}
		err := sendHTTPEventCreateRequest(eventCreateRequest)
		if err != nil {
			log.Fatalf("Error sending EventCreateRequest: %v - err: %v", eventCreateRequest, err)
		}
	}

	// and 10 DrivingConditions
	for i := 0; i < 10; i++ {
		startTime := tripStart.Add(time.Duration(rand.Intn(60) * int(time.Minute)))
		untilTripEnd := tripEnd.Sub(startTime)
		endTime := startTime.Add(time.Duration(rand.Int63n(untilTripEnd.Nanoseconds())))

		dcCreateRequest := &st.DrivingConditionCreateRequest{
			UserId: "123",
			DrivingCondition: &st.DrivingConditionInternal{
				UserId:        "123",
				ConditionType: st.ConditionType(rand.Intn(9)),
				Severity:      float64(rand.Intn(10)),
				StartTimeMs:   util.TimeToMilliseconds(startTime),
				EndTimeMs:     util.TimeToMilliseconds(endTime),
			},
		}
		err := sendHTTPDrivingConditionCreateRequest(dcCreateRequest)
		if err != nil {
			log.Fatalf("Error sending DrivingConditionCreateRequest: %v - err: %v", dcCreateRequest, err)
		}
	}

	// Then get the trip.
	tripRequest := &st.TripListRequest{
		UserId:      "123",
		StartTimeMs: util.TimeToMilliseconds(tripStart) - 1000,
		EndTimeMs:   util.TimeToMilliseconds(tripEnd) + 1000,
	}

	if err := sendHTTPTripListRequest(tripRequest); err != nil {
		log.Fatalf("Error sending TripListRequest: %v - err: %v", tripRequest, err)
	}
}

func sendHTTPTripListRequest(req *st.TripListRequest) error {
	buffer := &bytes.Buffer{}
	proto.MarshalText(buffer, req)

	client := &http.Client{}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:6060/users/%s/trips", req.UserId), buffer)
	if err != nil {
		return fmt.Errorf("error initializing HTTP get request: %w", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}

	defer resp.Body.Close()

	respProto := &st.TripListResponse{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading bytes: %w", err)
	}
	if err := proto.UnmarshalText(string(bodyBytes), respProto); err != nil {
		responseMessage := string(bodyBytes)
		return fmt.Errorf("error unmarshalling response: %w - string message: %q", err, responseMessage)
	}

	fmt.Printf("%v\n", respProto)
	return nil
}

func sendHTTPEventCreateRequest(req *st.EventCreateRequest) error {
	buffer := &bytes.Buffer{}
	proto.MarshalText(buffer, req)

	resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:6060/users/%s/events", req.UserId), "application/json", buffer)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respProto := &st.EventCreateResponse{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading bytes: %w", err)
	}
	if err := proto.UnmarshalText(string(bodyBytes), respProto); err != nil {
		responseMessage := string(bodyBytes)
		return fmt.Errorf("error unmarshalling response: %w - string message: %q", err, responseMessage)
	}

	fmt.Printf("%v\n", respProto)
	return nil
}

func sendHTTPDrivingConditionCreateRequest(req *st.DrivingConditionCreateRequest) error {
	buffer := &bytes.Buffer{}
	proto.MarshalText(buffer, req)

	resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:6060/users/%s/driving_conditions", req.UserId), "application/json", buffer)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respProto := &st.DrivingConditionCreateResponse{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading bytes: %w", err)
	}
	if err := proto.UnmarshalText(string(bodyBytes), respProto); err != nil {
		responseMessage := string(bodyBytes)
		return fmt.Errorf("error unmarshalling response: %w - string message: %q", err, responseMessage)
	}

	fmt.Printf("%v\n", respProto)
	return nil
}
