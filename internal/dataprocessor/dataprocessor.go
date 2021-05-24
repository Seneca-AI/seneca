package dataprocessor

import (
	"context"
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
)

const AlgosVersion = 0.01

type DataProcessor struct {
	algorithms          map[string]AlgorithmInterface
	rawMotionDAO        dao.RawMotionDAO
	rawVideoDAO         dao.RawVideoDAO
	eventDAO            dao.EventDAO
	drivingConditionDAO dao.DrivingConditionDAO
	logger              logging.LoggingInterface
}

type AlgorithmInterface interface {
	GenerateEvent(inputs []interface{}) (*st.EventInternal, error)
	GenerateDrivingCondition(inputs []interface{}) (*st.DrivingConditionInternal, error)
	Tag() string
}

func GetCurrentAlgorithms() []AlgorithmInterface {
	return []AlgorithmInterface{newAccelerationV0()}
}

func New(algorithmList []AlgorithmInterface, eventDAO dao.EventDAO, drivingConditionDAO dao.DrivingConditionDAO, rawMotionDAO dao.RawMotionDAO, rawVideoDAO dao.RawVideoDAO, logger logging.LoggingInterface) *DataProcessor {
	dp := &DataProcessor{
		algorithms:          map[string]AlgorithmInterface{},
		rawMotionDAO:        rawMotionDAO,
		rawVideoDAO:         rawVideoDAO,
		eventDAO:            eventDAO,
		drivingConditionDAO: drivingConditionDAO,
		logger:              logger,
	}

	for _, alg := range algorithmList {
		dp.algorithms[alg.Tag()] = alg
	}

	return dp
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

		event, err := alg.GenerateEvent([]interface{}{rawMotion})
		if err != nil {
			dp.logger.Error(fmt.Sprintf("GenerateEvent() returns err: %v", err))
		}
		if event == nil {
			continue
		}

		if _, err := dp.eventDAO.CreateEvent(context.TODO(), event); err != nil {
			dp.logger.Error(fmt.Sprintf("eventdao.CreateEvent() returns err: %v", err))
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
