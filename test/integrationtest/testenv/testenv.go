package testenv

import (
	"context"
	"fmt"
	"seneca/internal/client/cloud"
	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/cloud/gcp/datastore"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/dao/userdao"
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
	}, nil
}
