package dataprocessor

import (
	"context"
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/logging"
	"seneca/internal/client/weather/service"
	"seneca/internal/dao"
	"time"
)

const (
	AlgosVersion = 0.01
)

type DataProcessor struct {
	algorithms          map[string]AlgorithmInterface
	rawMotionDAO        dao.RawMotionDAO
	rawLocationDAO      dao.RawLocationDAO
	rawVideoDAO         dao.RawVideoDAO
	eventDAO            dao.EventDAO
	drivingConditionDAO dao.DrivingConditionDAO
	logger              logging.LoggingInterface
}

type AlgorithmInterface interface {
	GenerateEvents(inputs []interface{}) ([]*st.EventInternal, error)
	GenerateDrivingConditions(inputs []interface{}) ([]*st.DrivingConditionInternal, error)
	Tag() string
}

func New(
	customAlgorithmList []AlgorithmInterface,
	eventDAO dao.EventDAO,
	drivingConditionDAO dao.DrivingConditionDAO,
	rawMotionDAO dao.RawMotionDAO,
	rawLocationDAO dao.RawLocationDAO,
	rawVideoDAO dao.RawVideoDAO,
	logger logging.LoggingInterface,
) (*DataProcessor, error) {
	if len(customAlgorithmList) == 0 {
		accAlgo, err := newAccelerationV0()
		if err != nil {
			return nil, fmt.Errorf("newAccelerationV0() returns err: %w", err)
		}
		decAlgo, err := newDecelerationV0()
		if err != nil {
			return nil, fmt.Errorf("newDecelerationV0() returns err: %w", err)
		}

		customAlgorithmList = []AlgorithmInterface{accAlgo, decAlgo, newWeatherV0(service.NewWeatherStackService(time.Second * 60))}
	}

	dp := &DataProcessor{
		algorithms:          map[string]AlgorithmInterface{},
		rawMotionDAO:        rawMotionDAO,
		rawLocationDAO:      rawLocationDAO,
		rawVideoDAO:         rawVideoDAO,
		eventDAO:            eventDAO,
		drivingConditionDAO: drivingConditionDAO,
		logger:              logger,
	}

	for _, alg := range customAlgorithmList {
		dp.algorithms[alg.Tag()] = alg
	}

	return dp, nil
}

func (dp *DataProcessor) Run(userID string) {
	unprocessedRawVideoIDs, err := dp.rawVideoDAO.ListUnprocessedRawVideoIDs(userID, AlgosVersion)
	if err != nil {
		dp.logger.Error(fmt.Sprintf("ListUnprocessedRawVideoIDs(%s, %f) returns err: %v", userID, AlgosVersion, err))
	}

	for _, urvid := range unprocessedRawVideoIDs {
		rawVideo, err := dp.rawVideoDAO.GetRawVideoByID(urvid)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GetRawVideoByID(%s) returns err: %v", urvid, err))
		}
		dp.ProcessRawVideo(rawVideo)
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
		dp.ProcessRawMotion(rawMotion)
	}

	dp.ProcessRawLocations(userID)
}

