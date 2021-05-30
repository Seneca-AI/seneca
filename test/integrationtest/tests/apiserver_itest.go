package tests

import (
	"context"
	"fmt"
	st "seneca/api/type"
	"seneca/internal/util"
	"seneca/test/integrationtest/testenv"
	"sort"
	"time"
)

func E2EAPIServer(testUserEmail string, testEnv *testenv.TestEnvironment) error {
	defer testEnv.Clean()

	user, err := testEnv.UserDAO.GetUserByEmail(testUserEmail)
	if err != nil {
		return fmt.Errorf("GetUserByEmail(%s) returns err: %w", testUserEmail, err)
	}

	// Create a RawVideo for the source.
	rawVideo := &st.RawVideo{
		UserId:               user.Id,
		CloudStorageFileName: "",
	}
	rawVideo, err = testEnv.RawVideoDAO.InsertUniqueRawVideo(rawVideo)
	if err != nil {
		return fmt.Errorf("InsertUniqueRawVideo() returns err: %w", err)
	}

	tripStart := time.Date(2021, 5, 23, 1, 0, 0, 0, time.UTC)

	wantTrip := &st.Trip{
		StartTimeMs: util.TimeToMilliseconds(tripStart),
		EndTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Minute * 60)),
		Event: []*st.Event{
			{
				EventType:   1,
				Value:       5,
				Severity:    10,
				TimestampMs: util.TimeToMilliseconds(tripStart.Add(time.Minute * 15)),
			},
			{
				EventType:   2,
				Value:       6,
				Severity:    20,
				TimestampMs: util.TimeToMilliseconds(tripStart.Add(time.Minute * 45)),
			},
		},
		DrivingCondition: []*st.DrivingCondition{
			{
				ConditionType: []st.ConditionType{1},
				Severity:      []float64{7},
				StartTimeMs:   util.TimeToMilliseconds(tripStart),
				EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Minute*60)) - 1,
			},
		},
	}

	eventsInternal := []*st.EventInternal{
		{
			UserId:      user.Id,
			EventType:   1,
			Value:       5,
			Severity:    10,
			TimestampMs: util.TimeToMilliseconds(tripStart.Add(time.Minute * 15)),
			Source: &st.Source{
				SourceType: st.Source_RAW_VIDEO,
				SourceId:   rawVideo.Id,
			},
		},
		{
			UserId:      user.Id,
			EventType:   2,
			Value:       6,
			Severity:    20,
			TimestampMs: util.TimeToMilliseconds(tripStart.Add(time.Minute * 45)),
			Source: &st.Source{
				SourceType: st.Source_RAW_VIDEO,
				SourceId:   rawVideo.Id,
			},
		},
	}

	drivingConditionInternal := &st.DrivingConditionInternal{
		UserId:        user.Id,
		ConditionType: 1,
		Severity:      7,
		StartTimeMs:   util.TimeToMilliseconds(tripStart),
		EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Minute * 60)),
		Source: &st.Source{
			SourceType: st.Source_RAW_VIDEO,
			SourceId:   rawVideo.Id,
		},
	}

	for _, ev := range eventsInternal {
		if _, err := testEnv.EventDAO.CreateEvent(context.TODO(), ev); err != nil {
			return fmt.Errorf("CreateEvent() returns err: %w", err)
		}
	}

	if _, err := testEnv.DrivingConditionDAO.CreateDrivingCondition(context.TODO(), drivingConditionInternal); err != nil {
		return fmt.Errorf("CreateDrivingCondition() returns err: %w", err)
	}

	trips, err := testEnv.APIServer.ListTrips(user.Id, tripStart, tripStart.Add(time.Minute*50))
	if err != nil {
		return fmt.Errorf("ListTrips() returns err: %w", err)
	}

	if len(trips) != 1 {
		return fmt.Errorf("want 1 trip from ListTrips(), got %d", len(trips))
	}

	gotTrip := trips[0]
	if len(gotTrip.Event) != 2 {
		return fmt.Errorf("want 2 events from Trip, got %d", len(gotTrip.Event))
	}

	if len(gotTrip.DrivingCondition) != 1 {
		return fmt.Errorf("want 2 drivingConditions from Trip, got %d", len(gotTrip.DrivingCondition))
	}

	sort.Slice(gotTrip.Event, func(i, j int) bool { return gotTrip.Event[i].TimestampMs < gotTrip.Event[j].TimestampMs })

	for i, ev := range wantTrip.Event {
		if ev.TimestampMs != gotTrip.Event[i].TimestampMs {
			return fmt.Errorf("want %v timestamp from event, got %v", util.MillisecondsToTime(ev.TimestampMs), util.MillisecondsToTime(gotTrip.Event[i].TimestampMs))
		}
	}

	if gotTrip.DrivingCondition[0].StartTimeMs != wantTrip.DrivingCondition[0].StartTimeMs || gotTrip.DrivingCondition[0].EndTimeMs != wantTrip.DrivingCondition[0].EndTimeMs {
		return fmt.Errorf(
			"want driving condition between %v and %v, got between %v and %v",
			util.MillisecondsToTime(wantTrip.DrivingCondition[0].StartTimeMs),
			util.MillisecondsToTime(wantTrip.DrivingCondition[0].EndTimeMs),
			util.MillisecondsToTime(gotTrip.DrivingCondition[0].StartTimeMs),
			util.MillisecondsToTime(gotTrip.DrivingCondition[0].EndTimeMs),
		)
	}

	if gotTrip.StartTimeMs != wantTrip.StartTimeMs || gotTrip.EndTimeMs != wantTrip.EndTimeMs {
		return fmt.Errorf(
			"want trip between %v and %v, got between %v and %v",
			util.MillisecondsToTime(wantTrip.StartTimeMs),
			util.MillisecondsToTime(wantTrip.EndTimeMs),
			util.MillisecondsToTime(gotTrip.StartTimeMs),
			util.MillisecondsToTime(gotTrip.EndTimeMs),
		)
	}

	if testEnv.Logger.Failures() > 0 {
		return fmt.Errorf("got %d logging failures", testEnv.Logger.Failures())
	}

	return nil
}
