package dataprocessor

import (
	"fmt"
	"seneca/api/senecaerror"
	st "seneca/api/type"
)

// First very naive implementation.
type accelerationV0 struct {
	tag string
}

func newAccelerationV0() *accelerationV0 {
	return &accelerationV0{
		tag: "00001",
	}
}

func (acc *accelerationV0) GenerateEvent(inputs []interface{}) (*st.EventInternal, error) {
	if len(inputs) != 1 {
		return nil, fmt.Errorf("input to accelerationV0 must be a single RawMotion")
	}

	rawMotion, ok := inputs[0].(*st.RawMotion)
	if !ok {
		return nil, fmt.Errorf("input to accelerationV0 must be a single RawMotion")
	}

	severity := 0
	if rawMotion.Motion.AccelerationMphS < 7.5 {
		return nil, nil
	} else if rawMotion.Motion.AccelerationMphS < 10 {
		severity = 10
	} else if rawMotion.Motion.AccelerationMphS < 15 {
		severity = 30
	} else if rawMotion.Motion.AccelerationMphS < 20 {
		severity = 70
	} else if rawMotion.Motion.AccelerationMphS < 25 {
		severity = 100
	} else {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("rawMotion.Motion.AccelerationMphS of %f is impossible", rawMotion.Motion.AccelerationMphS))
	}

	return &st.EventInternal{
		UserId:      rawMotion.UserId,
		EventType:   st.EventType_FAST_ACCELERATION,
		Value:       rawMotion.Motion.AccelerationMphS,
		Severity:    float64(severity),
		TimestampMs: rawMotion.TimestampMs,
		Source: &st.Source{
			SourceId:   rawMotion.Id,
			SourceType: st.Source_RAW_MOTION,
		},
		AlgoTag: acc.tag,
	}, nil
}

func (acc *accelerationV0) GenerateDrivingCondition(inputs []interface{}) (*st.DrivingConditionInternal, error) {
	return nil, nil
}

func (acc *accelerationV0) Tag() string {
	return acc.tag
}
