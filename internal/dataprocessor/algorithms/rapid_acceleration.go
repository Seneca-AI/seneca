package algorithms

import (
	"fmt"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/dataprocessor"
	"seneca/internal/util"
)

type decelerationV0 struct {
	tag      string
	rangeMap *util.RangeMap
}

func newDecelerationV0() (*decelerationV0, error) {
	keys := []util.Range{
		{L: -256, U: -128},
		{L: -128, U: -64},
		{L: -64, U: -32},
		{L: -32, U: -16},
		{L: -16, U: -8},
		{L: -8, U: 9999999},
	}
	values := []interface{}{
		100.0,
		80.0,
		50.0,
		20.0,
		10.0,
		0.0,
	}

	rangeMap, err := util.NewRangeMap(keys, values)
	if err != nil {
		return nil, senecaerror.NewDevError(fmt.Errorf("NewRangeMap() returns err: %w", err))
	}

	return &decelerationV0{
		tag:      "00002",
		rangeMap: rangeMap,
	}, nil
}

func (dec *decelerationV0) GenerateEvents(inputs map[string][]interface{}) ([]*st.EventInternal, error) {
	motionsObjs := inputs[dataprocessor.RawMotionTypeString]
	if len(motionsObjs) == 0 {
		return nil, nil
	}

	events := []*st.EventInternal{}

	for _, motionObj := range motionsObjs {
		rawMotion, ok := motionObj.(*st.RawMotion)
		if !ok {
			return nil, fmt.Errorf("found a %T in map entry for %s", motionObj, dataprocessor.RawMotionTypeString)
		}

		valObj, ok := dec.rangeMap.Get(int64(rawMotion.Motion.AccelerationMphS))
		if !ok {
			return nil, fmt.Errorf("accelerationMPHS of %f for raw motion is impossible", rawMotion.Motion.AccelerationMphS)
		}
		severity, ok := valObj.(float64)
		if !ok {
			return nil, senecaerror.NewDevError(fmt.Errorf("trying to get float64 from rangeMap even though %T was inserted", valObj))
		}
		if severity == 0.0 {
			continue
		}

		events = append(events,
			&st.EventInternal{
				UserId:      rawMotion.UserId,
				EventType:   st.EventType_FAST_DECELERATION,
				Value:       rawMotion.Motion.AccelerationMphS,
				Severity:    severity,
				TimestampMs: rawMotion.TimestampMs,
				Source: &st.Source{
					SourceId:   rawMotion.Id,
					SourceType: st.Source_RAW_MOTION,
				},
				AlgoTag: dec.tag,
			})

	}
	return events, nil
}

func (dec *decelerationV0) GenerateDrivingConditions(inputs map[string][]interface{}) ([]*st.DrivingConditionInternal, error) {
	return nil, nil
}

func (dec *decelerationV0) Tag() string {
	return dec.tag
}

// First very naive implementation.
type accelerationV0 struct {
	tag      string
	rangeMap *util.RangeMap
}

func newAccelerationV0() (*accelerationV0, error) {
	keys := []util.Range{
		{L: -9999999, U: 8},
		{L: 8, U: 10},
		{L: 10, U: 15},
		{L: 15, U: 20},
		{L: 20, U: 25},
	}
	values := []interface{}{
		0.0,
		10.0,
		30.0,
		70.0,
		100.0,
	}
	rangeMap, err := util.NewRangeMap(keys, values)
	if err != nil {
		return nil, senecaerror.NewDevError(fmt.Errorf("NewRangeMap() returns err: %w", err))
	}

	return &accelerationV0{
		tag:      "00001",
		rangeMap: rangeMap,
	}, nil
}

func (acc *accelerationV0) GenerateEvents(inputs map[string][]interface{}) ([]*st.EventInternal, error) {
	motionsObjs := inputs[dataprocessor.RawMotionTypeString]
	if len(motionsObjs) == 0 {
		return nil, nil
	}

	events := []*st.EventInternal{}

	for _, motionObj := range motionsObjs {
		rawMotion, ok := motionObj.(*st.RawMotion)
		if !ok {
			return nil, fmt.Errorf("found a %T in map entry for %s", motionObj, dataprocessor.RawMotionTypeString)
		}

		valObj, ok := acc.rangeMap.Get(int64(rawMotion.Motion.AccelerationMphS))
		if !ok {
			return nil, fmt.Errorf("accelerationMPHS of %f for raw motion is impossible", rawMotion.Motion.AccelerationMphS)
		}
		severity, ok := valObj.(float64)
		if !ok {
			return nil, senecaerror.NewDevError(fmt.Errorf("trying to get float64 from rangeMap even though %T was inserted", valObj))
		}
		if severity == 0.0 {
			continue
		}

		events = append(events,
			&st.EventInternal{
				UserId:      rawMotion.UserId,
				EventType:   st.EventType_FAST_DECELERATION,
				Value:       rawMotion.Motion.AccelerationMphS,
				Severity:    severity,
				TimestampMs: rawMotion.TimestampMs,
				Source: &st.Source{
					SourceId:   rawMotion.Id,
					SourceType: st.Source_RAW_MOTION,
				},
				AlgoTag: acc.tag,
			})

	}
	return events, nil
}

func (acc *accelerationV0) GenerateDrivingConditions(inputs map[string][]interface{}) ([]*st.DrivingConditionInternal, error) {
	return nil, nil
}

func (acc *accelerationV0) Tag() string {
	return acc.tag
}
