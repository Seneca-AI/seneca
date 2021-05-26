package testenv

import (
	"context"
	"fmt"
	"seneca/internal/client/cloud"
	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/logging"
	"seneca/internal/controller/apiserver"
	"seneca/internal/controller/runner"
	"seneca/internal/controller/syncer"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/dao/userdao"
	"seneca/internal/dataaggregator/sanitizer"
	"seneca/internal/datagatherer/rawvideohandler"
	"seneca/internal/dataprocessor"
	"seneca/internal/util/mp4"
	"time"
)

type TestEnvironment struct {
	ProjectID           string
	Logger              logging.LoggingInterface
	SimpleStorage       cloud.SimpleStorageInterface
	UserDAO             dao.UserDAO
	RawVideoDAO         dao.RawVideoDAO
	RawLocationDAO      dao.RawLocationDAO
	RawMotionDAO        dao.RawMotionDAO
	TripDAO             dao.TripDAO
	EventDAO            dao.EventDAO
	DrivingConditionDAO dao.DrivingConditionDAO
	Syncer              *syncer.Syncer
	GDriveFactory       *googledrive.UserClientFactory
	Runner              *runner.Runner
	APIServer           *apiserver.APIServer
}

func New(projectID string, logger logging.LoggingInterface) (*TestEnvironment, error) {
	gcsc, err := gcp.NewGoogleCloudStorageClient(context.Background(), projectID, time.Second*10, time.Minute)
	if err != nil {
		return nil, fmt.Errorf("cloud.NewGoogleCloudStorageClient() returns - err: %w", err)
	}

	sqlService, err := datastore.New(context.Background(), projectID)
	if err != nil {
		return nil, fmt.Errorf("error initializing sql service - err: %w", err)
	}

	userDAO := userdao.NewSQLUserDAO(sqlService)
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlService, logger, time.Second*5)
	rawLocationDAO := rawlocationdao.NewSQLRawLocationDAO(sqlService)
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(sqlService, logger)
	tripDAO := tripdao.NewSQLTripDAO(sqlService, logger)
	eventDAO := eventdao.NewSQLEventDAO(sqlService, tripDAO, logger)
	dcDAO := drivingconditiondao.NewSQLDrivingConditionDAO(sqlService, tripDAO, eventDAO)

	mp4Tool, err := mp4.NewMP4Tool(logger)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("mp4.NewMP4Tool() returns - err: %v", err))
	}
	rawVideoHandler, err := rawvideohandler.NewRawVideoHandler(gcsc, mp4Tool, rawVideoDAO, rawLocationDAO, rawMotionDAO, logger, projectID)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("cloud.NewRawVideoHandler() returns - err: %v", err))
	}
	gDriveFactory := &googledrive.UserClientFactory{}
	syncer := syncer.New(rawVideoHandler, gDriveFactory, userDAO, logger)
	dataprocessor := dataprocessor.New(dataprocessor.GetCurrentAlgorithms(), eventDAO, dcDAO, rawMotionDAO, rawVideoDAO, logger)
	runner := runner.New(userDAO, dataprocessor, logger)
	sanitizer := sanitizer.New(rawMotionDAO, rawLocationDAO, rawVideoDAO, eventDAO, dcDAO)
	apiserver := apiserver.New(sanitizer, tripDAO)

	return &TestEnvironment{
		ProjectID:           projectID,
		Logger:              logger,
		SimpleStorage:       gcsc,
		UserDAO:             userDAO,
		RawVideoDAO:         rawVideoDAO,
		RawLocationDAO:      rawLocationDAO,
		RawMotionDAO:        rawMotionDAO,
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

			if err := te.SimpleStorage.DeleteBucketFile(cloud.RawVideoBucketName, rawVideo.CloudStorageFileName); err != nil {
				te.Logger.Warning(fmt.Sprintf("DeleteBucketFile(%s, %s) returns err: %v", cloud.RawVideoBucketName, rawVideo.CloudStorageFileName, err))
			}

			if err := te.RawVideoDAO.DeleteRawVideoByID(rvid); err != nil {
				te.Logger.Error(fmt.Sprintf("DeleteRawVideoByID(%s) returns err: %v", rvid, err))
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
		prefixes := []googledrive.FilePrefix{googledrive.WorkInProgress, googledrive.Error}
		for _, fid := range fileIDs {
			for _, prefix := range prefixes {
				if err := gDrive.MarkFileByID(fid, prefix, true); err != nil {
					te.Logger.Error(fmt.Sprintf("gDrive.MarkFileByID(%s, %s, true) for user %q returns err: %v", fid, prefix, user.Id, err))
				}
			}
		}

		rawMotionIDs, err := te.RawMotionDAO.ListUserRawMotionIDs(uid)
		if err != nil {
			te.Logger.Error(fmt.Sprintf("ListUserRawMotionIDs(%s) returns err: %v", uid, err))
		}
		for _, rmid := range rawMotionIDs {
			if err := te.RawMotionDAO.DeleteRawMotionByID(rmid); err != nil {
				te.Logger.Error(fmt.Sprintf("DeleteRawMotionByID(%s) returns err: %v", rmid, err))
			}
		}

		rawLocationIDs, err := te.RawLocationDAO.ListUserRawLocationIDs(uid)
		if err != nil {
			te.Logger.Error(fmt.Sprintf("ListUserRawLocationIDs(%s) returns err: %v", uid, err))
		}
		for _, rlid := range rawLocationIDs {
			if err := te.RawLocationDAO.DeleteRawLocationByID(rlid); err != nil {
				te.Logger.Error(fmt.Sprintf("DeleteRawLocationByID(%s) returns err: %v", rlid, err))
			}
		}

		tripIDs, err := te.TripDAO.ListUserTripIDs(uid)
		if err != nil {
			te.Logger.Error(fmt.Sprintf("ListUserTripIDs(%s) returns err: %v", uid, err))
		}
		for _, tid := range tripIDs {
			eventIDs, err := te.EventDAO.ListTripEventIDs(uid, tid)
			if err != nil {
				te.Logger.Error(fmt.Sprintf("ListTripEventIDs(%s, %s) returns err: %v", uid, tid, err))
			}
			for _, eid := range eventIDs {
				if err := te.EventDAO.DeleteEventByID(context.TODO(), uid, tid, eid); err != nil {
					te.Logger.Error(fmt.Sprintf("DeleteEventByID(%s, %s, %s) returns err: %v", uid, tid, eid, err))
				}
			}

			dcIDs, err := te.DrivingConditionDAO.ListTripDrivingConditionIDs(uid, tid)
			if err != nil {
				te.Logger.Error(fmt.Sprintf("DrivingConditionDAO(%s, %s) returns err: %v", uid, tid, err))
			}
			for _, dcid := range dcIDs {
				if err := te.DrivingConditionDAO.DeleteDrivingConditionByID(context.TODO(), uid, tid, dcid); err != nil {
					te.Logger.Error(fmt.Sprintf("DeleteDrivingConditionByID(%s, %s, %s) returns err: %v", uid, tid, dcid, err))
				}
			}

			if err := te.TripDAO.DeleteTripByID(context.TODO(), tid); err != nil {
				te.Logger.Error(fmt.Sprintf("DeleteTripByID(%s) returns err: %v", tid, err))
			}
		}
	}
}
