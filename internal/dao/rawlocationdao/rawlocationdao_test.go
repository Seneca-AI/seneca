package rawlocationdao

import (
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

var createTime = time.Date(1996, time.May, 23, 0, 0, 0, 0, time.UTC)

func TestInsertUniqueRawLocation(t *testing.T) {
	rawLocation := &st.RawLocation{
		UserId:      testutil.TestUserID,
		TimestampMs: util.TimeToMilliseconds(createTime),
	}

	dao, sql := newRawLocationDAOForTest()

	alreadyExistingRawLocation := &st.RawLocation{
		UserId:      testutil.TestUserID,
		TimestampMs: util.TimeToMilliseconds(createTime),
	}

	// Already exists.
	existingID, err := sql.Create(constants.RawLocationsTable, alreadyExistingRawLocation)
	if err != nil {
		t.Fatalf("sql.Create(_, alreadyExistingRawLocation) returns err: %v", err)
	}

	if _, err := dao.InsertUniqueRawLocation(rawLocation); err == nil {
		t.Fatalf("Expected err for InsertUniqueRawLocation() when already exists, got nil")
	}

	// No conflict with time difference.
	alreadyExistingRawLocation.TimestampMs += (int64(time.Second) * 100)
	if err := sql.Insert(constants.RawLocationsTable, existingID, alreadyExistingRawLocation); err != nil {
		t.Fatalf("DeleteByID() returns err: %v", err)
	}

	rawLocationWithID, err := dao.InsertUniqueRawLocation(rawLocation)
	if err != nil {
		t.Fatalf("InsertUniqueRawLocation() returns err: %v", err)
	}
	if rawLocationWithID.Id == "" {
		t.Fatalf("Newly created RawLocation not assigned ID")
	}

	// Now induce errors for coverage.
	if err := dao.DeleteRawLocationByID(rawLocationWithID.Id); err != nil {
		t.Fatalf("DeleteRawLocationByID() returns err: %v", err)
	}

	// 3 of them
	for i := 1; i < 4; i++ {
		sql.ErrorCalls = make(chan bool, 3)

		for j := 0; j < i; j++ {
			if j == i-1 {
				sql.ErrorCalls <- true
			} else {
				sql.ErrorCalls <- false
			}
		}

		if _, err := dao.InsertUniqueRawLocation(rawLocation); err == nil {
			t.Fatalf("Expected err from InsertUniqueRawLocation() when call %d fails, got nil", i)
		}

		close(sql.ErrorCalls)
	}
}

func TestListUserRawLocationIDs(t *testing.T) {
	dao, _ := newRawLocationDAOForTest()

	wantRawLocations := []*st.RawLocation{}
	for i := 0; i < 20; i++ {
		userID := testutil.TestUserID
		if i%2 == 0 {
			userID = "546"
		}

		rawLocation := &st.RawLocation{
			UserId:      userID,
			TimestampMs: util.TimeToMilliseconds(createTime) + (time.Second.Milliseconds() * int64(i)),
		}

		rawLocationWithID, err := dao.InsertUniqueRawLocation(rawLocation)
		if err != nil {
			t.Fatalf("InsertUniqueRawLocation(%d) returns err: %v", i, err)
		}

		if userID == testutil.TestUserID {
			wantRawLocations = append(wantRawLocations, rawLocationWithID)
		}
	}

	gotIDs, err := dao.ListUserRawLocationIDs(testutil.TestUserID)
	if err != nil {
		t.Fatalf("ListUserRawLocationIDs() returns err: %v", err)
	}

	if len(wantRawLocations) != len(gotIDs) {
		t.Fatalf("Want %d rawLocationIDs, got %d", len(wantRawLocations), len(gotIDs))
	}

	sort.Slice(wantRawLocations, func(i, j int) bool { return wantRawLocations[i].Id < wantRawLocations[j].Id })
	sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })

	for i := range wantRawLocations {
		if wantRawLocations[i].Id != gotIDs[i] {
			log.Fatalf("Want rawLocation ID %q, got %q", wantRawLocations[i].Id, gotIDs[i])
		}
	}
}

func TestGetRawLocationByID(t *testing.T) {
	rawLocation := &st.RawLocation{
		UserId:      testutil.TestUserID,
		TimestampMs: util.TimeToMilliseconds(createTime),
	}

	dao, sql := newRawLocationDAOForTest()

	_, err := dao.GetRawLocationByID(testutil.TestUserID)
	if err == nil {
		t.Fatalf("Expected err from GetRawLocationByID() for non-existant ID, got nil")
	}
	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("Want NotFoundError from GetRawLocationByID() for non-existant ID, got %v", err)
	}

	rawLocationWithID, err := dao.InsertUniqueRawLocation(rawLocation)
	if err != nil {
		t.Fatalf("InsertUniqueRawLocation() returns err: %v", err)
	}

	// No error.
	gotRawLocation, err := dao.GetRawLocationByID(rawLocationWithID.Id)
	if err != nil {
		t.Fatalf("GetRawLocationByID() returns err: %v", err)
	}
	if rawLocationWithID.UserId != gotRawLocation.UserId {
		t.Fatalf("RawLocations have same IDs but different values: %v - %v", rawLocationWithID, gotRawLocation)
	}

	// Induce error.
	sql.ErrorCalls = make(chan bool, 1)
	sql.ErrorCalls <- true
	if _, err := dao.GetRawLocationByID(rawLocationWithID.Id); err == nil {
		t.Fatalf("Expected err from GetRawLocationByID() when call fails, got nil")
	}
	close(sql.ErrorCalls)
}

func newRawLocationDAOForTest() (*SQLRawLocationDAO, *database.FakeSQLDBService) {
	fakeSQLService := database.NewFake()
	return NewSQLRawLocationDAO(fakeSQLService), fakeSQLService
}
