package cutvideohandler

import (
	"seneca/api/types"
	"seneca/internal/util"
	"seneca/internal/util/cloud"
	"seneca/internal/util/logging"
	"seneca/internal/util/mp4"
	"testing"
)

const (
	pathToNoMetadataMP4 = "../../../test/testdata/no_metadata.mp4"
)

func TestProcessAndCutRawVideoErrorHandling(t *testing.T) {
	fakeSimpleStorageClient := cloud.NewFakeSimpleStorageClient()
	fakeNoSQLDBClient := cloud.NewFakeNoSQLDatabaseClient()
	fakeMP4Tool := mp4.NewFakeMP4Tool()
	localLogger := logging.NewLocalLogger(true /* silent */)

	userID := "321"
	createTimeMs := int64(500)
	simpleStorageFileName := "rawVideo.mp4"
	rawVideo := &types.RawVideo{
		Id:                   util.GenerateRandID(),
		UserId:               userID,
		CreateTimeMs:         createTimeMs,
		CloudStorageFileName: simpleStorageFileName,
	}
	// cutVideo := &types.CutVideo{
	// 	Id:           util.GenerateRandID(),
	// 	UserId:       userID,
	// 	RawVideoId:   rawVideoID,
	// 	CreateTimeMs: createTimeMs,
	// }

	cutVideoHandler, err := NewCutVideoHandler(fakeSimpleStorageClient, fakeNoSQLDBClient, fakeMP4Tool, localLogger, "")
	if err != nil {
		t.Errorf("NewCutVideoHandler() returns err: %v", err)
	}

	fakeNoSQLDBClient.GetRawVideoByIDMock = func(id string) (*types.RawVideo, error) {
		return rawVideo, nil
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("idontexist"); err == nil {
		t.Error("Expected error from ProcessAndCutRawVideo() for non-existent raw video, got nil")
	}
}
