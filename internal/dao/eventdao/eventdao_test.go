package eventdao_test

import (
	"context"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/util"
	"seneca/test/testutil"
	"testing"
	"time"
)

var timestamp = time.Date(1996, time.May, 23, 0, 0, 0, 0, time.UTC)

func TestCreateEvent(t *testing.T) {
	userID := testutil.TestUserID
	eventOne := &st.EventInternal{
		UserId:      userID,
		EventType:   st.EventType_LANE_CHANGE,
		TimestampMs: util.TimeToMilliseconds(timestamp),
	}

	eventTwo := &st.EventInternal{
		UserId:      userID,
		EventType:   st.EventType_FAST_ACCELERATION,
		TimestampMs: util.TimeToMilliseconds(timestamp),
	}

	eventDAO, tripDAO, _ := newEventDAOForTest()

	// Verify the first event creation also creates a new trip.
	eventOneWithID, err := eventDAO.CreateEvent(context.TODO(), eventOne)
	if err != nil {
		t.Fatalf("CreateEvent(one) returns err: %v", err)
	}
	trip, err := tripDAO.GetTripByID(eventOneWithID.UserId, eventOneWithID.TripId)
	if err != nil {
		t.Fatalf("GetTripByID() returns err: %v", err)
	}
	if trip == nil {
		t.Fatalf("Trip is nil")
	}

	eventTwoWithID, err := eventDAO.CreateEvent(context.TODO(), eventTwo)
	if err != nil {
		t.Fatalf("CreateEvent(two) returns err: %v", err)
	}

	if eventOneWithID.TripId != eventTwoWithID.TripId {
		t.Fatalf("Trip IDs not equal though event have same timestamp.")
	}

	// User should only have one trip.
	tripIDs, err := tripDAO.ListUserTripIDs(userID)
	if err != nil {
		t.Fatalf("ListUserTripIDs() returns err: %v", err)
	}

	if len(tripIDs) != 1 {
		t.Fatalf("Want 1 trip IDs, got %d", len(tripIDs))
	}
}

func TestCreateEventWhenTripAlreadyExists(t *testing.T) {
	eventDAO, tripDAO, drivingConditionDAO := newEventDAOForTest()

	startTime := time.Date(2021, 05, 25, 0, 0, 0, 0, time.UTC)

	drivingCondition := &st.DrivingConditionInternal{
		UserId:      testutil.TestUserID,
		StartTimeMs: util.TimeToMilliseconds(startTime),
		EndTimeMs:   util.TimeToMilliseconds(startTime.Add(time.Hour)),
	}

	if _, err := drivingConditionDAO.CreateDrivingCondition(context.TODO(), drivingCondition); err != nil {
		t.Fatalf("CreateDrivingCondition() returns err: %v", err)
	}

	for i := 0; i < 60; i += 5 {
		event := &st.EventInternal{
			UserId:      testutil.TestUserID,
			TimestampMs: util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(i))),
		}

		if _, err := eventDAO.CreateEvent(context.TODO(), event); err != nil {
			t.Fatalf("CreateEvent() returns err: %v", err)
		}
	}

	tripIDs, err := tripDAO.ListUserTripIDs(testutil.TestUserID)
	if err != nil {
		t.Fatalf("ListUserTripIDs() returns err: %v", err)
	}
	if len(tripIDs) != 1 {
		t.Fatalf("Want 1 trip ID, got %d", len(tripIDs))
	}

	eventIDs, err := eventDAO.ListTripEventIDs(testutil.TestUserID, tripIDs[0])
	if err != nil {
		t.Fatalf("ListTripEventIDs() returns err: %v", err)
	}

	if len(eventIDs) != 12 {
		t.Fatalf("Want 12 event IDs, got %d", len(eventIDs))
	}
}

func TestGetListEventByID(t *testing.T) {
	event := &st.EventInternal{
		UserId:      testutil.TestUserID,
		EventType:   st.EventType_LANE_CHANGE,
		TimestampMs: util.TimeToMilliseconds(timestamp),
	}

	eventDAO, _, _ := newEventDAOForTest()

	eventWithID, err := eventDAO.CreateEvent(context.TODO(), event)
	if err != nil {
		t.Fatalf("CreateEvent() returns err: %v", err)
	}

	if _, err := eventDAO.GetEventByID(eventWithID.UserId, eventWithID.TripId, eventWithID.Id); err != nil {
		t.Fatalf("GetEventByID() returns err: %v", err)
	}

	eventIDs, err := eventDAO.ListTripEventIDs(eventWithID.UserId, eventWithID.TripId)
	if err != nil {
		t.Fatalf("ListTripEventIDs() returns err: %v", err)
	}

	if len(eventIDs) != 1 {
		t.Fatalf("Want eventIDs of length 1, got %d", len(eventIDs))
	}

	if err := eventDAO.DeleteEventByID(context.TODO(), eventWithID.UserId, eventWithID.TripId, eventWithID.Id); err != nil {
		t.Fatalf("DeleteEventByID() returns err: %v", err)
	}
}

func newEventDAOForTest() (*eventdao.SQLEventDAO, dao.TripDAO, dao.DrivingConditionDAO) {
	fakeSQLService := database.NewFake()
	logger := logging.NewLocalLogger(false)
	tripDAO := tripdao.NewSQLTripDAO(fakeSQLService, logger)
	eventDAO := eventdao.NewSQLEventDAO(fakeSQLService, tripDAO, logger)
	drivingConditionDAO := drivingconditiondao.NewSQLDrivingConditionDAO(fakeSQLService, tripDAO, eventDAO)

	return eventDAO, tripDAO, drivingConditionDAO
}
