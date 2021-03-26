package rawvideohandler

import (
	"fmt"
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

	userId := "user_id"
	bucketFileName := "bucket_file_name"
	creationTime := time.Date(2021, 3, 4, 0, 0, 0, 0, time.UTC)
	duration := time.Minute * 2
	fileMetdata := mp4.VideoMetadata{
		CreationTime: &creationTime,
		Duration:     &duration,
	}

	expectedRawVideo := &types.RawVideo{
		UserId:       userId,
		CreateTimeMs: util.TimeToMilliseconds(&creationTime),
	}

	// Verify value.
	returnedRawVideo, err := rawVideoHandler.writeMP4MetadataToGCD(userId, bucketFileName, &fileMetdata)
	if err != nil {
		t.Errorf("rawVideoHandler.writeMP4MetadataToGCD(%s, %s, _) returns err: %v", userId, bucketFileName, err)
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
	gotRawVideo, err := rawVideoHandler.noSqlDB.GetRawVideo(userId, creationTime)
	if err != nil {
		t.Errorf("fakeFakeNoSQLDBClient.GetRawVideo(%q, _) returns err: %v", userId, err)
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

	userId := "user_id"
	bucketFileName := "bucket_file_name"
	creationTime := time.Date(2021, 3, 4, 0, 0, 0, 0, time.UTC)
	duration := time.Minute * 2
	fileMetdata := mp4.VideoMetadata{
		CreationTime: &creationTime,
		Duration:     &duration,
	}

	if _, err = rawVideoHandler.writeMP4MetadataToGCD(userId, bucketFileName, &fileMetdata); err != nil {
		t.Errorf("rawVideoHandler.writeMP4MetadataToGCD(%s, %s, _) returns err: %v", userId, bucketFileName, err)
	}
	// Write twice but add an extra second, should still fail.
	newTime := fileMetdata.CreationTime.Add(time.Second)
	fileMetdata.CreationTime = &newTime
	if _, err = rawVideoHandler.writeMP4MetadataToGCD(userId, bucketFileName, &fileMetdata); err == nil {
		t.Errorf("rawVideoHandler.writeMP4MetadataToGCD(%s, %s, _) should have returned err for duplicate, but did not", userId, bucketFileName)
	}
}

func newRawVideoHandlerForTests() (*RawVideoHandler, error) {
	fakeSimpleStorageClient := cloud.NewFakeSimpleStorageClient()
	fakeFakeNoSQLDBClient := cloud.NewFakeNoSQLDatabaseClient(time.Second * 2)
	localLogger := logging.NewLocalLogger()

	rawVideoHandler, err := NewRawVideoHandler(fakeSimpleStorageClient, fakeFakeNoSQLDBClient, localLogger, "", "")
	if err != nil {
		return nil, fmt.Errorf("NewRawVideoHandler returns err: %v", err)
	}
	return rawVideoHandler, nil
}
