package syncer

import (
	"fmt"
	"log"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/logging"
	"testing"
)

func TestErrorHandling(t *testing.T) {
	fakeUserIDs := []string{"123", "456", "789"}

	logger := &logging.MockLogger{}
	syncer, intraSeneca, fakeGDrive, noSQLDB := newSyncerForTests(logger)

	callsMap := map[string]int{"critical": 0, "error": 0}
	logger.CriticalMock = func(message string) {
		callsMap["critical"]++
	}
	logger.ErrorMock = func(message string) {
		callsMap["error"]++
	}

	// Ensure we get a critical error if ListAllUserIDs returns an err.
	noSQLDB.ListAllUserIDsMock = func(pageToken string, maxResults int) ([]string, string, error) {
		return nil, "", fmt.Errorf("error")
	}
	syncer.ScanAllUsers()
	if callsMap["critical"] != 1 {
		t.Errorf("Want 1 call to logger.Critical, got %d", callsMap["critical"])
	}
	callsMap["critical"] = 0

	noSQLDB.ListAllUserIDsMock = func(pageToken string, maxResults int) ([]string, string, error) {
		return fakeUserIDs, "", nil
	}

	noSQLDB.GetUserByIDMock = func(id string) (*st.User, error) {
		return nil, fmt.Errorf("error")
	}
	syncer.ScanAllUsers()
	if callsMap["error"] != 3 {
		t.Errorf("Want 3 calls to logger.Error, got %d", callsMap["error"])
	}
	callsMap["error"] = 0

	noSQLDB.GetUserByIDMock = func(id string) (*st.User, error) {
		return &st.User{
			Id: id,
		}, nil
	}
	for _, fuid := range fakeUserIDs {
		fakeGDrive.InsertFakeClient(fuid, nil, fmt.Errorf("error"))
	}
	syncer.ScanAllUsers()
	if callsMap["error"] != 3 {
		t.Errorf("Want 3 calls to logger.Error, got %d", callsMap["error"])
	}
	callsMap["error"] = 0

	for _, fuid := range fakeUserIDs {
		fakeClient := &googledrive.FakeGoogleDriveUserClient{}
		fakeClient.ListFileIDsMock = func() ([]string, error) {
			return nil, fmt.Errorf("error")
		}
		fakeGDrive.InsertFakeClient(fuid, fakeClient, nil)
	}
	syncer.ScanAllUsers()
	if callsMap["error"] != 3 {
		t.Errorf("Want 3 calls to logger.Error, got %d", callsMap["error"])
	}
	callsMap["error"] = 0

	for _, fuid := range fakeUserIDs {
		fakeClient := &googledrive.FakeGoogleDriveUserClient{}
		fakeClient.ListFileIDsMock = func() ([]string, error) {
			fileIDs := []string{}
			for i := 0; i < 5; i++ {
				fileIDs = append(fileIDs, fmt.Sprintf("%s%d", fuid, i))
			}
			return fileIDs, nil
		}
		fakeClient.DownloadFileByIDMock = func(fileID string) (string, error) {
			return "", fmt.Errorf("error")
		}
		fakeGDrive.InsertFakeClient(fuid, fakeClient, nil)
	}
	syncer.ScanAllUsers()
	if callsMap["error"] != 3 {
		t.Errorf("Want 3 calls to logger.Error, got %d", callsMap["error"])
	}
	callsMap["error"] = 0

	numRequests := 0
	intraSeneca.HandleRawVideoProcessRequestMock = func(req *st.RawVideoProcessRequest) *st.RawVideoProcessResponse {
		numRequests++
		return &st.RawVideoProcessResponse{
			ErrorCode: 400,
		}
	}
	for _, fuid := range fakeUserIDs {
		fakeClient := &googledrive.FakeGoogleDriveUserClient{}
		fakeClient.ListFileIDsMock = func() ([]string, error) {
			fileIDs := []string{}
			for i := 0; i < 5; i++ {
				fileIDs = append(fileIDs, fmt.Sprintf("%s%d", fuid, i))
			}
			return fileIDs, nil
		}
		fakeClient.DownloadFileByIDMock = func(fileID string) (string, error) {
			return fileID, nil
		}
		fakeClient.MarkFileByIDMock = func(fileID string, failure bool) error {
			return fmt.Errorf("error")
		}
		fakeGDrive.InsertFakeClient(fuid, fakeClient, nil)
	}
	syncer.ScanAllUsers()
	if numRequests != 15 {
		t.Errorf("Want 15 calls to HandleRawVideoProcessRequest, got %d", numRequests)
	}
	if callsMap["error"] != 15 {
		t.Errorf("Want 15 calls to logger.Error, got %d", callsMap["error"])
	}
}

func TestUsesPageToken(t *testing.T) {
	logger := &logging.MockLogger{}
	syncer, _, _, noSQLDB := newSyncerForTests(logger)

	callsMap := map[string]int{"critical": 0, "error": 0}
	logger.CriticalMock = func(message string) {
		callsMap["critical"]++
	}
	logger.ErrorMock = func(message string) {
		callsMap["error"]++
	}

	userIDsListCalls := make(chan string, 5)
	userIDsListCalls <- "1"
	userIDsListCalls <- "2"
	userIDsListCalls <- "3"

	noSQLDB.ListAllUserIDsMock = func(pageToken string, maxResults int) ([]string, string, error) {
		if len(userIDsListCalls) == 0 {
			return []string{}, "", nil
		}

		return []string{<-userIDsListCalls}, "something", nil
	}

	noSQLDB.GetUserByIDMock = func(id string) (*st.User, error) {
		return nil, fmt.Errorf("error")
	}
	syncer.ScanAllUsers()
	if callsMap["error"] != 3 {
		t.Errorf("Want 3 calls to logger.Error, got %d", callsMap["error"])
	}
}

type fakeIntraSeneca struct {
	HandleRawVideoProcessRequestMock func(req *st.RawVideoProcessRequest) *st.RawVideoProcessResponse
}

func (fis *fakeIntraSeneca) HandleRawVideoProcessRequest(req *st.RawVideoProcessRequest) *st.RawVideoProcessResponse {
	if fis.HandleRawVideoProcessRequestMock == nil {
		log.Fatal("HandleRawVideoProcessRequestMock not set.")
	}
	return fis.HandleRawVideoProcessRequestMock(req)
}

func newSyncerForTests(logger logging.LoggingInterface) (*Syncer, *fakeIntraSeneca, *googledrive.FakeUserClientFactory, *cloud.FakeNoSQLDatabaseClient) {
	intraSeneca := &fakeIntraSeneca{}
	fakeGDrive := googledrive.NewFakeUserClientFactory()
	fakeNoSQL := &cloud.FakeNoSQLDatabaseClient{}

	fakeSyncer := &Syncer{
		intraSeneca:   intraSeneca,
		gdriveFactory: fakeGDrive,
		noSQLDB:       fakeNoSQL,
		logger:        logger,
	}
	return fakeSyncer, intraSeneca, fakeGDrive, fakeNoSQL
}
