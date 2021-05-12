package tripdao

import (
	"context"
	"errors"
	"log"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/util"
	"seneca/test/testutil"
	"sort"
	"testing"
	"time"
)

var (
	startTime = time.Date(1996, time.May, 23, 0, 0, 0, 0, time.UTC)
	endTime   = time.Date(1996, time.May, 24, 0, 0, 0, 0, time.UTC)
)

func TestCreateUniqueTrip(t *testing.T) {
	tripInternal := &st.TripInternal{
		UserId:      testutil.TestUserID,
		StartTimeMs: util.TimeToMilliseconds(startTime),
		EndTimeMs:   util.TimeToMilliseconds(endTime),
	}

	dao, sql := newTripDAOForTest()

	alreadyExistingTrip := &st.TripInternal{
		UserId:      testutil.TestUserID,
		StartTimeMs: util.TimeToMilliseconds(startTime) - 1,
		EndTimeMs:   util.TimeToMilliseconds(endTime) + 1,
	}

	// Already exists.
	existingID, err := sql.Create(constants.TripTable, alreadyExistingTrip)
	if err != nil {
		t.Fatalf("sql.Create(_, alreadyExistingTrip) returns err: %v", err)
	}

	if _, err := dao.CreateUniqueTrip(context.TODO(), tripInternal); err == nil {
		t.Fatalf("Expected err for CreateUniqueTrip() when already exists, got nil")
	}

	// No conflict with time difference.
	alreadyExistingTrip.StartTimeMs += (int64(time.Hour) * 100)
	alreadyExistingTrip.EndTimeMs += (int64(time.Hour) * 100)
	if err := sql.Insert(constants.TripTable, existingID, alreadyExistingTrip); err != nil {
		t.Fatalf("DeleteByID() returns err: %v", err)
	}

	tripInternalWithID, err := dao.CreateUniqueTrip(context.TODO(), tripInternal)
	if err != nil {
		t.Fatalf("CreateUniqueTrip() returns err: %v", err)
	}
	if tripInternalWithID.Id == "" {
		t.Fatalf("Newly created Trip not assigned ID")
	}

	// Now induce errors for coverage.
	if err := dao.DeleteTripByID(context.TODO(), tripInternalWithID.Id); err != nil {
		t.Fatalf("DeleteTripByID() returns err: %v", err)
	}

	// 7? of them
	for i := 1; i < 7; i++ {
		sql.ErrorCalls = make(chan bool, 8)

		for j := 0; j < i; j++ {
			if j == i-1 {
				sql.ErrorCalls <- true
			} else {
				sql.ErrorCalls <- false
			}
		}

		if _, err := dao.CreateUniqueTrip(context.TODO(), tripInternal); err == nil {
			t.Fatalf("Expected err from CreateUniqueTrip() when call %d fails, got nil", i)
		}

		close(sql.ErrorCalls)
	}
}

func TestListUserTripIDs(t *testing.T) {
	dao, _ := newTripDAOForTest()

	wantTrips := []*st.TripInternal{}
	for i := 0; i < 20; i++ {
		userID := testutil.TestUserID
		if i%2 == 0 {
			userID = "546"
		}

		tripInternal := &st.TripInternal{
			UserId:      userID,
			StartTimeMs: util.TimeToMilliseconds(startTime) + (time.Hour.Milliseconds() * int64((100 * i))),
			EndTimeMs:   util.TimeToMilliseconds(endTime) + (time.Hour.Milliseconds() * int64((100 * i))),
		}

		tripInternalWithID, err := dao.CreateUniqueTrip(context.TODO(), tripInternal)
		if err != nil {
			t.Fatalf("CreateUniqueTrip(%d) returns err: %v", i, err)
		}

		if userID == testutil.TestUserID {
			wantTrips = append(wantTrips, tripInternalWithID)
		}
	}

	gotIDs, err := dao.ListUserTripIDs(testutil.TestUserID)
	if err != nil {
		t.Fatalf("ListUserTripIDs() returns err: %v", err)
	}

	if len(wantTrips) != len(gotIDs) {
		t.Fatalf("Want %d tripInternalIDs, got %d", len(wantTrips), len(gotIDs))
	}

	sort.Slice(wantTrips, func(i, j int) bool { return wantTrips[i].Id < wantTrips[j].Id })
	sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })

	for i := range wantTrips {
		if wantTrips[i].Id != gotIDs[i] {
			log.Fatalf("Want tripInternal ID %q, got %q", wantTrips[i].Id, gotIDs[i])
		}
	}
}

func TestGetTripByID(t *testing.T) {
	tripInternal := &st.TripInternal{
		UserId:      testutil.TestUserID,
		StartTimeMs: util.TimeToMilliseconds(startTime),
		EndTimeMs:   util.TimeToMilliseconds(endTime),
	}

	dao, sql := newTripDAOForTest()

	_, err := dao.GetTripByID(testutil.TestUserID, "456")
	if err == nil {
		t.Fatalf("Expected err from GetTripByID() for non-existant ID, got nil")
	}
	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("Want NotFoundError from GetTripByID() for non-existant ID, got %v", err)
	}

	tripInternalWithID, err := dao.CreateUniqueTrip(context.TODO(), tripInternal)
	if err != nil {
		t.Fatalf("CreateUniqueTrip() returns err: %v", err)
	}

	// UserID mismatch.
	_, err = dao.GetTripByID("otheruser", tripInternalWithID.Id)
	if err == nil {
		t.Fatalf("Expected err from GetTripByID() for user ID mismatch, got nil")
	}
	var bse *senecaerror.BadStateError
	if !errors.As(err, &bse) {
		t.Fatalf("Want BadStateError from GetTripByID() for user ID mismatch, got %v", err)
	}

	// No error.
	gotTrip, err := dao.GetTripByID(tripInternalWithID.UserId, tripInternalWithID.Id)
	if err != nil {
		t.Fatalf("GetTripByID() returns err: %v", err)
	}
	if tripInternalWithID.UserId != gotTrip.UserId {
		t.Fatalf("Trips have same IDs but different values: %v - %v", tripInternalWithID, gotTrip)
	}

	// Induce error.
	sql.ErrorCalls = make(chan bool, 1)
	sql.ErrorCalls <- true
	if _, err := dao.GetTripByID(tripInternalWithID.UserId, tripInternalWithID.Id); err == nil {
		t.Fatalf("Expected err from GetTripByID() when call fails, got nil")
	}
	close(sql.ErrorCalls)
}

