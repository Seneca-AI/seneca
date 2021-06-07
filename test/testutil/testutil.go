package testutil

import (
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
	"seneca/internal/dao/drivingconditiondao"
	"seneca/internal/dao/eventdao"
	"seneca/internal/dao/rawframedao"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/tripdao"
	"seneca/internal/dao/userdao"
	"time"
)

func GenerateAllDAOSetWithFakeDB(logger logging.LoggingInterface, rawVideoCreateTimeOffset time.Duration) *dao.AllDAOSet {
	sqlService := database.NewFake()

	userDAO := userdao.NewSQLUserDAO(sqlService)
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlService, logger, rawVideoCreateTimeOffset)
	rawLocationDAO := rawlocationdao.NewSQLRawLocationDAO(sqlService)
	rawFrameDAO := rawframedao.NewSQLRawFrameDAO(sqlService)
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(sqlService, logger)
	tripDAO := tripdao.NewSQLTripDAO(sqlService, logger)
	eventDAO := eventdao.NewSQLEventDAO(sqlService, tripDAO, logger)
	drivingConditionDAO := drivingconditiondao.NewSQLDrivingConditionDAO(sqlService, tripDAO, eventDAO)

	return &dao.AllDAOSet{
		UserDAO:             userDAO,
		RawVideoDAO:         rawVideoDAO,
		RawLocationDAO:      rawLocationDAO,
		RawFrameDAO:         rawFrameDAO,
		RawMotionDAO:        rawMotionDAO,
		TripDAO:             tripDAO,
		EventDAO:            eventDAO,
		DrivingConditionDAO: drivingConditionDAO,
	}
}
