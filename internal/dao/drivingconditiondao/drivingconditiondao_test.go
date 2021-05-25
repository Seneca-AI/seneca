package drivingconditiondao

import (
	"context"
	"log"
	"math/rand"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/util"
	"seneca/test/testutil"
	"testing"
	"time"
)

func TestCreateDrivingCondition(t *testing.T) {
	userID := testutil.TestUserID
	drivingConditionOne := &st.DrivingConditionInternal{
		UserId:      userID,
		StartTimeMs: util.TimeToMilliseconds(time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)),
		EndTimeMs:   util.TimeToMilliseconds(time.Date(2021, time.January, 15, 0, 0, 0, 0, time.UTC)),
	}
	drivingConditionTwo := &st.DrivingConditionInternal{
		UserId:      userID,
		StartTimeMs: util.TimeToMilliseconds(time.Date(2021, time.January, 15, 0, 0, 0, 0, time.UTC)),
		EndTimeMs:   util.TimeToMilliseconds(time.Date(2021, time.January, 20, 0, 0, 0, 0, time.UTC)),
	}
	drivingConditionThree := &st.DrivingConditionInternal{
		UserId:      userID,
		StartTimeMs: util.TimeToMilliseconds(time.Date(2021, time.January, 20, 0, 0, 0, 0, time.UTC)),
		EndTimeMs:   util.TimeToMilliseconds(time.Date(2021, time.January, 25, 0, 0, 0, 0, time.UTC)),
	}

	drivingConditionDAO, tripDAO, eventDAO, _ := newDrivingConditionDAOForTest()

	drivingConditionOneWithID, err := drivingConditionDAO.CreateDrivingCondition(context.TODO(), drivingConditionOne)
	if err != nil {
		t.Fatalf("CreateDrivingCondition() returns err: %v", err)
	}

	drivingConditionThreeWithID, err := drivingConditionDAO.CreateDrivingCondition(context.TODO(), drivingConditionThree)
	if err != nil {
		t.Fatalf("CreateDrivingCondition() returns err: %v", err)
	}

	// Make sure trips are different.
	if drivingConditionOneWithID.TripId == drivingConditionThreeWithID.TripId {
		t.Fatalf("Expected different trip IDs for driving conditions created at separate time intervals, but both are %q", drivingConditionOneWithID.TripId)
	}

	// Insert a bunch of random events.
	for i := 0; i < 20; i++ {
		event := &st.EventInternal{
			UserId:      userID,
			TimestampMs: util.TimeToMilliseconds(time.Date(2021, time.January, (10 + rand.Intn(15)), 0, 0, 0, 0, time.UTC)),
		}
		if _, err := eventDAO.CreateEvent(context.TODO(), event); err != nil {
			log.Fatalf("CreateEvent returns err: %v", err)
		}
	}

	// Bridge the gap.
	drivingConditionTwoWithID, err := drivingConditionDAO.CreateDrivingCondition(context.TODO(), drivingConditionTwo)
	if err != nil {
		t.Fatalf("CreateDrivingCondition() returns err: %v", err)
	}

	// Expect one trip.
	tripIDs, err := tripDAO.ListUserTripIDs(userID)
	if err != nil {
		t.Fatalf("ListUserTripIDs() returns err: %v", err)
	}
	if len(tripIDs) != 1 {
		t.Fatalf("Want 1 for tripIDs length, got %d", len(tripIDs))
	}

	// Expect 3 drivingCondition periods.
	drivingConditionIDs, err := drivingConditionDAO.ListTripDrivingConditionIDs(userID, drivingConditionTwoWithID.TripId)
	if err != nil {
		t.Fatalf("ListTripDrivingConditionIDs() returns err: %v", err)
	}
	if len(drivingConditionIDs) != 3 {
		t.Fatalf("Want 3 for drivingConditionIDs length, got %d", len(drivingConditionIDs))
	}

	// Expect 20 events.
	eventIDs, err := eventDAO.ListTripEventIDs(userID, drivingConditionTwoWithID.TripId)
	if err != nil {
		t.Fatalf("ListTripEventIDs() returns err: %v", err)
	}
	if len(eventIDs) != 20 {
		t.Fatalf("Want 20 for eventIDs length, got %d", len(eventIDs))
	}
}

func newDrivingConditionDAOForTest() (*SQLDrivingConditionDAO, *tripdao.SQLTripDAO, *eventdao.SQLEventDAO, *database.FakeSQLDBService) {
	fakeSQLService := database.NewFake()
	logger := logging.NewLocalLogger(false)
	tripDAO := tripdao.NewSQLTripDAO(fakeSQLService, logger)
	eventDAO := eventdao.NewSQLEventDAO(fakeSQLService, tripDAO, logger)

	return NewSQLDrivingConditionDAO(fakeSQLService, tripDAO, eventDAO), tripDAO, eventDAO, fakeSQLService
}
