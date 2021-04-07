package syncer

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/logging"
)

const maxListResults = 10

type noSQLDBInterface interface {
	ListAllUserIDs(pageToken string, maxResults int) ([]string, string, error)
	GetUserByID(id string) (*st.User, error)
}

type intraSenecaRequestInterface interface {
	HandleRawVideoProcessRequest(req *st.RawVideoProcessRequest) *st.RawVideoProcessResponse
}

type UserClientFactory interface {
	New(user *st.User) (googledrive.GoogleDriveUserInterface, error)
}

type Syncer struct {
	intraSeneca   intraSenecaRequestInterface
	gdriveFactory UserClientFactory
	noSQLDB       noSQLDBInterface
	logger        logging.LoggingInterface
}

// ScanAllUsers scans all users' Google Drive folders for newly uploaded files.
func (sync *Syncer) ScanAllUsers() {
	nextPageToken := ""
	for {
		userIDs, nextPageToken, err := sync.noSQLDB.ListAllUserIDs(nextPageToken, maxListResults)
		if err != nil {
			sync.logger.Critical(fmt.Sprintf("Error listing all users - err: %v", err))
			return
		}

		for _, id := range userIDs {
			if err := sync.handleUser(id); err != nil {
				sync.logger.Error(fmt.Sprintf("Error in sync.handleUser(%s) - err: %v", id, err))
			}
		}

		if nextPageToken == "" {
			break
		}
	}
}

func (sync *Syncer) handleUser(id string) error {
	user, err := sync.noSQLDB.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("GetUserByID(%s) returns err: %w", id, err)
	}

	userDriveClient, err := sync.gdriveFactory.New(user)
	if err != nil {
		return fmt.Errorf("error initializing NewGoogleDriveUserClient - err: %w", err)
	}

	fileIDs, err := userDriveClient.ListFileIDs()
	if err != nil {
		return fmt.Errorf("ListFileIDs() returns err: %w", err)
	}

	for _, fid := range fileIDs {
		pathToFile, err := userDriveClient.DownloadFileByID(fid)
		if err != nil {
			return fmt.Errorf("userDriveClient.DownloadFileByID(%s) returns err: %w", fid, err)
		}
		response := sync.intraSeneca.HandleRawVideoProcessRequest(&st.RawVideoProcessRequest{
			UserId:    id,
			LocalPath: pathToFile,
		})
		if response.ErrorCode != 200 {
			if err := userDriveClient.MarkFileByID(fid, true); err != nil {
				sync.logger.Error(fmt.Sprintf("Error MarkFileByID(%s, true) for user %q returns err: %v", fid, id, err))
			}
		}
	}
	return nil
}
