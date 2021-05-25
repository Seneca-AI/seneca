package dataprocessor

import (
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/util"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	tripDAO, eventDAO, drivingConditionDAO, rawMotionDAO, rawVideoDAO, logger := newDataProcessorPartsForTest()
	alg := newAccelerationV0()
	dp := New([]AlgorithmInterface{alg}, eventDAO, drivingConditionDAO, rawMotionDAO, rawVideoDAO, logger)

	// Create a few raw motions.
	for _, acc := range []int{5, 10, 15, 20} {
		rawMotion := &st.RawMotion{
			UserId: "123",
			Motion: &st.Motion{
				AccelerationMphS: float64(acc),
			},
			TimestampMs: util.TimeToMilliseconds(time.Date(2021, 05, 05, 0, 0, acc, 0, time.UTC)),
		}

		if _, err := rawMotionDAO.InsertUniqueRawMotion(rawMotion); err != nil {
			t.Fatalf("InsertUniqueRawMotion() returns err: %v", err)
		}
	}

	// And a raw video to associate.
	rawVideo := &st.RawVideo{
		UserId:       "123",
		CreateTimeMs: util.TimeToMilliseconds(time.Date(2021, 05, 05, 0, 0, 0, 0, time.UTC)),
		DurationMs:   int64(time.Minute * 1),
	}
	if _, err := rawVideoDAO.InsertUniqueRawVideo(rawVideo); err != nil {
		t.Fatalf("InsertUniqueRawVideo() returns err: %v", err)
	}

	dp.Run("123")

	// Verify trip and events.
	tripIDs, err := tripDAO.ListUserTripIDs("123")
	if err != nil {
		t.Fatalf("ListUserTripIDs() returns err: %v", err)
	}
	if len(tripIDs) != 1 {
		t.Fatalf("Want 1 trip ID, got %d", len(tripIDs))
	}

	eventIDs, err := eventDAO.ListTripEventIDs("123", tripIDs[0])
	if err != nil {
		t.Fatalf("ListTripEventIDs() returns err: %v", err)
	}

	if len(eventIDs) != 3 {
		t.Fatalf("Want 3 event ID, got %d", len(eventIDs))
	}

	events := []*st.EventInternal{}
	for _, eid := range eventIDs {
		event, err := eventDAO.GetEventByID("123", tripIDs[0], eid)
		if err != nil {
			t.Fatalf("GetEventByID() returns err: %v", err)
		}
		events = append(events, event)
	}

	// Values don't matter too much, just that fields are non-nil.
	for _, ev := range events {
		if ev.AlgoTag != alg.Tag() {
			t.Fatalf("Want AlgoTag %q for event, got %q", alg.tag, ev.AlgoTag)
		}
		if ev.EventType != st.EventType_FAST_ACCELERATION {
			t.Fatalf("Want EventType %q, got %q", st.EventType_FAST_ACCELERATION.String(), ev.EventType.String())
		}
		if ev.Source.SourceType != st.Source_RAW_MOTION {
			t.Fatalf("Want SourceType %q, got %q", st.Source_RAW_MOTION.String(), ev.Source.SourceType.String())
		}
	}

	// Make sure raws were updated as well.
	unprocessedRawMotionIDs, err := rawMotionDAO.ListUnprocessedRawMotionIDs("123", AlgosVersion)
	if err != nil {
		t.Fatalf("ListUnprocessedRawMotionIDs() returns err: %v", err)
	}
	if len(unprocessedRawMotionIDs) != 0 {
		t.Fatalf("Want 0 unprocessedRawMotionIDs, got %d", len(unprocessedRawMotionIDs))
	}
	unprocessedRawVideoIDs, err := rawVideoDAO.ListUnprocessedRawVideoIDs("123", AlgosVersion)
	if err != nil {
		t.Fatalf("ListUnprocessedRawVideoIDs() returns err: %v", err)
	}
	if len(unprocessedRawVideoIDs) != 0 {
		t.Fatalf("Want 0 unprocessedRawVideoIDs, got %d", len(unprocessedRawVideoIDs))
	}
}

func newDataProcessorPartsForTest() (dao.TripDAO, dao.EventDAO, dao.DrivingConditionDAO, dao.RawMotionDAO, dao.RawVideoDAO, logging.LoggingInterface) {
	logger := logging.NewLocalLogger(false)
	sqlInterface := database.NewFake()
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(sqlInterface, logger)
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlInterface, logger, time.Minute)
	tripDAO := tripdao.NewSQLTripDAO(sqlInterface, logger)
	eventDAO := eventdao.NewSQLEventDAO(sqlInterface, tripDAO, logger)
	drivingConditionDAO := drivingconditiondao.NewSQLDrivingConditionDAO(sqlInterface, tripDAO, eventDAO)
	return tripDAO, eventDAO, drivingConditionDAO, rawMotionDAO, rawVideoDAO, logger
}
