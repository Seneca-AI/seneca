package testenv

import (
	"context"
	"fmt"
	"seneca/internal/client/cloud"
	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/client/database"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/logging"
	"seneca/internal/controller/apiserver"
	"seneca/internal/controller/runner"
	"seneca/internal/controller/syncer"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/rawframedao"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/dao/userdao"
	"seneca/internal/dataaggregator/sanitizer"
	"seneca/internal/datagatherer/rawvideohandler"
	"seneca/internal/dataprocessor"
	"seneca/internal/util/data"
	"seneca/internal/util/mp4"
	"time"
)

type TestEnvironment struct {
	ProjectID           string
	Logger              *ErrorCounterLogWrapper
	sqlService          database.SQLInterface
	SimpleStorage       cloud.SimpleStorageInterface
	UserDAO             dao.UserDAO
	RawVideoDAO         dao.RawVideoDAO
	RawLocationDAO      dao.RawLocationDAO
	RawMotionDAO        dao.RawMotionDAO
	RawFrameDAO         dao.RawFrameDAO
	TripDAO             dao.TripDAO
	EventDAO            dao.EventDAO
	DrivingConditionDAO dao.DrivingConditionDAO
	Syncer              *syncer.Syncer
	GDriveFactory       *googledrive.UserClientFactory
	Runner              *runner.Runner
	APIServer           *apiserver.APIServer
}

func New(projectID string, logger logging.LoggingInterface) (*TestEnvironment, error) {
	wrappedLogger := NewErrorCounterLogWrapper(logger)

	gcsc, err := gcp.NewGoogleCloudStorageClient(context.Background(), projectID, time.Second*10, time.Minute)
	if err != nil {
		return nil, fmt.Errorf("cloud.NewGoogleCloudStorageClient() returns - err: %w", err)
	}

	sqlService, err := datastore.New(context.Background(), projectID)
	if err != nil {
		return nil, fmt.Errorf("error initializing sql service - err: %w", err)
	}

	userDAO := userdao.NewSQLUserDAO(sqlService)
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlService, wrappedLogger, time.Second*5)
	rawLocationDAO := rawlocationdao.NewSQLRawLocationDAO(sqlService)
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(sqlService, wrappedLogger)
	rawFrameDAO := rawframedao.NewSQLRawFrameDAO(sqlService)
	tripDAO := tripdao.NewSQLTripDAO(sqlService, wrappedLogger)
	eventDAO := eventdao.NewSQLEventDAO(sqlService, tripDAO, wrappedLogger)
	dcDAO := drivingconditiondao.NewSQLDrivingConditionDAO(sqlService, tripDAO, eventDAO)

	mp4Tool, err := mp4.NewMP4Tool(wrappedLogger)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("mp4.NewMP4Tool() returns - err: %v", err))
	}
	rawVideoHandler, err := rawvideohandler.NewRawVideoHandler(gcsc, mp4Tool, rawVideoDAO, rawLocationDAO, rawMotionDAO, rawFrameDAO, wrappedLogger, projectID)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("cloud.NewRawVideoHandler() returns - err: %v", err))
	}
	gDriveFactory := &googledrive.UserClientFactory{}
	syncer := syncer.New(rawVideoHandler, gDriveFactory, userDAO, wrappedLogger)
	dataprocessor, err := dataprocessor.New(nil, eventDAO, dcDAO, rawMotionDAO, rawLocationDAO, rawVideoDAO, wrappedLogger)
	if err != nil {
		return nil, fmt.Errorf("dataprocessor.New() returns err: %w", err)
	}
	runner := runner.New(userDAO, dataprocessor, wrappedLogger)
	sanitizer := sanitizer.New(rawMotionDAO, rawLocationDAO, rawVideoDAO, eventDAO, dcDAO)
	apiserver := apiserver.New(sanitizer, tripDAO)

	return &TestEnvironment{
		ProjectID:           projectID,
		Logger:              wrappedLogger,
		sqlService:          sqlService,
		SimpleStorage:       gcsc,
		UserDAO:             userDAO,
		RawVideoDAO:         rawVideoDAO,
		RawLocationDAO:      rawLocationDAO,
		RawMotionDAO:        rawMotionDAO,
		RawFrameDAO:         rawFrameDAO,
		TripDAO:             tripDAO,
		EventDAO:            eventDAO,
		DrivingConditionDAO: dcDAO,
		GDriveFactory:       &googledrive.UserClientFactory{},
		Syncer:              syncer,
		Runner:              runner,
		APIServer:           apiserver,
	}, nil
}

