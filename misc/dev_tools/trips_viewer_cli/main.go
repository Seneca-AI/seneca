// This CLI allows you to query APIServer.ListTrips, and formats the output to make the Trips more readable.
package main

import (
	"fmt"
	"log"
	"os"
	st "seneca/api/type"
	"seneca/internal/client/intraseneca"
	"seneca/internal/client/intraseneca/http"
	"seneca/internal/util"
	"sort"
	"strconv"
	"strings"
	"time"
)

const labels = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var screenWidth = 0

func main() {
	config := &intraseneca.ServerConfig{
		SenecaServerHostName: "127.0.0.1",
		SenecaServerHostPort: "6060",
	}

	intraSenecaClient, err := http.New(config)
	if err != nil {
		log.Fatalf("intraseneca.http.New() returns err: %v", err)
	}

	for {
		var err error

		fmt.Println("At any point, enter -1 to exit.")
		fmt.Println("Enter the width of you terminal in characters so we can correctly pretty print trips:")
		widthString := scanOrExit()
		screenWidth, err = strconv.Atoi(widthString)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println("Enter the ID of the user you'd like to get Trip data for:")

		userID := scanOrExit()

		fmt.Println("Enter the start time of the query in the form YYYY-MM-DD hh:mm:ss - you may omit parts of the time:")
		startTimeString := scanOrExit()
		startTime, err := parseTime(startTimeString)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println("Enter the end time of the query in the form YYYY-MM-DD hh:mm:ss - you may omit parts of the time:")
		endTimeString := scanOrExit()
		endTime, err := parseTime(endTimeString)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		req := &st.TripListRequest{
			UserId:      userID,
			StartTimeMs: util.TimeToMilliseconds(startTime),
			EndTimeMs:   util.TimeToMilliseconds(endTime),
		}

	listTrips:
		resp, err := intraSenecaClient.ListTrips(req)
		if err != nil {
			fmt.Printf("Error listing trips: %v\n", err)
		}

		trips := printTripList(resp)
		fmt.Println("Choose a trip from the list above to get more info about, or -1 to exit.")

	selectTrip:

		selection := scanOrExit()
		selectionInt, err := strconv.Atoi(selection)
		if err != nil {
			fmt.Printf("Invalid selection string %q leads to err: %v\n", selection, err)
			goto selectTrip
		}

		events, drivingConditions := prettyPrintTrip(trips[selectionInt])
	selectEvent:
		fmt.Println("Choose a part of this trip to get more info on.  Prefix events with 'E' and drivingConditoins with 'C'. Or press 0 to view all trips again, or -1 to exit.")
		selection = scanOrExit()
		if selection[0:1] == "0" {
			goto listTrips
		}

		if len(selection) != 2 {
			fmt.Printf("Invalid selection string %q\n", selection)
			goto selectEvent
		}
		if selection[0:1] == "E" {
			fmt.Printf("Event: %s\n", util.EventExternalToPrettyString(events[selection[1:2]]))
		} else if selection[0:1] == "C" {
			fmt.Printf("DrivingCondition: %s\n", util.DrivingConditionExternalToPrettyString(drivingConditions[selection[1:2]]))
		} else {
			fmt.Printf("Invalid selection string %q\n", selection)
		}

		goto selectEvent
	}
}

