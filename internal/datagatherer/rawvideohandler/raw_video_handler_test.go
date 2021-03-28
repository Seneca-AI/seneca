package rawvideohandler

import (
	"errors"
	"fmt"
	"net/http"
	"seneca/api/senecaerror"
	"seneca/api/types"
	"seneca/internal/util"
	"seneca/internal/util/cloud"
	"seneca/internal/util/logging"
	"seneca/internal/util/mp4"
	"testing"
	"time"
)

func TestWriteMP4MetadataToGCD(t *testing.T) {
	rawVideoHandler, err := newRawVideoHandlerForTests()
	if err != nil {
		t.Errorf("newRawVideoHandlerForTests() returns err: %v", err)
	}

	userID := "user_id"
	bucketFileName := "bucket_file_name"
	creationTime := time.Date(2021, 3, 4, 0, 0, 0, 0, time.UTC)
	duration := time.Minute * 2
	fileMetdata := mp4.VideoMetadata{
		CreationTime: &creationTime,
		Duration:     &duration,
	}

	expectedRawVideo := &types.RawVideo{
		UserId:       userID,
		CreateTimeMs: util.TimeToMilliseconds(&creationTime),
	}

	// Verify value.
	returnedRawVideo, err := rawVideoHandler.writeMP4MetadataToGCD(userID, bucketFileName, &fileMetdata)
	if err != nil {
		t.Errorf("rawVideoHandler.writeMP4MetadataToGCD(%s, %s, _) returns err: %v", userID, bucketFileName, err)
	}
	if expectedRawVideo.UserId != returnedRawVideo.UserId {
		t.Errorf("Got returned RawVideo.UserId %q, want %q", returnedRawVideo.UserId, expectedRawVideo.UserId)
	}
	if expectedRawVideo.CreateTimeMs != returnedRawVideo.CreateTimeMs {
		t.Errorf("Got returned RawVideo.CreateTimeMs %d, want %d", returnedRawVideo.CreateTimeMs, expectedRawVideo.CreateTimeMs)
	}
	if returnedRawVideo.Id == "" {
		t.Errorf("Return RawVideo.Id is empty, should have been set")
	}

	// Verify exists in store.
	gotRawVideo, err := rawVideoHandler.noSQLDB.GetRawVideo(userID, creationTime)
	if err != nil {
		t.Errorf("fakeFakeNoSQLDBClient.GetRawVideo(%q, _) returns err: %v", userID, err)
	}
	if expectedRawVideo.UserId != gotRawVideo.UserId {
		t.Errorf("Got returned RawVideo.UserId %q, want %q", gotRawVideo.UserId, expectedRawVideo.UserId)
	}
	if expectedRawVideo.CreateTimeMs != gotRawVideo.CreateTimeMs {
		t.Errorf("Got returned RawVideo.CreateTimeMs %d, want %d", gotRawVideo.CreateTimeMs, expectedRawVideo.CreateTimeMs)
	}
	if gotRawVideo.Id == "" {
		t.Errorf("Return RawVideo.Id is empty, should have been set")
	}
}

func TestWriteMP4MetadataToGCDDisallowsDuplicates(t *testing.T) {
	rawVideoHandler, err := newRawVideoHandlerForTests()
	if err != nil {
		t.Errorf("newRawVideoHandlerForTests() returns err: %v", err)
	}

	userID := "user_id"
	bucketFileName := "bucket_file_name"
	creationTime := time.Date(2021, 3, 4, 0, 0, 0, 0, time.UTC)
	duration := time.Minute * 2
	fileMetdata := mp4.VideoMetadata{
		CreationTime: &creationTime,
		Duration:     &duration,
	}

	if _, err = rawVideoHandler.writeMP4MetadataToGCD(userID, bucketFileName, &fileMetdata); err != nil {
		t.Errorf("rawVideoHandler.writeMP4MetadataToGCD(%s, %s, _) returns err: %v", userID, bucketFileName, err)
	}
	// Write twice but add an extra second, should still fail.
	newTime := fileMetdata.CreationTime.Add(time.Second)
	fileMetdata.CreationTime = &newTime
	_, err = rawVideoHandler.writeMP4MetadataToGCD(userID, bucketFileName, &fileMetdata)
	if err == nil {
		t.Errorf("rawVideoHandler.writeMP4MetadataToGCD(%s, %s, _) should have returned err for duplicate, but did not", userID, bucketFileName)
	}
	var ue *senecaerror.UserError
	if !errors.As(err, &ue) {
		t.Errorf("rawVideoHandler.writeMP4MetadataToGCD(%s, %s, _) should have returned UserError, but got %w", userID, bucketFileName, err)
	}
}

func TestInsertRawVideoFromRequestRejectsMalformedRequest(t *testing.T) {
	var err error
	var userError *senecaerror.UserError

	rawVideoHandler, err := newRawVideoHandlerForTests()
	if err != nil {
		t.Errorf("newRawVideoHandlerForTests() returns err: %v", err)
	}

	request := &http.Request{}
	request.Method = "GET"

	err = rawVideoHandler.InsertRawVideoFromRequest(request)
	if err == nil {
		t.Error("Want err from InsertRawVideoFromRequest() with GET method, got nil")
	}
	if !errors.As(err, &userError) {
		t.Errorf("Want UserError from InsertRawVideoFromRequest() GET method, got %v", err)
	}

	request.Method = "POST"
	err = rawVideoHandler.InsertRawVideoFromRequest(request)
	if err == nil {
		t.Error("Want err from InsertRawVideoFromRequest() without userID, got nil")
	}
	if !errors.As(err, &userError) {
		t.Errorf("Want UserError from InsertRawVideoFromRequest() without userID, got %v", err)
	}

	request.PostForm.Add("user_id", "user")
	err = rawVideoHandler.InsertRawVideoFromRequest(request)
	if err == nil {
		t.Error("Want err from InsertRawVideoFromRequest() without mp4, got nil")
	}
	if !errors.As(err, &userError) {
		t.Errorf("Want UserError from InsertRawVideoFromRequest() without mp4, got %v", err)
	}

	// TODO: test malformed mp4 files
}

func newRawVideoHandlerForTests() (*RawVideoHandler, error) {
	fakeSimpleStorageClient := cloud.NewFakeSimpleStorageClient()
	fakeFakeNoSQLDBClient := cloud.NewFakeNoSQLDatabaseClient(time.Second * 2)
	fakeMP4Tool := mp4.NewFakeMP4Tool()
	localLogger := logging.NewLocalLogger()

	rawVideoHandler, err := NewRawVideoHandler(fakeSimpleStorageClient, fakeFakeNoSQLDBClient, fakeMP4Tool, localLogger, "", "")
	if err != nil {
		return nil, fmt.Errorf("NewRawVideoHandler returns err: %v", err)
	}
	return rawVideoHandler, nil
}
