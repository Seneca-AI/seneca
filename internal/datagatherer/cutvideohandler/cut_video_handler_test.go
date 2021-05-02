package cutvideohandler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
	"seneca/internal/client/logging"
	"seneca/internal/util"
	"seneca/internal/util/mp4"
	"testing"
	"time"
)

func TestHandleRawVideoPostRequestErrorHandling(t *testing.T) {
	cutVideoHandler, _, _, _ := newCutVideoHandlerForTests(t)

	request := &http.Request{}
	writer := httptest.NewRecorder()

	request.Method = "GET"
	cutVideoHandler.HandleRawVideoPostRequest(writer, request)
	if writer.Code != 400 {
		t.Errorf("Expected 400 response for non-POST request, got %d", writer.Code)
	}

	request.Method = "POST"
	writer = httptest.NewRecorder()
	cutVideoHandler.HandleRawVideoPostRequest(writer, request)
	if writer.Code != 400 {
		t.Errorf("Expected 400 response for request without RawVideoID, got %d", writer.Code)
	}

	request.PostForm.Add(rawVideoIDPostFormKey, "ID")
	writer = httptest.NewRecorder()
	cutVideoHandler.HandleRawVideoPostRequest(writer, request)
	if writer.Code != 500 {
		t.Errorf("Expected 500 response for request that fails internally, got %d", writer.Code)
	}
}