func (dp *DataProcessor) ProcessRawLocations(userID string) {
	// TODO(lucaloncar): utilize next page token here
	locationIDs, err := dp.rawLocationDAO.ListUnprocessedRawLocationsIDs(userID, AlgosVersion)
	if err != nil {
		dp.logger.Error(fmt.Sprintf("ListUnprocessedRawLocationsIDs(%s, %f) returns err: %v", userID, AlgosVersion, err))
		return
	}

	locations := []interface{}{}
	for _, lid := range locationIDs {
		rawLocation, err := dp.rawLocationDAO.GetRawLocationByID(lid)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GetRawLocationByID(%s) returns err: %v", lid, err))
			continue
		}
		locations = append(locations, rawLocation)
	}

	drivingConditions := []*st.DrivingConditionInternal{}
	algosRan := []string{}
	for _, alg := range dp.algorithms {
		drivingConditionsFromAlg, err := alg.GenerateDrivingConditions(locations)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GenerateDrivingConditions() for algo %q returns err: %v", alg.Tag(), err))
			continue
		}
		if drivingConditionsFromAlg != nil {
			algosRan = append(algosRan, alg.Tag())
			drivingConditions = append(drivingConditions, drivingConditionsFromAlg...)
		}
	}

	for _, dc := range drivingConditions {
		if _, err := dp.drivingConditionDAO.CreateDrivingCondition(context.TODO(), dc); err != nil {
			dp.logger.Error(fmt.Sprintf("CreateDrivingCondition() for user %q returns err: %v", userID, err))
		}
	}

	// Update the algo tags.
	for _, loc := range locations {
		rawLocation, ok := loc.(*st.RawLocation)
		if !ok {
			dp.logger.Critical(fmt.Sprintf("DeveloperError: want *RawLocation, got %T", loc))
		}
		rawLocation.AlgoTag = append(rawLocation.AlgoTag, algosRan...)
		rawLocation.AlgosVersion = AlgosVersion
		if err := dp.rawLocationDAO.PutRawLocationByID(context.TODO(), rawLocation.Id, rawLocation); err != nil {
			dp.logger.Error(fmt.Sprintf("PutRawLocationByID(_, %s, _) returns err: %v", rawLocation.Id, err))
		}
	}
}

func (dp *DataProcessor) ProcessRawMotion(rawMotion *st.RawMotion) {
	if rawMotion.AlgosVersion >= AlgosVersion {
		return
	}

	rawMotion.AlgosVersion = AlgosVersion

	for _, alg := range dp.algorithms {
		for _, ranTag := range rawMotion.AlgoTag {
			if ranTag == alg.Tag() {
				return
			}
		}

		events, err := alg.GenerateEvents([]interface{}{rawMotion})
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GenerateEvents() returns err: %v", err))
		}
		if events == nil {
			continue
		}

		for _, event := range events {
			if _, err := dp.eventDAO.CreateEvent(context.TODO(), event); err != nil {
				dp.logger.Error(fmt.Sprintf("eventdao.CreateEvent() returns err: %v", err))
			}
		}

		rawMotion.AlgoTag = append(rawMotion.AlgoTag, alg.Tag())

		// DrivingConditions are not processed for RawMotions.
	}

	// Update the algo tags.
	if err := dp.rawMotionDAO.PutRawMotionByID(context.TODO(), rawMotion.Id, rawMotion); err != nil {
		dp.logger.Error(fmt.Sprintf("PutRawMotionByID(%s) returns err: %v", rawMotion.Id, err))
	}
}

func (dp *DataProcessor) ProcessRawVideo(rawVideo *st.RawVideo) {
	// Just create an 'empty' driving condition for now.
	drivingCondition := &st.DrivingConditionInternal{
		UserId:        rawVideo.UserId,
		ConditionType: st.ConditionType_NONE_CONDITION_TYPE,
		StartTimeMs:   rawVideo.CreateTimeMs,
		EndTimeMs:     rawVideo.CreateTimeMs + rawVideo.DurationMs,
		Source: &st.Source{
			SourceId:   rawVideo.Id,
			SourceType: st.Source_RAW_VIDEO,
		},
		AlgoTag: "",
	}

	if _, err := dp.drivingConditionDAO.CreateDrivingCondition(context.TODO(), drivingCondition); err != nil {
		dp.logger.Error(fmt.Sprintf("CreateDrivingCondition() returns err: %v", err))
	}

	rawVideo.AlgosVersion = AlgosVersion

	if err := dp.rawVideoDAO.PutRawVideoByID(context.TODO(), rawVideo.Id, rawVideo); err != nil {
		dp.logger.Error(fmt.Sprintf("PutRawVideoByID(%s) returns err: %v", rawVideo.Id, err))

	}
}
