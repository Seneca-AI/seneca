package rawvideohandler

import (
	"errors"
	"fmt"
	"net/http"
	"seneca/api/senecaerror"
	"seneca/internal/client/cloud"
	"seneca/internal/client/logging"
	"seneca/internal/util/mp4"
	"testing"
)

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
}

func newRawVideoHandlerForTests() (*RawVideoHandler, error) {
	fakeSimpleStorageClient := cloud.NewFakeSimpleStorageClient()
	fakeFakeNoSQLDBClient := cloud.NewFakeNoSQLDatabaseClient()
	fakeMP4Tool := mp4.NewFakeMP4Tool()
	localLogger := logging.NewLocalLogger(true /* silent */)

	rawVideoHandler, err := NewRawVideoHandler(fakeSimpleStorageClient, fakeFakeNoSQLDBClient, fakeMP4Tool, localLogger, "")
	if err != nil {
		return nil, fmt.Errorf("NewRawVideoHandler returns err: %v", err)
	}
	return rawVideoHandler, nil
}
