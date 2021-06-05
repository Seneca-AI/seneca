package rawframedao

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

func TestInsertUniqueRawFrame(t *testing.T) {
	RawFrame := &st.RawFrame{
		UserId:      testutil.TestUserID,
		TimestampMs: util.TimeToMilliseconds(createTime),
	}

	dao, sql := newRawFrameDAOForTest()

	alreadyExistingRawFrame := &st.RawFrame{
		UserId:      testutil.TestUserID,
		TimestampMs: util.TimeToMilliseconds(createTime),
	}

	// Already exists.
	existingID, err := sql.Create(constants.RawFramesTable, alreadyExistingRawFrame)
	if err != nil {
		t.Fatalf("sql.Create(_, alreadyExistingRawFrame) returns err: %v", err)
	}

	if _, err := dao.InsertUniqueRawFrame(RawFrame); err == nil {
		t.Fatalf("Expected err for InsertUniqueRawFrame() when already exists, got nil")
	}

	// No conflict with time difference.
	alreadyExistingRawFrame.TimestampMs += (int64(time.Second) * 100)
	if err := sql.Insert(constants.RawFramesTable, existingID, alreadyExistingRawFrame); err != nil {
		t.Fatalf("DeleteByID() returns err: %v", err)
	}

	RawFrameWithID, err := dao.InsertUniqueRawFrame(RawFrame)
	if err != nil {
		t.Fatalf("InsertUniqueRawFrame() returns err: %v", err)
	}
	if RawFrameWithID.Id == "" {
		t.Fatalf("Newly created RawFrame not assigned ID")
	}

	// Now induce errors for coverage.
	if err := dao.DeleteRawFrameByID(RawFrameWithID.Id); err != nil {
		t.Fatalf("DeleteRawFrameByID() returns err: %v", err)
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

		if _, err := dao.InsertUniqueRawFrame(RawFrame); err == nil {
			t.Fatalf("Expected err from InsertUniqueRawFrame() when call %d fails, got nil", i)
		}

		close(sql.ErrorCalls)
	}
}

func TestListUserRawFrameIDs(t *testing.T) {
	dao, _ := newRawFrameDAOForTest()

	wantRawFrames := []*st.RawFrame{}
	for i := 0; i < 20; i++ {
		userID := testutil.TestUserID
		if i%2 == 0 {
			userID = "546"
		}

		RawFrame := &st.RawFrame{
			UserId:      userID,
			TimestampMs: util.TimeToMilliseconds(createTime) + (time.Second.Milliseconds() * int64(i)),
		}

		RawFrameWithID, err := dao.InsertUniqueRawFrame(RawFrame)
		if err != nil {
			t.Fatalf("InsertUniqueRawFrame(%d) returns err: %v", i, err)
		}

		if userID == testutil.TestUserID {
			wantRawFrames = append(wantRawFrames, RawFrameWithID)
		}
	}

	gotIDs, err := dao.ListUserRawFrameIDs(testutil.TestUserID)
	if err != nil {
		t.Fatalf("ListUserRawFrameIDs() returns err: %v", err)
	}

	if len(wantRawFrames) != len(gotIDs) {
		t.Fatalf("Want %d RawFrameIDs, got %d", len(wantRawFrames), len(gotIDs))
	}

	sort.Slice(wantRawFrames, func(i, j int) bool { return wantRawFrames[i].Id < wantRawFrames[j].Id })
	sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })

	for i := range wantRawFrames {
		if wantRawFrames[i].Id != gotIDs[i] {
			log.Fatalf("Want RawFrame ID %q, got %q", wantRawFrames[i].Id, gotIDs[i])
		}
	}
}

func TestGetRawFrameByID(t *testing.T) {
	RawFrame := &st.RawFrame{
		UserId:      testutil.TestUserID,
		TimestampMs: util.TimeToMilliseconds(createTime),
	}

	dao, sql := newRawFrameDAOForTest()

	_, err := dao.GetRawFrameByID(testutil.TestUserID)
	if err == nil {
		t.Fatalf("Expected err from GetRawFrameByID() for non-existant ID, got nil")
	}
	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("Want NotFoundError from GetRawFrameByID() for non-existant ID, got %v", err)
	}

	RawFrameWithID, err := dao.InsertUniqueRawFrame(RawFrame)
	if err != nil {
		t.Fatalf("InsertUniqueRawFrame() returns err: %v", err)
	}

	// No error.
	gotRawFrame, err := dao.GetRawFrameByID(RawFrameWithID.Id)
	if err != nil {
		t.Fatalf("GetRawFrameByID() returns err: %v", err)
	}
	if RawFrameWithID.UserId != gotRawFrame.UserId {
		t.Fatalf("RawFrames have same IDs but different values: %v - %v", RawFrameWithID, gotRawFrame)
	}

	// Induce error.
	sql.ErrorCalls = make(chan bool, 1)
	sql.ErrorCalls <- true
	if _, err := dao.GetRawFrameByID(RawFrameWithID.Id); err == nil {
		t.Fatalf("Expected err from GetRawFrameByID() when call fails, got nil")
	}
	close(sql.ErrorCalls)
}

func TestListUnprocessedRawFramesIDs(t *testing.T) {
	RawFrameDAO, _ := newRawFrameDAOForTest()

	for i := 0; i < 10; i++ {
		RawFrame := &st.RawFrame{
			UserId:      "123",
			TimestampMs: util.TimeToMilliseconds(time.Now().Add(time.Minute * time.Duration(i))),
		}
		if i%2 == 0 {
			RawFrame.AlgosVersion = 2
			RawFrame.AlgoTag = []string{"02"}
		}
		if _, err := RawFrameDAO.InsertUniqueRawFrame(RawFrame); err != nil {
			t.Fatalf("InsertUniqueRawFrame() returns err: %v", err)
		}
	}

	ids, err := RawFrameDAO.ListUnprocessedRawFramesIDs("123", 2)
	if err != nil {
		t.Fatalf("ListUnprocessedRawFramesIDs() returns err: %v", err)
	}
	if len(ids) != 5 {
		t.Fatalf("Want 5 IDs, got %d", len(ids))
	}

}

func newRawFrameDAOForTest() (*SQLRawFrameDAO, *database.FakeSQLDBService) {
	fakeSQLService := database.NewFake()
	return NewSQLRawFrameDAO(fakeSQLService), fakeSQLService
}
