package syncer

import (
	"fmt"
	"log"
	st "seneca/api/type"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/logging"
	"seneca/internal/dao/userdao"
	"seneca/test/testutil"
	"testing"
)

func TestErrorHandling(t *testing.T) {
	fakeUserIDs := []string{testutil.TestUserID, "456", "789"}

	logger := &logging.MockLogger{}
	syncer, intraSeneca, fakeGDrive, mockUserDAO := newSyncerForTests(logger)

	callsMap := map[string]int{"critical": 0, "error": 0}
	logger.CriticalMock = func(message string) {
		fmt.Printf("Critical: %q\n", message)
		callsMap["critical"]++
	}
	logger.ErrorMock = func(message string) {
		fmt.Printf("Error: %q\n", message)
		callsMap["error"]++
	}
	logger.LogMock = func(message string) {
		fmt.Printf("Log: %q\n", message)
		callsMap["log"]++
	}

	// Ensure we get a critical error if ListAllUserIDs returns an err.
	mockUserDAO.ListAllUserIDsMock = func() ([]string, error) {
		return nil, fmt.Errorf("error")
	}
	syncer.ScanAllUsers()
	if callsMap["critical"] != 1 {
		t.Errorf("Want 1 call to logger.Critical, got %d", callsMap["critical"])
	}
	callsMap["critical"] = 0

	mockUserDAO.ListAllUserIDsMock = func() ([]string, error) {
		return fakeUserIDs, nil
	}

	mockUserDAO.GetUserByIDMock = func(id string) (*st.User, error) {
		return nil, fmt.Errorf("error")
	}
	syncer.ScanAllUsers()
	if callsMap["error"] != 3 {
		t.Errorf("Want 3 calls to logger.Error, got %d", callsMap["error"])
	}
	callsMap["error"] = 0

	mockUserDAO.GetUserByIDMock = func(id string) (*st.User, error) {
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
		fakeClient.ListFileIDsMock = func(gdQuery googledrive.GDriveQuery) ([]string, error) {
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
		fakeClient.ListFileIDsMock = func(gdQuery googledrive.GDriveQuery) ([]string, error) {
			fileIDs := []string{}
			for i := 0; i < 5; i++ {
				fileIDs = append(fileIDs, fmt.Sprintf("%s%d", fuid, i))
			}
			return fileIDs, nil
		}
		fakeClient.DownloadFileByIDMock = func(fileID string) (string, error) {
			return "", fmt.Errorf("error")
		}
		fakeClient.GetFileInfoMock = func(fileID string) (*googledrive.FileInfo, error) {
			return nil, fmt.Errorf("error")
		}
		fakeGDrive.InsertFakeClient(fuid, fakeClient, nil)
	}
	syncer.ScanAllUsers()
	if callsMap["error"] != 15 {
		t.Errorf("Want 15 calls to logger.Error, got %d", callsMap["error"])
	}
	callsMap["error"] = 0

	numRequests := 0
	intraSeneca.HandleRawVideoProcessRequestMock = func(req *st.RawVideoProcessRequest) (*st.RawVideoProcessResponse, error) {
		numRequests++
		return nil, fmt.Errorf("error")
	}
	for _, fuid := range fakeUserIDs {
		fakeClient := &googledrive.FakeGoogleDriveUserClient{}
		fakeClient.ListFileIDsMock = func(gdQuery googledrive.GDriveQuery) ([]string, error) {
			fileIDs := []string{}
			for i := 0; i < 5; i++ {
				fileIDs = append(fileIDs, fmt.Sprintf("%s%d", fuid, i))
			}
			return fileIDs, nil
		}
		fakeClient.DownloadFileByIDMock = func(fileID string) (string, error) {
			return fileID, nil
		}
		fakeClient.GetFileInfoMock = func(fileID string) (*googledrive.FileInfo, error) {
			return &googledrive.FileInfo{}, nil
		}
		fakeClient.MarkFileByIDMock = func(fileID string, prefix googledrive.FilePrefix, remove bool) error {
			return fmt.Errorf("error")
		}
		fakeGDrive.InsertFakeClient(fuid, fakeClient, nil)
	}
	syncer.ScanAllUsers()
	if numRequests != 15 {
		t.Errorf("Want 15 calls to HandleRawVideoProcessRequest, got %d", numRequests)
	}
	if callsMap["error"] != 45 {
		t.Errorf("Want 30 calls to logger.Error, got %d", callsMap["error"])
	}
}

type fakeIntraSeneca struct {
	HandleRawVideoProcessRequestMock func(req *st.RawVideoProcessRequest) (*st.RawVideoProcessResponse, error)
}

func (fis *fakeIntraSeneca) HandleRawVideoProcessRequest(req *st.RawVideoProcessRequest) (*st.RawVideoProcessResponse, error) {
	if fis.HandleRawVideoProcessRequestMock == nil {
		log.Fatal("HandleRawVideoProcessRequestMock not set.")
	}
	return fis.HandleRawVideoProcessRequestMock(req)
}

func newSyncerForTests(logger logging.LoggingInterface) (*Syncer, *fakeIntraSeneca, *googledrive.FakeUserClientFactory, *userdao.MockUserDAO) {
	intraSeneca := &fakeIntraSeneca{}
	fakeGDrive := googledrive.NewFakeUserClientFactory()
	mockUserDAO := &userdao.MockUserDAO{}

	fakeSyncer := &Syncer{
		intraSeneca:   intraSeneca,
		gdriveFactory: fakeGDrive,
		userDAO:       mockUserDAO,
		logger:        logger,
	}
	return fakeSyncer, intraSeneca, fakeGDrive, mockUserDAO
}
