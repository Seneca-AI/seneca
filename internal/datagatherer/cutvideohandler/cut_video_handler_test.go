package cutvideohandler

import (
	"seneca/api/types"
	"seneca/internal/util/cloud"
	"seneca/internal/util/logging"
	"seneca/internal/util/mp4"
	"testing"
	"time"
)

const (
	pathToNoMetadataMP4 = "../../../test/testdata/no_metadata.mp4"
)

func TestProcessAndCutRawVideoErrorHandling(t *testing.T) {
	fakeSimpleStorageClient := cloud.NewFakeSimpleStorageClient()
	fakeNoSQLDBClient := cloud.NewFakeNoSQLDatabaseClient(time.Second * 2)
	fakeMP4Tool := mp4.NewFakeMP4Tool()
	localLogger := logging.NewLocalLogger(true /* silent */)

	userID := "321"
	var rawVideoID string
	createTimeMs := int64(500)
	simpleStorageFileName := "rawVideo.mp4"
	rawVideo := &types.RawVideo{
		UserId:               userID,
		CreateTimeMs:         createTimeMs,
		CloudStorageFileName: simpleStorageFileName,
	}
	var cutVideoID string
	cutVideo := &types.CutVideo{
		UserId:       userID,
		RawVideoId:   rawVideoID,
		CreateTimeMs: createTimeMs,
	}

	cutVideoHandler, err := NewCutVideoHandler(fakeSimpleStorageClient, fakeNoSQLDBClient, fakeMP4Tool, localLogger, "")
	if err != nil {
		t.Errorf("NewCutVideoHandler() returns err: %v", err)
	}

	if err := cutVideoHandler.ProcessAndCutRawVideo("idontexist"); err == nil {
		t.Error("Expected error from ProcessAndCutRawVideo() for non-existent raw video, got nil")
	}
	rawVideoID, err = fakeNoSQLDBClient.InsertUniqueRawVideo(rawVideo)
	if err != nil {
		t.Errorf("InsertUniqueRawVideo(%v) returns err: %v", rawVideo, err)
	}
	rawVideo.Id = rawVideoID

	cutVideoID, err = fakeNoSQLDBClient.InsertUniqueCutVideo(cutVideo)
	if err != nil {
		t.Errorf("InsertUniqueCutVideo(%v) returns err: %v", rawVideo, err)
	}
	if err = cutVideoHandler.ProcessAndCutRawVideo(rawVideo.Id); err != nil {
		t.Errorf("ProcessAndCutRawVideo() returns err for already existing cut video: %v", err)
	}

	if err := fakeNoSQLDBClient.DeleteCutVideoByID(cutVideoID); err != nil {
		t.Errorf("DeleteCutVideoByID(%s) returns err: %v", cutVideoID, err)
	}

	if err := cutVideoHandler.ProcessAndCutRawVideo(rawVideo.Id); err == nil {
		t.Error("Expected error from ProcessAndCutRawVideo() for non-existent rawVideo simple storage file, got nil")
	}

	if err := fakeSimpleStorageClient.CreateBucket(cloud.RawVideoBucketName); err != nil {
		t.Errorf("CreateBucket(%q) returns err: %w", cloud.RawVideoBucketName, err)
	}

	if err := fakeSimpleStorageClient.WriteBucketFile(cloud.RawVideoBucketName, pathToNoMetadataMP4, simpleStorageFileName); err != nil {
		t.Errorf("WriteBucketFile(%q, %q, %q) returns err: %w", cloud.RawVideoBucketName, pathToNoMetadataMP4, simpleStorageFileName, err)
	}
}
