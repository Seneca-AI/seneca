package apiserver

import (
	"context"
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/controller/apiserver"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/dataaggregator/sanitizer"
	"seneca/internal/util"
	"sort"
	"time"
)

func E2EAPIServer(projectID, testUserEmail string, sqlService database.SQLInterface, userDAO dao.UserDAO, logger logging.LoggingInterface) error {
	user, err := userDAO.GetUserByEmail(testUserEmail)
	if err != nil {
		return fmt.Errorf("GetUserByEmail(%s) returns err: %w", testUserEmail, err)
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
		},
		{
			UserId:      user.Id,
			EventType:   2,
			Value:       6,
			Severity:    20,
			TimestampMs: util.TimeToMilliseconds(tripStart.Add(time.Minute * 45)),
		},
	}

	drivingConditionInternal := &st.DrivingConditionInternal{
		UserId:        user.Id,
		ConditionType: 1,
		Severity:      7,
		StartTimeMs:   util.TimeToMilliseconds(tripStart),
		EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Minute * 60)),
	}

	tripDAO := tripdao.NewSQLTripDAO(sqlService, logger)
	eventDAO := eventdao.NewSQLEventDAO(sqlService, tripDAO)
	drivingConditionDAO := drivingconditiondao.NewSQLDrivingConditionDAO(sqlService, tripDAO, eventDAO)
	sanitizer := sanitizer.New(eventDAO, drivingConditionDAO)
	apiserver := apiserver.New(sanitizer, tripDAO)

	createdEventIDs := []string{}
	createdDrivingConditionID := ""
	createdTripID := ""

	for _, ev := range eventsInternal {
		createdEvent, err := eventDAO.CreateEvent(context.TODO(), ev)
		if err != nil {
			return fmt.Errorf("CreateEvent() returns err: %w", err)
		}
		createdEventIDs = append(createdEventIDs, createdEvent.Id)
	}

	createdDrivingCondition, err := drivingConditionDAO.CreateDrivingCondition(context.TODO(), drivingConditionInternal)
	if err != nil {
		return fmt.Errorf("CreateDrivingCondition() returns err: %w", err)
	}
	createdDrivingConditionID = createdDrivingCondition.Id
	createdTripID = createdDrivingCondition.TripId

	// Cleanup.
	defer func() {
		for _, eid := range createdEventIDs {
			if err := eventDAO.DeleteEventByID(context.TODO(), user.Id, createdTripID, eid); err != nil {
				logger.Error(fmt.Sprintf("DeleteEventByID(%s) returns err: %v", eid, err))
			}
		}

		if err := drivingConditionDAO.DeleteDrivingConditionByID(context.TODO(), user.Id, createdTripID, createdDrivingConditionID); err != nil {
			logger.Error(fmt.Sprintf("DeleteDrivingConditionByID(%s) returns err: %v", createdDrivingConditionID, err))
		}

		if err := tripDAO.DeleteTripByID(context.TODO(), createdTripID); err != nil {
			logger.Error(fmt.Sprintf("DeleteTripByID(%s) returns err: %v", createdTripID, err))
		}
	}()

	trips, err := apiserver.ListTrips(user.Id, tripStart, tripStart.Add(time.Minute*50))
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

	return nil
}
