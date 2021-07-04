package syncer

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
	"time"
)

type intraSenecaRequestInterface interface {
	HandleRawVideoProcessRequest(req *st.RawVideoProcessRequest) (*st.RawVideoProcessResponse, error)
}

type UserClientFactory interface {
	New(user *st.User) (googledrive.GoogleDriveUserInterface, error)
}

type Syncer struct {
	intraSeneca   intraSenecaRequestInterface
	gdriveFactory UserClientFactory
	userDAO       dao.UserDAO
	logger        logging.LoggingInterface
}

func New(intraSeneca intraSenecaRequestInterface, gdriveFactory UserClientFactory, userDAO dao.UserDAO, logger logging.LoggingInterface) *Syncer {
	return &Syncer{
		intraSeneca:   intraSeneca,
		gdriveFactory: gdriveFactory,
		userDAO:       userDAO,
		logger:        logger,
	}
}

// ScanAllUsers scans all users' Google Drive folders for newly uploaded files.
func (sync *Syncer) ScanAllUsers() {
	sync.logger.Log(fmt.Sprintf("Scanning all users at %q", time.Now().String()))
	userIDs, err := sync.userDAO.ListAllUserIDs()
	if err != nil {
		sync.logger.Critical(fmt.Sprintf("Error listing all users - err: %v", err))
		return
	}

	for _, id := range userIDs {
		if err := sync.SyncUser(id); err != nil {
			sync.logger.Error(fmt.Sprintf("Error in sync.handleUser(%s) - err: %v", id, err))
		}
	}
}

func (sync *Syncer) SyncUser(id string) error {
	user, err := sync.userDAO.GetUserByID(id)
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

		err := func() error {
			fileInfo, err := userDriveClient.GetFileInfo(fid)
			if err != nil {
				return fmt.Errorf("userDriveClient.GetFileInfo(%s) for user %q returns err: %w", fid, user.Id, err)
			}
			if err := userDriveClient.MarkFileByID(fid, googledrive.WorkInProgress, false); err != nil {
				sync.logger.Error(fmt.Sprintf("Error MarkFileByID(%s, %s, false) for user %q returns err: %v", fid, googledrive.WorkInProgress, id, err))
			}
			pathToFile, err := userDriveClient.DownloadFileByID(fid)
			if err != nil {
				return fmt.Errorf("userDriveClient.DownloadFileByID(%s) for user %q returns err: %w", fid, user.Id, err)
			}

			rawVideoProcessRequest := &st.RawVideoProcessRequest{
				UserId:    id,
				LocalPath: pathToFile,
				VideoName: fileInfo.FileName,
			}

			if _, err = sync.intraSeneca.HandleRawVideoProcessRequest(rawVideoProcessRequest); err != nil {
				sync.logger.Error(fmt.Sprintf("Error in HandleRawVideoProcessRequest for user %q: %v", id, err))
				return err
			}

			return nil
		}()

		if err != nil {
			// TODO(lucaloncar): add DUPLICATE_ prefix
			if err := userDriveClient.MarkFileByID(fid, googledrive.Error, false); err != nil {
				sync.logger.Error(fmt.Sprintf("Error MarkFileByID(%s, %s, false) for user %q returns err: %v", fid, googledrive.Error, id, err))
			}
		} else {
			if err := userDriveClient.MarkFileByID(fid, googledrive.Success, false); err != nil {
				sync.logger.Error(fmt.Sprintf("Error MarkFileByID(%s, %s, false) for user %q returns err: %v", fid, googledrive.Success, id, err))
			}
		}
	}
	sync.logger.Log(fmt.Sprintf("Finished syncing user %q", user.Email))
	return nil
}