func TestListUserTripIDsByTime(t *testing.T) {
	userID := testutil.TestUserID
	anchorTrip := &st.TripInternal{
		UserId:      userID,
		StartTimeMs: util.TimeToMilliseconds(time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)),
		EndTimeMs:   util.TimeToMilliseconds(time.Date(2021, time.January, 15, 0, 0, 0, 0, time.UTC)),
	}

	prevTrip := &st.TripInternal{
		UserId:      userID,
		StartTimeMs: util.TimeToMilliseconds(time.Date(2021, time.January, 3, 0, 0, 0, 0, time.UTC)),
		EndTimeMs:   util.TimeToMilliseconds(time.Date(2021, time.January, 8, 0, 0, 0, 0, time.UTC)),
	}

	nextTrip := &st.TripInternal{
		UserId:      userID,
		StartTimeMs: util.TimeToMilliseconds(time.Date(2021, time.January, 20, 0, 0, 0, 0, time.UTC)),
		EndTimeMs:   util.TimeToMilliseconds(time.Date(2021, time.January, 25, 0, 0, 0, 0, time.UTC)),
	}

	// [ ] is trip span, { } is query span
	testCases := []struct {
		desc       string
		trips      []*st.TripInternal
		queryStart time.Time
		queryEnd   time.Time
		wantNumIDs int
	}{
		{
			desc:       "[ ] { [ ] } [ ]",
			trips:      []*st.TripInternal{anchorTrip, prevTrip, nextTrip},
			queryStart: time.Date(2021, time.January, 9, 0, 0, 0, 0, time.UTC),
			queryEnd:   time.Date(2021, time.January, 16, 0, 0, 0, 0, time.UTC),
			wantNumIDs: 1,
		},
		{
			desc:       "[ ] [ { } ] [ ]",
			trips:      []*st.TripInternal{anchorTrip, prevTrip, nextTrip},
			queryStart: time.Date(2021, time.January, 11, 0, 0, 0, 0, time.UTC),
			queryEnd:   time.Date(2021, time.January, 14, 0, 0, 0, 0, time.UTC),
			wantNumIDs: 1,
		},
		{
			desc:       "[ ] [ ] { } [ ]",
			trips:      []*st.TripInternal{anchorTrip, prevTrip, nextTrip},
			queryStart: time.Date(2021, time.January, 16, 0, 0, 0, 0, time.UTC),
			queryEnd:   time.Date(2021, time.January, 19, 0, 0, 0, 0, time.UTC),
			wantNumIDs: 0,
		},
		{
			desc:       "[ ] { [ } ] [ ]",
			trips:      []*st.TripInternal{anchorTrip, prevTrip, nextTrip},
			queryStart: time.Date(2021, time.January, 9, 0, 0, 0, 0, time.UTC),
			queryEnd:   time.Date(2021, time.January, 12, 0, 0, 0, 0, time.UTC),
			wantNumIDs: 1,
		},
		{
			desc:       "[ ] [ { ] } [ ]",
			trips:      []*st.TripInternal{anchorTrip, prevTrip, nextTrip},
			queryStart: time.Date(2021, time.January, 12, 0, 0, 0, 0, time.UTC),
			queryEnd:   time.Date(2021, time.January, 18, 0, 0, 0, 0, time.UTC),
			wantNumIDs: 1,
		},
		{
			desc:       "[ { ] [ ] [ } ]",
			trips:      []*st.TripInternal{anchorTrip, prevTrip, nextTrip},
			queryStart: time.Date(2021, time.January, 8, 0, 0, 0, 0, time.UTC),
			queryEnd:   time.Date(2021, time.January, 20, 0, 0, 0, 0, time.UTC),
			wantNumIDs: 3,
		},
		{
			desc:       "[  ] { [ ] [ } ]",
			trips:      []*st.TripInternal{anchorTrip, prevTrip, nextTrip},
			queryStart: time.Date(2021, time.January, 9, 0, 0, 0, 0, time.UTC),
			queryEnd:   time.Date(2021, time.January, 20, 0, 0, 0, 0, time.UTC),
			wantNumIDs: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			dao, _ := newTripDAOForTest()

			for _, trip := range tc.trips {
				if _, err := dao.CreateUniqueTrip(context.TODO(), trip); err != nil {
					t.Fatalf("CreateUniqueTrip() for trip %v returns err: %v", trip, err)
				}
			}

			tripIDs, err := dao.ListUserTripIDsByTime(userID, tc.queryStart, tc.queryEnd)
			if err != nil {
				t.Fatalf("ListUserTripIDsByTime(_, %s, %s) returns err: %v", tc.queryStart, tc.queryEnd, err)
			}

			if tc.wantNumIDs != len(tripIDs) {
				t.Fatalf("Want %d tripIDs, got %d", tc.wantNumIDs, len(tripIDs))
			}
		})
	}

}

func newTripDAOForTest() (*SQLTripDAO, *database.FakeSQLDBService) {
	fakeSQLService := database.NewFake()
	return NewSQLTripDAO(fakeSQLService), fakeSQLService
}
