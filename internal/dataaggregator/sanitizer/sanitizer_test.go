package sanitizer

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/rawframedao"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/util"
	"seneca/test/testutil"
	"testing"
	"time"
)

func TestListTrips(t *testing.T) {
	// InternalDrivingConditions and Events - Trip is 24 hours
	// 		[______NNNNNNNNNNNN______] DrivingCondition 0
	//		[_______CC_______________] DrivingCondition 1
	// 		[____SSSSSS______________] DrivingCondition 2
	//		[________+++_____________] DrivingCondition 3
	//		[---------------------___] DrivingCondition 4
	//		doesnt really matter what the events are, sprinkle them in

	userID := testutil.TestUserID

	tripStart := time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)

	sanitizer, rawVideoDAO, rawMotionDAO, tripDAO, eventDAO, dcDAO := newSanitizerForTests()

	rawVideo := &st.RawVideo{
		UserId:               userID,
		CloudStorageFileName: "test.com.",
	}
	rawVideo, err := rawVideoDAO.InsertUniqueRawVideo(rawVideo)
	if err != nil {
		t.Fatalf("InsertUniqueRawVideo() returns err: %v", err)
	}

	rawMotion := &st.RawMotion{
		UserId: userID,
		Source: &st.Source{
			SourceId:   rawVideo.Id,
			SourceType: st.Source_RAW_VIDEO,
		},
	}
	rawMotion, err = rawMotionDAO.InsertUniqueRawMotion(rawMotion)
	if err != nil {
		t.Fatalf("InsertUniqueRawMotion() returns err: %v", err)
	}

	externalSource := &st.ExternalSource{
		SourceType: st.ExternalSource_DASHCAM_VIDEO,
		VideoUrl:   rawVideo.CloudStorageFileName,
	}

	drivingConditions := []*st.DrivingConditionInternal{
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 6)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 18)),
			ConditionType: st.ConditionType_NIGHT,
			Severity:      5,
			Source: &st.Source{
				SourceType: st.Source_RAW_VIDEO,
				SourceId:   rawVideo.Id,
			},
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 7)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 9)),
			ConditionType: st.ConditionType_URBAN,
			Severity:      7,
			Source: &st.Source{
				SourceType: st.Source_RAW_VIDEO,
				SourceId:   rawVideo.Id,
			},
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 4)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 10)),
			ConditionType: st.ConditionType_SNOW,
			Severity:      2,
			Source: &st.Source{
				SourceType: st.Source_RAW_VIDEO,
				SourceId:   rawVideo.Id,
			},
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 8)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 11)),
			ConditionType: st.ConditionType_SNOW,
			Severity:      3,
			Source: &st.Source{
				SourceType: st.Source_RAW_VIDEO,
				SourceId:   rawVideo.Id,
			},
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 0)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 21)),
			ConditionType: st.ConditionType_SNOW,
			Severity:      1,
			Source: &st.Source{
				SourceType: st.Source_RAW_VIDEO,
				SourceId:   rawVideo.Id,
			},
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 0)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 24)),
			ConditionType: st.ConditionType_NONE_CONDITION_TYPE,
			Source: &st.Source{
				SourceType: st.Source_RAW_VIDEO,
				SourceId:   rawVideo.Id,
			},
		},
	}

	expectedDrivingConditionsOut := []*st.DrivingCondition{
		{
			ConditionType: []st.ConditionType{st.ConditionType_SNOW},
			Severity:      []float64{1},
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 0)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add((time.Hour * 4) - 1)),
		},
		{
			ConditionType: []st.ConditionType{st.ConditionType_SNOW},
			Severity:      []float64{2},
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 4)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add((time.Hour * 6) - 1)),
		},
		{
			ConditionType: []st.ConditionType{st.ConditionType_SNOW, st.ConditionType_NIGHT},
			Severity:      []float64{2, 5},
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 6)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add((time.Hour * 7) - 1)),
		},
		{
			ConditionType: []st.ConditionType{st.ConditionType_SNOW, st.ConditionType_NIGHT, st.ConditionType_URBAN},
			Severity:      []float64{2, 5, 7},
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 7)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add((time.Hour * 8) - 1)),
		},
		{
			ConditionType: []st.ConditionType{st.ConditionType_SNOW, st.ConditionType_NIGHT, st.ConditionType_URBAN},
			Severity:      []float64{3, 5, 7},
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 8)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add((time.Hour * 9) - 1)),
		},
		{
			ConditionType: []st.ConditionType{st.ConditionType_SNOW, st.ConditionType_NIGHT},
			Severity:      []float64{3, 5},
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 9)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add((time.Hour * 11) - 1)),
		},
		{
			ConditionType: []st.ConditionType{st.ConditionType_SNOW, st.ConditionType_NIGHT},
			Severity:      []float64{1, 5},
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 11)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add((time.Hour * 18) - 1)),
		},
		{
			ConditionType: []st.ConditionType{st.ConditionType_SNOW},
			Severity:      []float64{1},
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 18)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add((time.Hour * 21) - 1)),
		},
		{
			ConditionType: []st.ConditionType{},
			Severity:      []float64{},
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 21)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add((time.Hour * 24) - 1)),
		},
	}
	// Just add the same external source for each.
	for _, edco := range expectedDrivingConditionsOut {
		for i := 0; i < len(edco.ConditionType); i++ {
			edco.ExternalSource = append(edco.ExternalSource, externalSource)
		}
	}

	for i := 0; i < 50; i++ {
		event := &st.EventInternal{
			UserId:      userID,
			EventType:   st.EventType(1 + rand.Intn(3)),
			Severity:    float64(rand.Intn(100)),
			TimestampMs: util.TimeToMilliseconds(tripStart.Add(time.Hour * time.Duration(rand.Intn(24)))),
			Source: &st.Source{
				SourceType: st.Source_RAW_MOTION,
				SourceId:   rawMotion.Id,
			},
		}
		if _, err := eventDAO.CreateEvent(context.TODO(), event); err != nil {
			t.Fatalf("CreateEvent() returns err: %v", err)
		}
	}

	for _, dc := range drivingConditions {
		dc.UserId = userID
		if _, err := dcDAO.CreateDrivingCondition(context.TODO(), dc); err != nil {
			t.Fatalf("CreateDrivingCondition() returns err: %v", err)
		}
	}

	tripInteralList, err := tripDAO.ListUserTripIDs(userID)
	if err != nil {
		log.Fatalf("tripDAO.ListUserTripIDs() returns err: %v", err)
	}
	if len(tripInteralList) != 1 {
		log.Fatalf("Want 1 trip for user, got %d", len(tripInteralList))
	}
	tripInternal, err := tripDAO.GetTripByID(userID, tripInteralList[0])
	if err != nil {
		log.Fatalf("tripDAO.GetTripByID() returns err: %v", err)
	}

	tripExternal, err := sanitizer.TripInternalToTripExternal(tripInternal)
	if err != nil {
		log.Fatalf("sanitizer.ListTrips() returns err: %v", err)
	}

	if len(tripExternal.Event) != 50 {
		log.Fatalf("Wanted 50 events for trip, got %d", len(tripExternal.Event))
	}

	for _, event := range tripExternal.Event {
		if event.ExternalSource.VideoUrl != rawVideo.CloudStorageFileName {
			t.Fatalf("Want %q for event.ExternalSource.VideoUrl, got %q", rawVideo.CloudStorageFileName, event.ExternalSource.VideoUrl)
		}
	}

	if len(expectedDrivingConditionsOut) != len(tripExternal.DrivingCondition) {
		log.Fatalf("Wanted %d drivingConditions for trip, got %d", len(expectedDrivingConditionsOut), len(tripExternal.DrivingCondition))
	}

	for i := range expectedDrivingConditionsOut {
		if !drivingConditionExternalEqual(expectedDrivingConditionsOut[i], tripExternal.DrivingCondition[i]) {
			log.Fatalf("DrivingConditions not equal: %v != %v", expectedDrivingConditionsOut[i], tripExternal.DrivingCondition[i])
		}
	}
}

