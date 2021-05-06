package syncer

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/logging"
	"time"
)

const maxListResults = 10

type noSQLDBInterface interface {
	ListAllUserIDs(pageToken string, maxResults int) ([]string, string, error)
	GetUserByID(id string) (*st.User, error)
}

type intraSenecaRequestInterface interface {
	HandleRawVideoProcessRequest(req *st.RawVideoProcessRequest) (*st.RawVideoProcessResponse, error)
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

func New(intraSeneca intraSenecaRequestInterface, gdriveFactory UserClientFactory, noSQLDB noSQLDBInterface, logger logging.LoggingInterface) *Syncer {
	return &Syncer{
		intraSeneca:   intraSeneca,
		gdriveFactory: gdriveFactory,
		noSQLDB:       noSQLDB,
		logger:        logger,
	}
}

// ScanAllUsers scans all users' Google Drive folders for newly uploaded files.
func (sync *Syncer) ScanAllUsers() {
	sync.logger.Log(fmt.Sprintf("Scanning all users at %q", time.Now().String()))

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

	fileIDs, err := userDriveClient.ListFileIDs(googledrive.UnprocessedMP4s)
	if err != nil {
		return fmt.Errorf("ListFileIDs() returns err: %w", err)
	}

	sync.logger.Log(fmt.Sprintf("User with ID %q has %d files to process.", id, len(fileIDs)))

	for _, fid := range fileIDs {
		pathToFile, err := userDriveClient.DownloadFileByID(fid)
		if err != nil {
			return fmt.Errorf("userDriveClient.DownloadFileByID(%s) returns err: %w", fid, err)
		}
		_, err = sync.intraSeneca.HandleRawVideoProcessRequest(&st.RawVideoProcessRequest{
			UserId:    id,
			LocalPath: pathToFile,
		})
		if err != nil {
			sync.logger.Error(fmt.Sprintf("Error in HandleRawVideoProcessRequest for user %q: %v", id, err))
			if err := userDriveClient.MarkFileByID(fid, googledrive.Error, false); err != nil {
				sync.logger.Error(fmt.Sprintf("Error MarkFileByID(%s, %s, false) for user %q returns err: %v", fid, googledrive.Error, id, err))
			}
			continue
		}
		if err := userDriveClient.MarkFileByID(fid, googledrive.WorkInProgress, false); err != nil {
			sync.logger.Error(fmt.Sprintf("Error MarkFileByID(%s, %s, false) for user %q returns err: %v", fid, googledrive.WorkInProgress, id, err))
		}
	}
	return nil
}
