package dataprocessor

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/client/weather"
	"seneca/internal/client/weather/service"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/util"
	"sort"
	"testing"
	"time"
)

func TestRunForRawMotions(t *testing.T) {
	tripDAO, eventDAO, drivingConditionDAO, rawMotionDAO, rawLocationDAO, rawVideoDAO, logger := newDataProcessorPartsForTest()

	accAlgo, err := newAccelerationV0()
	if err != nil {
		t.Fatalf("newAccelerationV0() returns err: %v", err)
	}
	decAlgo, err := newDecelerationV0()
	if err != nil {
		t.Fatalf("newDecelerationV0() returns err: %v", err)
	}

	dp, err := New([]AlgorithmInterface{accAlgo, decAlgo}, eventDAO, drivingConditionDAO, rawMotionDAO, rawLocationDAO, rawVideoDAO, logger)
	if err != nil {
		t.Fatalf("New() returns err: %v", err)
	}

	// Create a few raw motions.
	for _, acc := range []int{-35, -22, -12, -2, 5, 10, 15, 20} {
		rawMotion := &st.RawMotion{
			UserId: "123",
			Motion: &st.Motion{
				AccelerationMphS: float64(acc),
			},
		}

		if acc < 0 {
			acc *= -1
		}

		rawMotion.TimestampMs = util.TimeToMilliseconds(time.Date(2021, 05, 05, 0, 0, acc, 0, time.UTC))

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

	if len(eventIDs) != 6 {
		t.Fatalf("Want 6 event IDs, got %d", len(eventIDs))
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
		if ev.AlgoTag == "" {
			t.Fatalf("Empty AlgoTag for event")
		}
		if ev.EventType != st.EventType_FAST_ACCELERATION && ev.EventType != st.EventType_FAST_DECELERATION {
			t.Fatalf("Want EventType %q or %q, got %q", st.EventType_FAST_ACCELERATION, st.EventType_FAST_DECELERATION, ev.EventType)
		}
		if ev.Source.SourceType != st.Source_RAW_MOTION {
			t.Fatalf("Want SourceType %q, got %q", st.Source_RAW_MOTION, ev.Source.SourceType)
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

func TestRunForRawLocations(t *testing.T) {
	tripDAO, eventDAO, drivingConditionDAO, rawMotionDAO, rawLocationDAO, rawVideoDAO, logger := newDataProcessorPartsForTest()
	fakeWeatherService := service.NewMock()
	dp, err := New([]AlgorithmInterface{newWeatherV0(fakeWeatherService)}, eventDAO, drivingConditionDAO, rawMotionDAO, rawLocationDAO, rawVideoDAO, logger)
	if err != nil {
		t.Fatalf("New() returns err: %v", err)
	}

	startTime := time.Date(2021, 05, 30, 0, 0, 0, 0, time.UTC)
	weatherCodes := []int{113, 113, 230, 230, 230, 314, 314, 314, 314, 308, 308, 308, 308, 308, 308, 143, 143, 143, 143, 143, 113, 113, 113, 113}

	// The goal is to be in this state: __SSSFFFFRRRRRRMMMMM____ (M is mist)
	results := []*weather.TimestampedWeatherCondition{}
	for i := 0; i < 24; i++ {
		twc := &weather.TimestampedWeatherCondition{
			StartTime:   startTime.Add(time.Hour * time.Duration(i)),
			EndTime:     startTime.Add(time.Hour * time.Duration(i+1)),
			Source:      weather.FakeWeatherSource,
			Lat:         &st.Latitude{},
			Long:        &st.Longitude{},
			WeatherCode: weatherCodes[i],
		}
		results = append(results, twc)
	}
	fakeWeatherService.InsertHistoricalWeather(results, &st.Latitude{}, &st.Longitude{})

	// Insert RawLocation every 59 minutes.
	for i := 0; i < 1440; i += 59 {
		rawLocation := &st.RawLocation{
			Id:     fmt.Sprintf("%d", i),
			UserId: "123",
			Location: &st.Location{
				Lat:  &st.Latitude{},
				Long: &st.Longitude{},
			},
			TimestampMs: util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(i))),
		}
		if _, err := rawLocationDAO.InsertUniqueRawLocation(rawLocation); err != nil {
			t.Fatalf("PutRawLocationByID() returns err: %v", err)
		}
	}

	// And a raw video to associate.
	rawVideo := &st.RawVideo{
		UserId:       "123",
		CreateTimeMs: util.TimeToMilliseconds(startTime),
		DurationMs:   (time.Duration(24) * time.Hour).Milliseconds(),
	}
	if _, err := rawVideoDAO.InsertUniqueRawVideo(rawVideo); err != nil {
		t.Fatalf("InsertUniqueRawVideo() returns err: %v", err)
	}

	dp.Run("123")

	tripIDs, err := tripDAO.ListUserTripIDs("123")
	if err != nil {
		t.Fatalf("ListUserTripIDs() returns err: %v", err)
	}
	if len(tripIDs) != 1 {
		t.Fatalf("Want 1 tripID, got %d", len(tripIDs))
	}

	drivingConditionIDs, err := drivingConditionDAO.ListTripDrivingConditionIDs("123", tripIDs[0])
	if err != nil {
		t.Fatalf("ListTripDrivingConditionIDs() returns err: %v", err)
	}
	if len(drivingConditionIDs) != 7 {
		t.Fatalf("Want 7 drivingConditionIDs, got %d", len(drivingConditionIDs))
	}

	// The goal is to be in this state: __SSSFFFFRRRRRRMMMMM____ (M is mist)
	wantDrivingConditions := []*st.DrivingConditionInternal{
		{
			ConditionType: st.ConditionType_NONE_CONDITION_TYPE,
			StartTimeMs:   util.TimeToMilliseconds(startTime),
			EndTimeMs:     util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*2))),
		},
		{
			ConditionType: st.ConditionType_SNOW,
			StartTimeMs:   util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*3))),
			EndTimeMs:     util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*5))),
		},
		{
			ConditionType: st.ConditionType_FREEZING_RAIN,
			StartTimeMs:   util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*6))),
			EndTimeMs:     util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*9))),
		},
		{
			ConditionType: st.ConditionType_RAIN,
			StartTimeMs:   util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*10))),
			EndTimeMs:     util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*15))),
		},
		{
			ConditionType: st.ConditionType_FOG,
			StartTimeMs:   util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*16))),
			EndTimeMs:     util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*20))),
		},
		{
			ConditionType: st.ConditionType_NONE_CONDITION_TYPE,
			StartTimeMs:   util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*21))),
			EndTimeMs:     util.TimeToMilliseconds(startTime.Add(time.Minute * time.Duration(59*24))),
		},
		{
			ConditionType: st.ConditionType_NONE_CONDITION_TYPE,
			StartTimeMs:   util.TimeToMilliseconds(startTime),
			EndTimeMs:     util.TimeToMilliseconds(startTime.Add(time.Hour * time.Duration(24))),
		},
	}

	gotDrivingConditions := []*st.DrivingConditionInternal{}
	for _, dcid := range drivingConditionIDs {
		dc, err := drivingConditionDAO.GetDrivingConditionByID("123", tripIDs[0], dcid)
		if err != nil {
			t.Fatalf("GetDrivingConditionByID() returns err: %v", err)
		}
		gotDrivingConditions = append(gotDrivingConditions, dc)
	}

	sort.Slice(gotDrivingConditions, func(i, j int) bool {
		if gotDrivingConditions[i].StartTimeMs != gotDrivingConditions[j].StartTimeMs {
			return gotDrivingConditions[i].StartTimeMs < gotDrivingConditions[j].StartTimeMs
		}
		return gotDrivingConditions[i].EndTimeMs < gotDrivingConditions[j].EndTimeMs
	})
	sort.Slice(wantDrivingConditions, func(i, j int) bool {
		if wantDrivingConditions[i].StartTimeMs != wantDrivingConditions[j].StartTimeMs {
			return wantDrivingConditions[i].StartTimeMs < wantDrivingConditions[j].StartTimeMs
		}
		return wantDrivingConditions[i].EndTimeMs < wantDrivingConditions[j].EndTimeMs
	})

	for i, wantDC := range wantDrivingConditions {
		if wantDC.StartTimeMs != gotDrivingConditions[i].StartTimeMs {
			t.Fatalf("Want startTimeMs %v, got %v", util.MillisecondsToTime(wantDC.StartTimeMs), util.MillisecondsToTime(gotDrivingConditions[i].StartTimeMs))
		}
		if wantDC.EndTimeMs != gotDrivingConditions[i].EndTimeMs {
			t.Fatalf("Want endTimeMs %v, got %v", util.MillisecondsToTime(wantDC.EndTimeMs), util.MillisecondsToTime(gotDrivingConditions[i].EndTimeMs))
		}
		if wantDC.ConditionType != gotDrivingConditions[i].ConditionType {
			t.Fatalf("Want conditionType %q, got %q", wantDC.ConditionType, gotDrivingConditions[i].ConditionType)
		}
	}
}

func newDataProcessorPartsForTest() (dao.TripDAO, dao.EventDAO, dao.DrivingConditionDAO, dao.RawMotionDAO, dao.RawLocationDAO, dao.RawVideoDAO, logging.LoggingInterface) {
	logger := logging.NewLocalLogger(false)
	sqlInterface := database.NewFake()
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(sqlInterface, logger)
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlInterface, logger, time.Minute)
	rawLocationDAO := rawlocationdao.NewSQLRawLocationDAO(sqlInterface)
	tripDAO := tripdao.NewSQLTripDAO(sqlInterface, logger)
	eventDAO := eventdao.NewSQLEventDAO(sqlInterface, tripDAO, logger)
	drivingConditionDAO := drivingconditiondao.NewSQLDrivingConditionDAO(sqlInterface, tripDAO, eventDAO)
	return tripDAO, eventDAO, drivingConditionDAO, rawMotionDAO, rawLocationDAO, rawVideoDAO, logger
}
