package algorithms

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/dataprocessor"
)

// Just creates NONE driving conditions for RawVideos.
type base struct {
	tag string
}

func newBase() *base {
	return &base{
		tag: "00000",
	}
}

func (bs *base) GenerateEvents(inputs map[string][]interface{}) ([]*st.EventInternal, error) {
	return nil, nil
}

func (bs *base) GenerateDrivingConditions(inputs map[string][]interface{}) ([]*st.DrivingConditionInternal, error) {
	rawVideoObjs := inputs[dataprocessor.RawVideoTypeString]
	if len(rawVideoObjs) == 0 {
		return nil, nil
	}

	drivingConditions := []*st.DrivingConditionInternal{}

	for _, rawVideoObj := range rawVideoObjs {
		rawVideo, ok := rawVideoObj.(*st.RawVideo)
		if !ok {
			return nil, fmt.Errorf("found a %T in map entry for %s", rawVideoObj, dataprocessor.RawVideoTypeString)
		}

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

		drivingConditions = append(drivingConditions, drivingCondition)
	}

	return drivingConditions, nil
}

func (bs *base) Tag() string {
	return bs.tag
}
