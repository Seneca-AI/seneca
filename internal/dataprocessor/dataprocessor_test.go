package dataprocessor_test

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/logging"
	"seneca/internal/client/weather"
	"seneca/internal/client/weather/service"
	"seneca/internal/dao"
	"seneca/internal/dataprocessor"
	"seneca/internal/dataprocessor/algorithms"
	"seneca/internal/util"
	"seneca/test/testutil"
	"sort"
	"testing"
	"time"
)

func TestRunForRawMotions(t *testing.T) {
	allDAOSet, logger := newDataProcessorPartsForTest()

	algoFactory, err := algorithms.NewFactory(nil, nil)
	if err != nil {
		t.Fatalf("algorithms.NewFactory() returns err: %v", err)
	}
	algos := []dataprocessor.AlgorithmInterface{}
	algoTags := []string{"00000", "00001", "00002"}
	for _, tag := range algoTags {
		algo, err := algoFactory.GetAlgorithm(tag)
		if err != nil {
			t.Fatalf("GetAlgorithm(%q) returns err: %v", tag, err)
		}
		algos = append(algos, algo)
	}

	dp, err := dataprocessor.New(algos, allDAOSet, logger)
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

		if _, err := allDAOSet.RawMotionDAO.InsertUniqueRawMotion(rawMotion); err != nil {
			t.Fatalf("InsertUniqueRawMotion() returns err: %v", err)
		}
	}

	// And a raw video to associate.
	rawVideo := &st.RawVideo{
		UserId:       "123",
		CreateTimeMs: util.TimeToMilliseconds(time.Date(2021, 05, 05, 0, 0, 0, 0, time.UTC)),
		DurationMs:   int64(time.Minute * 1),
	}
	if _, err := allDAOSet.RawVideoDAO.InsertUniqueRawVideo(rawVideo); err != nil {
		t.Fatalf("InsertUniqueRawVideo() returns err: %v", err)
	}

	dp.Run("123")

	// Verify trip and events.
	tripIDs, err := allDAOSet.TripDAO.ListUserTripIDs("123")
	if err != nil {
		t.Fatalf("ListUserTripIDs() returns err: %v", err)
	}
	if len(tripIDs) != 1 {
		t.Fatalf("Want 1 trip ID, got %d", len(tripIDs))
	}

	eventIDs, err := allDAOSet.EventDAO.ListTripEventIDs("123", tripIDs[0])
	if err != nil {
		t.Fatalf("ListTripEventIDs() returns err: %v", err)
	}

	if len(eventIDs) != 6 {
		t.Fatalf("Want 6 event IDs, got %d", len(eventIDs))
	}

	events := []*st.EventInternal{}
	for _, eid := range eventIDs {
		event, err := allDAOSet.EventDAO.GetEventByID("123", tripIDs[0], eid)
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
	unprocessedRawMotionIDs, err := allDAOSet.RawMotionDAO.ListUnprocessedRawMotionIDs("123", dataprocessor.AlgosVersion)
	if err != nil {
		t.Fatalf("ListUnprocessedRawMotionIDs() returns err: %v", err)
	}
	if len(unprocessedRawMotionIDs) != 0 {
		t.Fatalf("Want 0 unprocessedRawMotionIDs, got %d", len(unprocessedRawMotionIDs))
	}
	unprocessedRawVideoIDs, err := allDAOSet.RawVideoDAO.ListUnprocessedRawVideoIDs("123", dataprocessor.AlgosVersion)
	if err != nil {
		t.Fatalf("ListUnprocessedRawVideoIDs() returns err: %v", err)
	}
	if len(unprocessedRawVideoIDs) != 0 {
		t.Fatalf("Want 0 unprocessedRawVideoIDs, got %d", len(unprocessedRawVideoIDs))
	}
}

func TestRunForRawLocations(t *testing.T) {
	allDAOSet, logger := newDataProcessorPartsForTest()
	fakeWeatherService := service.NewMock()

	algoFactory, err := algorithms.NewFactory(fakeWeatherService, nil)
	if err != nil {
		t.Fatalf("algorithms.NewFactory() returns err: %v", err)
	}

	algos := []dataprocessor.AlgorithmInterface{}
	algoTags := []string{"00000", "00003"}

	for _, tag := range algoTags {
		algo, err := algoFactory.GetAlgorithm(tag)
		if err != nil {
			t.Fatalf("GetAlgorithm(%q) returns err: %v", tag, err)
		}
		algos = append(algos, algo)
	}

	dp, err := dataprocessor.New(algos, allDAOSet, logger)
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
		if _, err := allDAOSet.RawLocationDAO.InsertUniqueRawLocation(rawLocation); err != nil {
			t.Fatalf("PutRawLocationByID() returns err: %v", err)
		}
	}

	// And a raw video to associate.
	rawVideo := &st.RawVideo{
		UserId:       "123",
		CreateTimeMs: util.TimeToMilliseconds(startTime),
		DurationMs:   (time.Duration(24) * time.Hour).Milliseconds(),
	}
	if _, err := allDAOSet.RawVideoDAO.InsertUniqueRawVideo(rawVideo); err != nil {
		t.Fatalf("InsertUniqueRawVideo() returns err: %v", err)
	}

	dp.Run("123")

	tripIDs, err := allDAOSet.TripDAO.ListUserTripIDs("123")
	if err != nil {
		t.Fatalf("ListUserTripIDs() returns err: %v", err)
	}
	if len(tripIDs) != 1 {
		t.Fatalf("Want 1 tripID, got %d", len(tripIDs))
	}

	drivingConditionIDs, err := allDAOSet.DrivingConditionDAO.ListTripDrivingConditionIDs("123", tripIDs[0])
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
		dc, err := allDAOSet.DrivingConditionDAO.GetDrivingConditionByID("123", tripIDs[0], dcid)
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

func newDataProcessorPartsForTest() (*dao.AllDAOSet, logging.LoggingInterface) {
	logger := logging.NewLocalLogger(false)
	return testutil.GenerateAllDAOSetWithFakeDB(logger, 0), logger
}