func prettyPrintTrip(trip *st.Trip) (map[string]*st.Event, map[string]*st.DrivingCondition) {
	if len(trip.Event) > 26 || len(trip.DrivingCondition) > 26 {
		log.Fatalf("Tool not able to handle trip with more than 26 events or driving conditions.")
	}

	// Assign each event a number.
	eventLabel := 0
	eventsMap := map[string]*st.Event{}
	for _, ev := range trip.Event {
		eventsMap[toChar(eventLabel)] = ev
		eventLabel++
	}

	timeMonad := int64((trip.EndTimeMs - trip.StartTimeMs) / int64(screenWidth))
	eventBuckets := [][]string{}
	for i := int64(0); i < int64(screenWidth); i++ {
		eventBuckets = append(eventBuckets, []string{})
	}

	largestBucket := 0
	for k, ev := range eventsMap {
		for i := int64(0); i < int64(screenWidth); i++ {
			if ev.TimestampMs >= (trip.StartTimeMs+i*timeMonad) && ev.TimestampMs < (trip.StartTimeMs+(i+1)*timeMonad) {
				eventBuckets[i] = append(eventBuckets[i], k)
				if len(eventBuckets[i]) > largestBucket {
					largestBucket = len(eventBuckets[i])
				}
				break
			}
		}
	}

	// Print events.
	bucketsLine := largestBucket
	for bucketsLine > 0 {
		for _, bucket := range eventBuckets {
			if len(bucket) >= bucketsLine {
				fmt.Print(bucket[bucketsLine-1])
			} else {
				fmt.Print(" ")
			}
		}
		bucketsLine--
		fmt.Printf("\n")
	}

	// Print drivingConditions.
	dcLabel := 25
	dcMap := map[string]*st.DrivingCondition{}
	for _, dc := range trip.DrivingCondition {
		dcMap[toChar(dcLabel)] = dc
		dcLabel--
	}

	for i := int64(0); i < int64(screenWidth); i++ {
		for k, dc := range dcMap {
			if dc.StartTimeMs <= (trip.StartTimeMs+i*timeMonad) && dc.EndTimeMs >= (trip.StartTimeMs+i*timeMonad) {
				fmt.Print(k)
			}
		}
	}
	fmt.Printf("\n")

	return eventsMap, dcMap
}

func toChar(i int) string {
	return labels[i : i+1]
}

func printTripList(resp *st.TripListResponse) map[int]*st.Trip {
	if len(resp.Trip) == 0 {
		fmt.Println("There were no trips for the user between those times.")
		return nil
	}
	sort.Slice(resp.Trip, func(i, j int) bool { return resp.Trip[i].StartTimeMs < resp.Trip[j].StartTimeMs })
	tripsMap := map[int]*st.Trip{}

	for i := 0; i < len(resp.Trip); i++ {
		tripsMap[i] = resp.Trip[i]
	}

	for i := 0; i < len(tripsMap); i++ {
		fmt.Printf("|%6d|: %v - %v\n", i, util.MillisecondsToTime(tripsMap[i].StartTimeMs), util.MillisecondsToTime(tripsMap[i].EndTimeMs))
	}

	return tripsMap
}

func parseTime(timeString string) (time.Time, error) {
	dateAndTime := strings.Split(timeString, " ")
	if len(dateAndTime) != 1 && len(dateAndTime) != 2 {
		return time.Now(), fmt.Errorf("invalid input %q", timeString)
	}

	dateParts := strings.Split(dateAndTime[0], "-")
	if len(dateParts) != 3 {
		return time.Now(), fmt.Errorf("invalid input %q", timeString)
	}

	year, err := strconv.Atoi(dateParts[0])
	if err != nil {
		return time.Now(), fmt.Errorf("error parsing year for %q - err: %w", timeString, err)
	}
	month, err := strconv.Atoi(dateParts[1])
	if err != nil {
		return time.Now(), fmt.Errorf("error parsing month for %q - err: %w", timeString, err)
	}
	day, err := strconv.Atoi(dateParts[2])
	if err != nil {
		return time.Now(), fmt.Errorf("error parsing day for %q - err: %w", timeString, err)
	}

	hour := 0
	minute := 0
	second := 0
	if len(dateAndTime) == 2 {
		timeParts := strings.Split(dateAndTime[1], ":")
		if len(timeParts) > 3 {
			return time.Now(), fmt.Errorf("invalid input %q", timeString)
		}

		hour, err = strconv.Atoi(timeParts[0])
		if err != nil {
			return time.Now(), fmt.Errorf("error parsing hour for %q - err: %w", timeString, err)
		}
		if len(timeParts) > 1 {
			minute, err = strconv.Atoi(timeParts[1])
			if err != nil {
				return time.Now(), fmt.Errorf("error parsing minute for %q - err: %w", timeString, err)
			}
		}
		if len(timeParts) > 2 {
			second, err = strconv.Atoi(timeParts[2])
			if err != nil {
				return time.Now(), fmt.Errorf("error parsing second for %q - err: %w", timeString, err)
			}
		}
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
}

func scanOrExit() string {
	var inputString string
	fmt.Scanln(&inputString)
	if inputString == "-1" {
		fmt.Println("Goodbye")
		os.Exit(0)
	}
	return inputString
}
