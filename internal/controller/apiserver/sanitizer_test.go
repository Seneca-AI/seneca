package apiserver

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
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

	drivingConditions := []*st.DrivingConditionInternal{
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 6)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 18)),
			ConditionType: st.ConditionType_NIGHT,
			Severity:      5,
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 7)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 9)),
			ConditionType: st.ConditionType_URBAN,
			Severity:      7,
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 4)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 10)),
			ConditionType: st.ConditionType_SNOW,
			Severity:      2,
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 8)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 11)),
			ConditionType: st.ConditionType_SNOW,
			Severity:      3,
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 0)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 21)),
			ConditionType: st.ConditionType_SNOW,
			Severity:      1,
		},
		{
			StartTimeMs:   util.TimeToMilliseconds(tripStart.Add(time.Hour * 0)),
			EndTimeMs:     util.TimeToMilliseconds(tripStart.Add(time.Hour * 24)),
			ConditionType: st.ConditionType_NONE_CONDITION_TYPE,
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

	sanitizer, _, eventDAO, dcDAO := newSanitizerForTests()

	for i := 0; i < 50; i++ {
		event := &st.EventInternal{
			UserId:      userID,
			EventType:   st.EventType(1 + rand.Intn(3)),
			Severity:    float64(rand.Intn(100)),
			TimestampMs: util.TimeToMilliseconds(tripStart.Add(time.Hour * time.Duration(rand.Intn(24)))),
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

	tripsExternal, err := sanitizer.ListTrips(userID, tripStart.Add(time.Hour*-100), tripStart.Add(time.Hour*100))
	if err != nil {
		log.Fatalf("sanitizer.ListTrips() returns err: %v", err)
	}

	if len(tripsExternal) != 1 {
		log.Fatalf("Want 1 trip for user, got %d", len(tripsExternal))
	}

	if len(tripsExternal[0].Event) != 50 {
		log.Fatalf("Wanted 50 events for trip, got %d", len(tripsExternal[0].Event))
	}

	if len(expectedDrivingConditionsOut) != len(tripsExternal[0].DrivingCondition) {
		log.Fatalf("Wanted %d drivingConditions for trip, got %d", len(expectedDrivingConditionsOut), len(tripsExternal[0].DrivingCondition))
	}

	for i := range expectedDrivingConditionsOut {
		if !drivingConditionExternalEqual(expectedDrivingConditionsOut[i], tripsExternal[0].DrivingCondition[i]) {
			log.Fatalf("DrivingConditions not equal: %v != %v", expectedDrivingConditionsOut[i], tripsExternal[0].DrivingCondition[i])
		}
	}

}

func newSanitizerForTests() (*Sanitizer, *tripdao.SQLTripDAO, *eventdao.SQLEventDAO, *drivingconditiondao.SQLDrivingConditionDAO) {
	fakeSQL := database.NewFake()
	tripDAO := tripdao.NewSQLTripDAO(fakeSQL)
	eventDAO := eventdao.NewSQLEventDAO(fakeSQL, tripDAO)
	dcDAO := drivingconditiondao.NewSQLDrivingConditionDAO(fakeSQL, tripDAO, eventDAO)
	return NewSanitizer(tripDAO, eventDAO, dcDAO), tripDAO, eventDAO, dcDAO
}

func drivingConditionExternalEqual(lhs *st.DrivingCondition, rhs *st.DrivingCondition) bool {
	if lhs.StartTimeMs != rhs.StartTimeMs || lhs.EndTimeMs != rhs.EndTimeMs {
		return false
	}

	lhsMap := map[st.ConditionType]float64{}
	for i := range lhs.ConditionType {
		lhsMap[lhs.ConditionType[i]] = lhs.Severity[i]
	}

	for i := range rhs.ConditionType {
		if lhsMap[rhs.ConditionType[i]] != rhs.Severity[i] {
			return false
		}
		delete(lhsMap, rhs.ConditionType[i])
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
