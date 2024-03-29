package testenv

import (
	"context"
	"fmt"
	"seneca/internal/client/cloud"
	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/client/database"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/intraseneca"
	"seneca/internal/client/intraseneca/http"
	"seneca/internal/client/logging"
	weatherservice "seneca/internal/client/weather/service"
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
	"seneca/internal/dataprocessor/algorithms"
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

func New(projectID string, serverConfig *intraseneca.ServerConfig, logger logging.LoggingInterface) (*TestEnvironment, error) {
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
	allDAOSet := &dao.AllDAOSet{
		UserDAO:             userDAO,
		RawVideoDAO:         rawVideoDAO,
		RawLocationDAO:      rawLocationDAO,
		RawMotionDAO:        rawMotionDAO,
		RawFrameDAO:         rawFrameDAO,
		TripDAO:             tripDAO,
		EventDAO:            eventDAO,
		DrivingConditionDAO: dcDAO,
	}

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

	intraSenecaClient, err := http.New(serverConfig)
	if err != nil {
		return nil, fmt.Errorf("intraseneca.http.New() returns err: %w", err)
	}

	algoFactory, err := algorithms.NewFactory(weatherservice.NewWeatherStackService(time.Second*10), intraSenecaClient)
	if err != nil {
		return nil, fmt.Errorf("algorithms.NewFactory() returns err: %v", err)
	}
	algos := []dataprocessor.AlgorithmInterface{}
	algoTags := []string{"00000", "00001", "00002", "00003"}
	for _, tag := range algoTags {
		algo, err := algoFactory.GetAlgorithm(tag)
		if err != nil {
			return nil, fmt.Errorf("GetAlgorithm(%q) returns err: %v", tag, err)
		}
		algos = append(algos, algo)
	}

	dataprocessor, err := dataprocessor.New(algos, allDAOSet, wrappedLogger)
	if err != nil {
		return nil, fmt.Errorf("dataprocessor.New() returns err: %w", err)
	}
	runner := runner.New(userDAO, dataprocessor, wrappedLogger)
	sanitizer := sanitizer.New(rawMotionDAO, rawLocationDAO, rawVideoDAO, rawFrameDAO, eventDAO, dcDAO)
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

		if err := data.DeleteAllUserData(uid, false, te.sqlService, te.SimpleStorage, te.Logger); err != nil {
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