// Clean the database of everything except for users themselves.
func (te *TestEnvironment) Clean() {
	userIDs, err := te.UserDAO.ListAllUserIDs()
	if err != nil {
		te.Logger.Error(fmt.Sprintf("ListAllUserIDs returns err: %v", err))
	}

	for _, uid := range userIDs {
		user, err := te.UserDAO.GetUserByID(uid)
		if err != nil {
			te.Logger.Error(fmt.Sprintf("GetUserByID(%s) returns err: %v", uid, err))
		}

		rawVideoIDs, err := te.RawVideoDAO.ListUserRawVideoIDs(uid)
		if err != nil {
			te.Logger.Error(fmt.Sprintf("ListUserRawVideoIDs(%s) returns err: %v", uid, err))
		}
		for _, rvid := range rawVideoIDs {
			rawVideo, err := te.RawVideoDAO.GetRawVideoByID(rvid)
			if err != nil {
				te.Logger.Error(fmt.Sprintf("GetRawVideoByID(%s) returns err: %v", rvid, err))
			}

			if rawVideo.CloudStorageFileName == "" {
				continue
			}

			bucketName, fileName, err := data.GCSURLToBucketNameAndFileName(rawVideo.CloudStorageFileName)
			if err != nil {
				te.Logger.Error(fmt.Sprintf("GCSURLToBucketNameAndFileName(%s) returns err: %v", rawVideo.CloudStorageFileName, err))
				continue
			}

			if err := te.SimpleStorage.DeleteBucketFile(bucketName, fileName); err != nil {
				te.Logger.Warning(fmt.Sprintf("DeleteBucketFile(%s, %s) returns err: %v", bucketName, fileName, err))
			}
		}

		rawFrameIDs, err := te.RawFrameDAO.ListUserRawFrameIDs(uid)
		if err != nil {
			te.Logger.Error(fmt.Sprintf("ListUserRawFrameIDs(%s) returns err: %v", uid, err))
		}

		for _, rfid := range rawFrameIDs {
			rawFrame, err := te.RawFrameDAO.GetRawFrameByID(rfid)
			if err != nil {
				te.Logger.Error(fmt.Sprintf("GetRawFrameByID(%s) returns err: %v", rfid, err))
			}

			if rawFrame.CloudStorageFileName == "" {
				continue
			}

			bucketName, fileName, err := data.GCSURLToBucketNameAndFileName(rawFrame.CloudStorageFileName)
			if err != nil {
				te.Logger.Error(fmt.Sprintf("GCSURLToBucketNameAndFileName(%s) returns err: %v", rawFrame.CloudStorageFileName, err))
				continue
			}

			if err := te.SimpleStorage.DeleteBucketFile(bucketName, fileName); err != nil {
				te.Logger.Warning(fmt.Sprintf("DeleteBucketFile(%s, %s) returns err: %v", bucketName, fileName, err))
			}
		}

		gDrive, err := te.GDriveFactory.New(user)
		if err != nil {
			te.Logger.Error(fmt.Sprintf("Error initializing gdrive client for user %q", user.Id))
			return
		}
		fileIDs, err := gDrive.ListFileIDs(googledrive.AllMP4s)
		if err != nil {
			te.Logger.Error(fmt.Sprintf("Error listing all file IDs for user %q", user.Id))
		}
		for _, fid := range fileIDs {
			for _, prefix := range googledrive.FilePrefixes {
				if err := gDrive.MarkFileByID(fid, prefix, true); err != nil {
					te.Logger.Error(fmt.Sprintf("gDrive.MarkFileByID(%s, %s, true) for user %q returns err: %v", fid, prefix, user.Id, err))
				}
			}
		}

		if err := data.DeleteAllUserDataInDB(uid, false, te.sqlService); err != nil {
			te.Logger.Error(fmt.Sprintf("DeleteAllUserDataInDB(%s, %t, _) returns err: %v", uid, false, err))
		}
	}
}

// ErrorCounterLogWrapper counts how many calls to Error() and Critical() there were.
type ErrorCounterLogWrapper struct {
	logger   logging.LoggingInterface
	failures int
}

func NewErrorCounterLogWrapper(logger logging.LoggingInterface) *ErrorCounterLogWrapper {
	return &ErrorCounterLogWrapper{
		logger:   logger,
		failures: 0,
	}
}

func (foe *ErrorCounterLogWrapper) Log(message string) {
	foe.logger.Log(message)
}

func (foe *ErrorCounterLogWrapper) Warning(message string) {
	foe.logger.Warning(message)
}

func (foe *ErrorCounterLogWrapper) Error(message string) {
	foe.failures++
	foe.logger.Error(message)
}

func (foe *ErrorCounterLogWrapper) Critical(message string) {
	foe.failures++
	foe.logger.Critical(message)
}

func (foe *ErrorCounterLogWrapper) Failures() int {
	return foe.failures
}
