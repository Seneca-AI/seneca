package dataprocessor

import (
	"context"
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
)

const (
	AlgosVersion = 0.01
)

var (
	RawVideoTypeString    = fmt.Sprintf("%T", &st.RawVideo{})
	RawLocationTypeString = fmt.Sprintf("%T", &st.RawLocation{})
	RawMotionTypeString   = fmt.Sprintf("%T", &st.RawMotion{})
	RawFrameTypeString    = fmt.Sprintf("%T", &st.RawFrame{})
)

type DataProcessor struct {
	algorithms          map[string]AlgorithmInterface
	rawMotionDAO        dao.RawMotionDAO
	rawLocationDAO      dao.RawLocationDAO
	rawFrameDAO         dao.RawFrameDAO
	rawVideoDAO         dao.RawVideoDAO
	eventDAO            dao.EventDAO
	drivingConditionDAO dao.DrivingConditionDAO
	logger              logging.LoggingInterface
}

type AlgorithmInterface interface {
	GenerateEvents(inputs map[string][]interface{}) ([]*st.EventInternal, error)
	GenerateDrivingConditions(inputs map[string][]interface{}) ([]*st.DrivingConditionInternal, error)
	Tag() string
}

type AlgorithmFactoryInterface interface {
	GetAlgorithm(algoTag string) (AlgorithmInterface, error)
}

func New(algorithmList []AlgorithmInterface, allDaos *dao.AllDAOSet, logger logging.LoggingInterface) (*DataProcessor, error) {
	dp := &DataProcessor{
		algorithms:          map[string]AlgorithmInterface{},
		rawMotionDAO:        allDaos.RawMotionDAO,
		rawLocationDAO:      allDaos.RawLocationDAO,
		rawVideoDAO:         allDaos.RawVideoDAO,
		rawFrameDAO:         allDaos.RawFrameDAO,
		eventDAO:            allDaos.EventDAO,
		drivingConditionDAO: allDaos.DrivingConditionDAO,
		logger:              logger,
	}

	for _, algo := range algorithmList {
		dp.algorithms[algo.Tag()] = algo
	}

	return dp, nil
}