func TestProcessAndCutRawVideo(t *testing.T) {
	cutVideoHandler, fakeNoSQLDBClient, fakeSimpleStorageClient, fakeMP4Tool := newCutVideoHandlerForTests(t)
	rawVideo := &st.RawVideo{
		Id:     util.GenerateRandID(),
		UserId: "user",
	}
	cutVideos := []*st.CutVideo{
		{
			Id:         util.GenerateRandID(),
			UserId:     "user",
			RawVideoId: rawVideo.Id,
		},
		{
			Id:         util.GenerateRandID(),
			UserId:     "user",
			RawVideoId: rawVideo.Id,
		},
	}

	fakeNoSQLDBClient.GetRawVideoByIDMock = func(id string) (*st.RawVideo, error) {
		return nil, fmt.Errorf("some error")
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when GetRawVideoByID throws err, got nil")
	}

	fakeNoSQLDBClient.GetRawVideoByIDMock = func(id string) (*st.RawVideo, error) {
		return rawVideo, nil
	}
	fakeNoSQLDBClient.GetCutVideoMock = func(userID string, createTime time.Time) (*st.CutVideo, error) {
		return nil, fmt.Errorf("some error")
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when GetCutVideoMock throws non-NotFoundError err, got nil")
	}

	fakeNoSQLDBClient.GetCutVideoMock = func(userID string, createTime time.Time) (*st.CutVideo, error) {
		return cutVideos[0], nil
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when GetCutVideoMock returns non-nil CutVideo, got nil")
	}

	fakeNoSQLDBClient.GetCutVideoMock = func(userID string, createTime time.Time) (*st.CutVideo, error) {
		return nil, senecaerror.NewNotFoundError(nil)
	}
	fakeSimpleStorageClient.GetBucketFileMock = func(bucketName cloud.BucketName, bucketFileName string) (string, error) {
		return "", fmt.Errorf("some error")
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when GetBucketFileMock returns err, got nil")
	}

	locations := []*st.Location{
		{
			Lat: &st.Latitude{
				Degrees:       1,
				DegreeMinutes: 2,
				DegreeSeconds: 3,
				LatDirection:  st.Latitude_NORTH,
			},
		},
		{
			Lat: &st.Latitude{
				Degrees:       3,
				DegreeMinutes: 4,
				DegreeSeconds: 5,
				LatDirection:  st.Latitude_NORTH,
			},
		},
	}

	motions := []*st.Motion{
		{
			VelocityMph:      10,
			AccelerationMphS: 1,
		},
		{
			VelocityMph:      20,
			AccelerationMphS: 2,
		},
	}

	times := []time.Time{
		time.Date(2021, 4, 6, 7, 8, 9, 0, time.UTC),
	}

	fakeSimpleStorageClient.GetBucketFileMock = func(bucketName cloud.BucketName, bucketFileName string) (string, error) {
		return "path", nil
	}
	fakeMP4Tool.ParseOutGPSMetadataMock = func(pathToVideo string) ([]*st.Location, []*st.Motion, []time.Time, error) {
		return nil, nil, nil, fmt.Errorf("some error")
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when ParseOutGPSMetadataMock returns err, got nil")
	}

	fakeMP4Tool.ParseOutGPSMetadataMock = func(pathToVideo string) ([]*st.Location, []*st.Motion, []time.Time, error) {
		return locations, motions, times, nil
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when ConstructRawLocationDatas returns err (due to mismatched locations and times len), got nil")
	}

	times = append(times, time.Date(2021, 4, 6, 7, 8, 10, 0, time.UTC))
	fakeMP4Tool.ParseOutGPSMetadataMock = func(pathToVideo string) ([]*st.Location, []*st.Motion, []time.Time, error) {
		return locations, motions, times, nil
	}
	fakeMP4Tool.CutRawVideoMock = func(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.CutVideo, []string, error) {
		return nil, nil, fmt.Errorf("some error")
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when CutRawVideoMock returns err, got nil")
	}

	fakeMP4Tool.CutRawVideoMock = func(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.CutVideo, []string, error) {
		return cutVideos, []string{"pathToVideo", "pathToVideo2"}, nil
	}

	errors := make(chan error, 2)
	errors <- nil
	errors <- fmt.Errorf("some error")
	fakeNoSQLDBClient.InsertUniqueCutVideoMock = func(cutVideo *st.CutVideo) (string, error) {
		return "id", <-errors
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when InsertUniqueCutVideoMock returns err, got nil")
	}

	fakeNoSQLDBClient.InsertUniqueCutVideoMock = func(cutVideo *st.CutVideo) (string, error) {
		return util.GenerateRandID(), nil
	}
	errors <- nil
	errors <- fmt.Errorf("some error")
	fakeNoSQLDBClient.InsertUniqueRawLocationMock = func(rawLocation *st.RawLocation) (string, error) {
		return "id", <-errors
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when InsertUniqueRawLocationMock returns err, got nil")
	}

	fakeNoSQLDBClient.InsertUniqueRawLocationMock = func(rawLocation *st.RawLocation) (string, error) {
		return util.GenerateRandID(), nil
	}
	errors <- nil
	errors <- fmt.Errorf("some error")
	fakeNoSQLDBClient.InsertUniqueRawMotionMock = func(rawMotion *st.RawMotion) (string, error) {
		return "id", <-errors
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when InsertUniqueRawMotionMock returns err, got nil")
	}

	fakeNoSQLDBClient.InsertUniqueRawMotionMock = func(rawMotion *st.RawMotion) (string, error) {
		return util.GenerateRandID(), nil
	}
	errors <- nil
	errors <- fmt.Errorf("some error")
	fakeSimpleStorageClient.WriteBucketFileMock = func(bucketName cloud.BucketName, localFileNameAndPath, bucketFileName string) error {
		return <-errors
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err == nil {
		t.Errorf("Expected err when WriteBucketFileMock returns err, got nil")
	}

	fakeSimpleStorageClient.WriteBucketFileMock = func(bucketName cloud.BucketName, localFileNameAndPath, bucketFileName string) error {
		return nil
	}
	if err := cutVideoHandler.ProcessAndCutRawVideo("ID"); err != nil {
		t.Errorf("ProcessAndCutRawVideo returns err: %w", err)
	}
}

func newCutVideoHandlerForTests(t *testing.T) (*CutVideoHandler, *cloud.FakeNoSQLDatabaseClient, *cloud.FakeSimpleStorageClient, *mp4.FakeMP4Tool) {
	fakeSimpleStorageClient := cloud.NewFakeSimpleStorageClient()
	fakeNoSQLDBClient := cloud.NewFakeNoSQLDatabaseClient()
	fakeMP4Tool := mp4.NewFakeMP4Tool()
	localLogger := logging.NewLocalLogger(true /* silent */)

	cutVideoHandler, err := NewCutVideoHandler(fakeSimpleStorageClient, fakeNoSQLDBClient, fakeMP4Tool, localLogger, "")
	if err != nil {
		t.Errorf("NewCutVideoHandler() returns err: %v", err)
	}

	return cutVideoHandler, fakeNoSQLDBClient, fakeSimpleStorageClient, fakeMP4Tool
}