func newSanitizerForTests() (*Sanitizer, dao.RawVideoDAO, dao.RawMotionDAO, dao.TripDAO, dao.EventDAO, dao.DrivingConditionDAO) {
	fakeSQL := database.NewFake()
	logger := logging.NewLocalLogger(false)
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(fakeSQL, logger, time.Minute)
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(fakeSQL, logger)
	rawLocationDAO := rawlocationdao.NewSQLRawLocationDAO(fakeSQL)
	rawFrameDAO := rawframedao.NewSQLRawFrameDAO(fakeSQL)
	tripDAO := tripdao.NewSQLTripDAO(fakeSQL, logger)
	eventDAO := eventdao.NewSQLEventDAO(fakeSQL, tripDAO, logger)
	dcDAO := drivingconditiondao.NewSQLDrivingConditionDAO(fakeSQL, tripDAO, eventDAO)
	return New(rawMotionDAO, rawLocationDAO, rawVideoDAO, rawFrameDAO, eventDAO, dcDAO), rawVideoDAO, rawMotionDAO, tripDAO, eventDAO, dcDAO
}

func drivingConditionExternalEqual(lhs *st.DrivingCondition, rhs *st.DrivingCondition) bool {
	if lhs.StartTimeMs != rhs.StartTimeMs || lhs.EndTimeMs != rhs.EndTimeMs {
		return false
	}

	lhsMap := map[string]float64{}
	for i := range lhs.ConditionType {
		condAndSrc := fmt.Sprintf("%s/%s/%s", lhs.ConditionType[i], lhs.ExternalSource[i].SourceType, lhs.ExternalSource[i].VideoUrl)
		lhsMap[condAndSrc] = lhs.Severity[i]
	}

	for i := range rhs.ConditionType {
		condAndSrc := fmt.Sprintf("%s/%s/%s", rhs.ConditionType[i], rhs.ExternalSource[i].SourceType, rhs.ExternalSource[i].VideoUrl)
		if lhsMap[condAndSrc] != rhs.Severity[i] {
			return false
		}
		delete(lhsMap, condAndSrc)
	}

	return len(lhsMap) == 0
}

func prettyPrintDrivingConditionList(drivingConditions []*st.DrivingCondition) string {
	output := "[\n"
	for _, dc := range drivingConditions {
		output += "  {\n"
		output += fmt.Sprintf("    Conditions: [%v],\n", dc.ConditionType)
		output += fmt.Sprintf("    Severities: [%v],\n", dc.Severity)
		output += fmt.Sprintf("    Start: %v,\n", util.MillisecondsToTime(dc.StartTimeMs))
		output += fmt.Sprintf("    End: %v,\n", util.MillisecondsToTime(dc.EndTimeMs))
		output += "  },\n"
	}
	output += "]\n"
	return output
}