func (dp *DataProcessor) Run(userID string) {
	allUnprocessedData := map[string][]interface{}{
		RawVideoTypeString:    {},
		RawLocationTypeString: {},
		RawMotionTypeString:   {},
		RawFrameTypeString:    {},
	}

	unprocessedRawVideoIDs, err := dp.rawVideoDAO.ListUnprocessedRawVideoIDs(userID, AlgosVersion)
	if err != nil {
		dp.logger.Error(fmt.Sprintf("ListUnprocessedRawVideoIDs(%s, %f) returns err: %v", userID, AlgosVersion, err))
	}

	for _, urvid := range unprocessedRawVideoIDs {
		rawVideo, err := dp.rawVideoDAO.GetRawVideoByID(urvid)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GetRawVideoByID(%s) returns err: %v", urvid, err))
		}
		allUnprocessedData[RawVideoTypeString] = append(allUnprocessedData[RawVideoTypeString], rawVideo)
	}

	unprocessedRawMotionIDs, err := dp.rawMotionDAO.ListUnprocessedRawMotionIDs(userID, AlgosVersion)
	if err != nil {
		dp.logger.Error(fmt.Sprintf("ListUnprocessedRawMotionIDs(%s, %f) returns err: %v", userID, AlgosVersion, err))
	}

	for _, urmid := range unprocessedRawMotionIDs {
		rawMotion, err := dp.rawMotionDAO.GetRawMotionByID(urmid)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GetRawMotionByID(%s) returns err: %v", urmid, err))
		}
		allUnprocessedData[RawMotionTypeString] = append(allUnprocessedData[RawMotionTypeString], rawMotion)
	}

	locationIDs, err := dp.rawLocationDAO.ListUnprocessedRawLocationsIDs(userID, AlgosVersion)
	if err != nil {
		dp.logger.Error(fmt.Sprintf("ListUnprocessedRawLocationsIDs(%s, %f) returns err: %v", userID, AlgosVersion, err))
	}

	for _, lid := range locationIDs {
		rawLocation, err := dp.rawLocationDAO.GetRawLocationByID(lid)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GetRawLocationByID(%s) returns err: %v", lid, err))
			continue
		}
		allUnprocessedData[RawLocationTypeString] = append(allUnprocessedData[RawLocationTypeString], rawLocation)
	}

	unprocessedRawFrameIDs, err := dp.rawFrameDAO.ListUnprocessedRawFramesIDs(userID, AlgosVersion)
	if err != nil {
		dp.logger.Error(fmt.Sprintf("ListUnprocessedRawFramesIDs(%s, %f) returns err: %v", userID, AlgosVersion, err))
	}

	for _, rfid := range unprocessedRawFrameIDs {
		rawFrame, err := dp.rawFrameDAO.GetRawFrameByID(rfid)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GetRawFrameByID(%s) returns err: %v", rfid, err))
			continue
		}
		allUnprocessedData[RawFrameTypeString] = append(allUnprocessedData[RawFrameTypeString], rawFrame)
	}

	allEvents := []*st.EventInternal{}
	allDrivingConditions := []*st.DrivingConditionInternal{}

	for _, alg := range dp.algorithms {
		eventsFromAlgo, err := alg.GenerateEvents(allUnprocessedData)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GenerateEvents() for algo %q returns err: %v", alg.Tag(), err))
		}
		if eventsFromAlgo != nil {
			allEvents = append(allEvents, eventsFromAlgo...)
		}

		drivingConditionsFromAlg, err := alg.GenerateDrivingConditions(allUnprocessedData)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GenerateDrivingConditions() for algo %q returns err: %v", alg.Tag(), err))
		}
		if drivingConditionsFromAlg != nil {
			allDrivingConditions = append(allDrivingConditions, drivingConditionsFromAlg...)
		}
	}

	for _, event := range allEvents {
		if _, err := dp.eventDAO.CreateEvent(context.TODO(), event); err != nil {
			dp.logger.Error(fmt.Sprintf("CreateEvent() for user %q returns err: %v", userID, err))
		}
	}

	for _, drivingCondition := range allDrivingConditions {
		if _, err := dp.drivingConditionDAO.CreateDrivingCondition(context.TODO(), drivingCondition); err != nil {
			dp.logger.Error(fmt.Sprintf("CreateEvent() for user %q returns err: %v", userID, err))
		}
	}

	// Updated the algos version of all processed data.
	for _, rawVideoObj := range allUnprocessedData[RawVideoTypeString] {
		rawVideo, ok := rawVideoObj.(*st.RawVideo)
		if !ok {
			dp.logger.Error(fmt.Sprintf("Found a %T in map entry for %s", rawVideoObj, RawVideoTypeString))
			continue
		}
		rawVideo.AlgosVersion = AlgosVersion
		if err := dp.rawVideoDAO.PutRawVideoByID(context.TODO(), rawVideo.Id, rawVideo); err != nil {
			dp.logger.Error(fmt.Sprintf("PutRawVideoByID(%s) returns err: %v", rawVideo.Id, err))
		}
	}

	for _, rawLocationObj := range allUnprocessedData[RawLocationTypeString] {
		rawLocation, ok := rawLocationObj.(*st.RawLocation)
		if !ok {
			dp.logger.Error(fmt.Sprintf("Found a %T in map entry for %s", rawLocationObj, RawVideoTypeString))
			continue
		}
		rawLocation.AlgosVersion = AlgosVersion
		if err := dp.rawLocationDAO.PutRawLocationByID(context.TODO(), rawLocation.Id, rawLocation); err != nil {
			dp.logger.Error(fmt.Sprintf("PutRawVideoByID(%s) returns err: %v", rawLocation.Id, err))
		}
	}

	for _, rawMotionObj := range allUnprocessedData[RawMotionTypeString] {
		rawMotion, ok := rawMotionObj.(*st.RawMotion)
		if !ok {
			dp.logger.Error(fmt.Sprintf("Found a %T in map entry for %s", rawMotionObj, RawVideoTypeString))
			continue
		}
		rawMotion.AlgosVersion = AlgosVersion
		if err := dp.rawMotionDAO.PutRawMotionByID(context.TODO(), rawMotion.Id, rawMotion); err != nil {
			dp.logger.Error(fmt.Sprintf("PutRawVideoByID(%s) returns err: %v", rawMotion.Id, err))
		}
	}

	dp.logger.Log(fmt.Sprintf("Finished running dataprocessor on user with ID %q", userID))
}
