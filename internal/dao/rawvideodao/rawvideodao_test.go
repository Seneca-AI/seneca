package rawvideodao

import (
	"errors"
	"log"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/util"
	"seneca/test/testutil"
	"sort"
	"testing"
	"time"
)

var createTime = time.Date(1996, time.May, 23, 0, 0, 0, 0, time.UTC)

func TestInsertUniqueRawVideo(t *testing.T) {
	rawVideo := &st.RawVideo{
		UserId:       testutil.TestUserID,
		CreateTimeMs: util.TimeToMilliseconds(createTime),
	}

	dao, sql := newRawVideoDAOForTest(time.Second * 5)

	alreadyExistingRawVideo := &st.RawVideo{
		UserId:       testutil.TestUserID,
		CreateTimeMs: util.TimeToMilliseconds(createTime) + (time.Second.Milliseconds() * 3),
	}

	// Already exists.
	existingID, err := sql.Create(constants.RawVideosTable, alreadyExistingRawVideo)
	if err != nil {
		t.Fatalf("sql.Create(_, alreadyExistingRawVideo) returns err: %v", err)
	}

	if _, err := dao.InsertUniqueRawVideo(rawVideo); err == nil {
		t.Fatalf("Expected err for InsertUniqueRawVideo() when already exists, got nil")
	}

	// No conflict with time difference.
	alreadyExistingRawVideo.CreateTimeMs += (int64(time.Second) * 100)
	if err := sql.Insert(constants.RawVideosTable, existingID, alreadyExistingRawVideo); err != nil {
		t.Fatalf("DeleteByID() returns err: %v", err)
	}

	rawVideoWithID, err := dao.InsertUniqueRawVideo(rawVideo)
	if err != nil {
		t.Fatalf("InsertUniqueRawVideo() returns err: %v", err)
	}
	if rawVideoWithID.Id == "" {
		t.Fatalf("Newly created RawVideo not assigned ID")
	}

	// Now induce errors for coverage.
	if err := dao.DeleteRawVideoByID(rawVideoWithID.Id); err != nil {
		t.Fatalf("DeleteRawVideoByID() returns err: %v", err)
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

		if _, err := dao.InsertUniqueRawVideo(rawVideo); err == nil {
			t.Fatalf("Expected err from InsertUniqueRawVideo() when call %d fails, got nil", i)
		}

		close(sql.ErrorCalls)
	}
}

func TestListUserRawVideoIDs(t *testing.T) {
	dao, _ := newRawVideoDAOForTest(0)

	wantRawVideos := []*st.RawVideo{}
	for i := 0; i < 20; i++ {
		userID := testutil.TestUserID
		if i%2 == 0 {
			userID = "546"
		}

		rawVideo := &st.RawVideo{
			UserId:       userID,
			CreateTimeMs: util.TimeToMilliseconds(createTime) + (time.Second.Milliseconds() * int64(i)),
		}

		rawVideoWithID, err := dao.InsertUniqueRawVideo(rawVideo)
		if err != nil {
			t.Fatalf("InsertUniqueRawVideo(%d) returns err: %v", i, err)
		}

		if userID == testutil.TestUserID {
			wantRawVideos = append(wantRawVideos, rawVideoWithID)
		}
	}

	gotIDs, err := dao.ListUserRawVideoIDs(testutil.TestUserID)
	if err != nil {
		t.Fatalf("ListUserRawVideoIDs() returns err: %v", err)
	}

	if len(wantRawVideos) != len(gotIDs) {
		t.Fatalf("Want %d rawVideoIDs, got %d", len(wantRawVideos), len(gotIDs))
	}

	sort.Slice(wantRawVideos, func(i, j int) bool { return wantRawVideos[i].Id < wantRawVideos[j].Id })
	sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })

	for i := range wantRawVideos {
		if wantRawVideos[i].Id != gotIDs[i] {
			log.Fatalf("Want rawVideo ID %q, got %q", wantRawVideos[i].Id, gotIDs[i])
		}
	}
}

func TestGetRawVideoByID(t *testing.T) {
	rawVideo := &st.RawVideo{
		UserId:       testutil.TestUserID,
		CreateTimeMs: util.TimeToMilliseconds(createTime),
	}

	dao, sql := newRawVideoDAOForTest(time.Second * 5)

	_, err := dao.GetRawVideoByID(testutil.TestUserID)
	if err == nil {
		t.Fatalf("Expected err from GetRawVideoByID() for non-existant ID, got nil")
	}
	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("Want NotFoundError from GetRawVideoByID() for non-existant ID, got %v", err)
	}

	rawVideoWithID, err := dao.InsertUniqueRawVideo(rawVideo)
	if err != nil {
		t.Fatalf("InsertUniqueRawVideo() returns err: %v", err)
	}

	// No error.
	gotRawVideo, err := dao.GetRawVideoByID(rawVideoWithID.Id)
	if err != nil {
		t.Fatalf("GetRawVideoByID() returns err: %v", err)
	}
	if rawVideoWithID.UserId != gotRawVideo.UserId {
		t.Fatalf("RawVideos have same IDs but different values: %v - %v", rawVideoWithID, gotRawVideo)
	}

	// Induce error.
	sql.ErrorCalls = make(chan bool, 1)
	sql.ErrorCalls <- true
	if _, err := dao.GetRawVideoByID(rawVideoWithID.Id); err == nil {
		t.Fatalf("Expected err from GetRawVideoByID() when call fails, got nil")
	}
	close(sql.ErrorCalls)
}

func newRawVideoDAOForTest(offset time.Duration) (*SQLRawVideoDAO, *database.FakeSQLDBService) {
	fakeSQLService := database.NewFake()
	logger := logging.NewLocalLogger(true)

	return NewSQLRawVideoDAO(fakeSQLService, logger, offset), fakeSQLService
}
