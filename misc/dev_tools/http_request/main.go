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

	// Create 24 event requests.
	for i := 0; i < 24; i++ {
		eventCreateRequest := &st.EventCreateRequest{
			UserId: "5685335367352320",
			Event: &st.EventInternal{
				UserId:      "5685335367352320",
				EventType:   st.EventType(1 + rand.Intn(3)),
				Severity:    float64(rand.Intn(10)),
				TimestampMs: util.TimeToMilliseconds(tripStart.Add(time.Duration(rand.Intn(60) * int(time.Minute)))),
				Source: &st.Source{
					SourceType: st.Source_RAW_VIDEO,
					SourceId:   "5929195020484608",
				},
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
			UserId: "5685335367352320",
			DrivingCondition: &st.DrivingConditionInternal{
				UserId:        "5685335367352320",
				ConditionType: st.ConditionType(rand.Intn(9)),
				Severity:      float64(rand.Intn(10)),
				StartTimeMs:   util.TimeToMilliseconds(startTime),
				EndTimeMs:     util.TimeToMilliseconds(endTime),
				Source: &st.Source{
					SourceType: st.Source_RAW_VIDEO,
					SourceId:   "5929195020484608",
				},
			},
		}

		err := sendHTTPDrivingConditionCreateRequest(dcCreateRequest)
		if err != nil {
			log.Fatalf("Error sending DrivingConditionCreateRequest: %v - err: %v", dcCreateRequest, err)
		}
	}

	// And one to merge them all together.
	dcCreateRequest := &st.DrivingConditionCreateRequest{
		UserId: "5685335367352320",
		DrivingCondition: &st.DrivingConditionInternal{
			UserId:        "5685335367352320",
			ConditionType: st.ConditionType_NONE_CONDITION_TYPE,
			Severity:      float64(rand.Intn(10)),
			StartTimeMs:   util.TimeToMilliseconds(tripStart),
			EndTimeMs:     util.TimeToMilliseconds(tripEnd),
			Source: &st.Source{
				SourceType: st.Source_RAW_VIDEO,
				SourceId:   "5929195020484608",
			},
		},
	}
	err := sendHTTPDrivingConditionCreateRequest(dcCreateRequest)
	if err != nil {
		log.Fatalf("Error sending DrivingConditionCreateRequest: %v - err: %v", dcCreateRequest, err)
	}

	// Then get the trip.
	tripRequest := &st.TripListRequest{
		UserId:      "5685335367352320",
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
