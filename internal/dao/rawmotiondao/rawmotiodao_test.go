package rawmotiondao_test

import (
	"errors"
	"log"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/util"
	"seneca/test/testutil"
	"sort"
	"testing"
	"time"
)

var createTime = time.Date(1996, time.May, 23, 0, 0, 0, 0, time.UTC)

func TestInsertUniqueRawMotion(t *testing.T) {
	rawMotion := &st.RawMotion{
		UserId:      testutil.TestUserID,
		TimestampMs: util.TimeToMilliseconds(createTime),
	}

	dao, sql := newRawMotionDAOForTest()

	alreadyExistingRawMotion := &st.RawMotion{
		UserId:      testutil.TestUserID,
		TimestampMs: util.TimeToMilliseconds(createTime),
	}

	// Already exists.
	existingID, err := sql.Create(constants.RawMotionsTable, alreadyExistingRawMotion)
	if err != nil {
		t.Fatalf("sql.Create(_, alreadyExistingRawMotion) returns err: %v", err)
	}

	if _, err := dao.InsertUniqueRawMotion(rawMotion); err == nil {
		t.Fatalf("Expected err for InsertUniqueRawMotion() when already exists, got nil")
	}

	// No conflict with time difference.
	alreadyExistingRawMotion.TimestampMs += (int64(time.Second) * 100)
	if err := sql.Insert(constants.RawMotionsTable, existingID, alreadyExistingRawMotion); err != nil {
		t.Fatalf("DeleteByID() returns err: %v", err)
	}

	rawMotionWithID, err := dao.InsertUniqueRawMotion(rawMotion)
	if err != nil {
		t.Fatalf("InsertUniqueRawMotion() returns err: %v", err)
	}
	if rawMotionWithID.Id == "" {
		t.Fatalf("Newly created RawMotion not assigned ID")
	}

	// Now induce errors for coverage.
	if err := dao.DeleteRawMotionByID(rawMotionWithID.Id); err != nil {
		t.Fatalf("DeleteRawMotionByID() returns err: %v", err)
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

		if _, err := dao.InsertUniqueRawMotion(rawMotion); err == nil {
			t.Fatalf("Expected err from InsertUniqueRawMotion() when call %d fails, got nil", i)
		}

		close(sql.ErrorCalls)
	}
}

func TestListUserRawMotionIDs(t *testing.T) {
	dao, _ := newRawMotionDAOForTest()

	wantRawMotions := []*st.RawMotion{}
	for i := 0; i < 20; i++ {
		userID := testutil.TestUserID
		if i%2 == 0 {
			userID = "546"
		}

		rawMotion := &st.RawMotion{
			UserId:      userID,
			TimestampMs: util.TimeToMilliseconds(createTime) + (time.Second.Milliseconds() * int64(i)),
		}

		rawMotionWithID, err := dao.InsertUniqueRawMotion(rawMotion)
		if err != nil {
			t.Fatalf("InsertUniqueRawMotion(%d) returns err: %v", i, err)
		}

		if userID == testutil.TestUserID {
			wantRawMotions = append(wantRawMotions, rawMotionWithID)
		}
	}

	gotIDs, err := dao.ListUserRawMotionIDs(testutil.TestUserID)
	if err != nil {
		t.Fatalf("ListUserRawMotionIDs() returns err: %v", err)
	}

	if len(wantRawMotions) != len(gotIDs) {
		t.Fatalf("Want %d rawMotionIDs, got %d", len(wantRawMotions), len(gotIDs))
	}

	sort.Slice(wantRawMotions, func(i, j int) bool { return wantRawMotions[i].Id < wantRawMotions[j].Id })
	sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })

	for i := range wantRawMotions {
		if wantRawMotions[i].Id != gotIDs[i] {
			log.Fatalf("Want rawMotion ID %q, got %q", wantRawMotions[i].Id, gotIDs[i])
		}
	}
}

func TestGetRawMotionByID(t *testing.T) {
	rawMotion := &st.RawMotion{
		UserId:      testutil.TestUserID,
		TimestampMs: util.TimeToMilliseconds(createTime),
	}

	dao, sql := newRawMotionDAOForTest()

	_, err := dao.GetRawMotionByID(testutil.TestUserID)
	if err == nil {
		t.Fatalf("Expected err from GetRawMotionByID() for non-existant ID, got nil")
	}
	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("Want NotFoundError from GetRawMotionByID() for non-existant ID, got %v", err)
	}

	rawMotionWithID, err := dao.InsertUniqueRawMotion(rawMotion)
	if err != nil {
		t.Fatalf("InsertUniqueRawMotion() returns err: %v", err)
	}

	// No error.
	gotRawMotion, err := dao.GetRawMotionByID(rawMotionWithID.Id)
	if err != nil {
		t.Fatalf("GetRawMotionByID() returns err: %v", err)
	}
	if rawMotionWithID.UserId != gotRawMotion.UserId {
		t.Fatalf("RawMotions have same IDs but different values: %v - %v", rawMotionWithID, gotRawMotion)
	}

	// Induce error.
	sql.ErrorCalls = make(chan bool, 1)
	sql.ErrorCalls <- true
	if _, err := dao.GetRawMotionByID(rawMotionWithID.Id); err == nil {
		t.Fatalf("Expected err from GetRawMotionByID() when call fails, got nil")
	}
	close(sql.ErrorCalls)
}

func newRawMotionDAOForTest() (*rawmotiondao.SQLRawMotionDAO, *database.FakeSQLDBService) {
	fakeSQLService := database.NewFake()
	logger := logging.NewLocalLogger(true)
	return rawmotiondao.NewSQLRawMotionDAO(fakeSQLService, logger), fakeSQLService
}
