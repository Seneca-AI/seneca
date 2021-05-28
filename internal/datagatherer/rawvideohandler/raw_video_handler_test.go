package rawvideohandler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
	"seneca/internal/client/logging"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/util"
	"seneca/internal/util/mp4"
	"seneca/test/testutil"
	"testing"
	"time"
)

func TestHandleRawVideoHTTPRequestRejectsMalformed(t *testing.T) {
	var err error
	var userError *senecaerror.UserError

	rawVideoHandler, _, _, _, _, _, err := newRawVideoHandlerForTests()
	if err != nil {
		t.Errorf("newRawVideoHandlerForTests() returns err: %v", err)
	}

	request := &http.Request{}
	request.Method = "GET"

	_, err = rawVideoHandler.convertHTTPRequestToRawVideoProcessRequest(request)
	if err == nil {
		t.Error("Want err from HandleRawVideoProcessRequest() with GET method, got nil")
	}
	if !errors.As(err, &userError) {
		t.Errorf("Want UserError from HandleRawVideoProcessRequest() GET method, got %v", err)
	}

	// TODO(lucaloncar): test the rest of HTTP parsing
}

func TestInsertRawVideoFromRequestErrorHandling(t *testing.T) {
	// TODO(lucaloncar): use mocks here
	if util.IsCIEnv() {
		t.Skip("Skipping exiftool test in GitHub env.")
	}

	rawVidHandler, fakeMP4Tool, fakeSSC, mockRawVideoDAO, _, _, err := newRawVideoHandlerForTests()
	if err != nil {
		t.Error(err)
	}

	request := &st.RawVideoProcessRequest{
		UserId:    testutil.TestUserID,
		VideoName: "illegalname{&}",
	}

	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err == nil {
		t.Errorf("Want err from RawVideoRequest with invalid file name, got nil")
	}
	request.VideoName = "no_metadata.mp4"
	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err == nil {
		t.Errorf("Want err from RawVideoRequest with no bytes, got nil")
	}

	data, err := ioutil.ReadFile("../../../test/testdata/no_metadata.mp4")
	if err != nil {
		t.Errorf("Error reading mp4 bytes: %v", err)
	}
	request.VideoBytes = data
	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err == nil {
		t.Errorf("Want err from RawVideoRequest with no metadata, got nil")
	}

	fakeMP4Tool.ParseVideoMetadataMock = func(pathToVideo string) (*st.RawVideo, []*st.Location, []*st.Motion, []time.Time, error) {
		return &st.RawVideo{
			DurationMs: time.Hour.Milliseconds(),
		}, nil, nil, nil, nil
	}

	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err == nil {
		t.Errorf("Want err from RawVideoRequest with long video, got nil")
	}
	fakeMP4Tool.ParseVideoMetadataMock = func(pathToVideo string) (*st.RawVideo, []*st.Location, []*st.Motion, []time.Time, error) {
		return &st.RawVideo{
			DurationMs: time.Minute.Milliseconds(),
		}, nil, nil, nil, nil
	}

	mockRawVideoDAO.DeleteRawVideoByIDMock = func(id string) error {
		return nil
	}

	mockRawVideoDAO.InsertUniqueRawVideoMock = func(rawVideo *st.RawVideo) (*st.RawVideo, error) {
		return nil, fmt.Errorf("")
	}

	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err == nil {
		t.Errorf("Want err from RawVideoRequest when InsertUniqueRawVideo returns err, got nil")
	}
	mockRawVideoDAO.InsertUniqueRawVideoMock = func(rawVideo *st.RawVideo) (*st.RawVideo, error) {
		rawVideo.Id = "1"
		return rawVideo, nil
	}

	fakeSSC.BucketExistsMock = func(bucketName cloud.BucketName) (bool, error) {
		return false, fmt.Errorf("error")
	}
	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err == nil {
		t.Errorf("Want err from RawVideoRequest when BucketExists returns err, got nil")
	}
	fakeSSC.BucketExistsMock = func(bucketName cloud.BucketName) (bool, error) {
		return false, nil
	}

	fakeSSC.CreateBucketMock = func(bucketName cloud.BucketName) error {
		return fmt.Errorf("")
	}
	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err == nil {
		t.Errorf("Want err from RawVideoRequest when CreateBucket returns err, got nil")
	}
	fakeSSC.CreateBucketMock = func(bucketName cloud.BucketName) error {
		return nil
	}

	fakeSSC.BucketFileExistsMock = func(bucketName cloud.BucketName, bucketFileName string) (bool, error) {
		return false, fmt.Errorf("error")
	}
	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err == nil {
		t.Errorf("Want err from RawVideoRequest when BucketFileExists returns err, got nil")
	}
	fakeSSC.BucketFileExistsMock = func(bucketName cloud.BucketName, bucketFileName string) (bool, error) {
		return false, nil
	}

	fakeSSC.WriteBucketFileMock = func(bucketName cloud.BucketName, localFileNameAndPath, bucketFileName string) error {
		return fmt.Errorf("")
	}
	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err == nil {
		t.Errorf("Want err from RawVideoRequest when WriteBucketFile returns err, got nil")
	}
	fakeSSC.WriteBucketFileMock = func(bucketName cloud.BucketName, localFileNameAndPath, bucketFileName string) error {
		return nil
	}

	_, err = rawVidHandler.HandleRawVideoProcessRequest(request)
	if err != nil {
		t.Errorf("Want nil from HandleRawVideoProcessRequest err, got %v", err)
	}
}

func newRawVideoHandlerForTests() (*RawVideoHandler, *mp4.FakeMP4Tool, *cloud.FakeSimpleStorageClient, *rawvideodao.MockRawVideoDAO, *rawlocationdao.MockRawLocatinDAO, *rawmotiondao.MockRawMotionDAO, error) {
	fakeSimpleStorageClient := cloud.NewFakeSimpleStorageClient()
	fakeMP4Tool := mp4.NewFakeMP4Tool()
	localLogger := logging.NewLocalLogger(true /* silent */)

	mockRawVideoDAO := &rawvideodao.MockRawVideoDAO{}
	mockRawLocationDAO := &rawlocationdao.MockRawLocatinDAO{}
	mockRawMotionDAO := &rawmotiondao.MockRawMotionDAO{}

	rawVideoHandler, err := NewRawVideoHandler(fakeSimpleStorageClient, fakeMP4Tool, mockRawVideoDAO, mockRawLocationDAO, mockRawMotionDAO, localLogger, "")
	if err != nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("NewRawVideoHandler returns err: %v", err)
	}
	return rawVideoHandler, fakeMP4Tool, fakeSimpleStorageClient, mockRawVideoDAO, mockRawLocationDAO, mockRawMotionDAO, nil
}
