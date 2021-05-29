package dataprocessor

import (
	"fmt"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/util"
)

type decelerationV0 struct {
	tag      string
	rangeMap util.RangeMap
}

func newDecelerationV0() *decelerationV0 {
	return &decelerationV0{
		tag: "00002",
		rangeMap: util.RangeMap{
			Keys: []util.Range{
				{L: -256, U: -128},
				{L: -128, U: -64},
				{L: -64, U: -32},
				{L: -32, U: -16},
				{L: -16, U: -8},
				{L: -8, U: 9999999},
			},
			Values: []interface{}{
				100.0,
				80.0,
				50.0,
				20.0,
				10.0,
				0.0,
			},
		},
	}
}

func (dec *decelerationV0) GenerateEvent(inputs []interface{}) (*st.EventInternal, error) {
	if len(inputs) != 1 {
		return nil, fmt.Errorf("input to decelerationV0 must be a single RawMotion")
	}

	rawMotion, ok := inputs[0].(*st.RawMotion)
	if !ok {
		return nil, fmt.Errorf("input to decelerationV0 must be a single RawMotion")
	}

	valObj, ok := dec.rangeMap.Get(rawMotion.Motion.AccelerationMphS)
	if !ok {
		return nil, fmt.Errorf("accelerationMPHS of %f for raw motion is impossible", rawMotion.Motion.AccelerationMphS)
	}
	severity, ok := valObj.(float64)
	if !ok {
		return nil, senecaerror.NewDevError(fmt.Errorf("trying to get float64 from rangeMap even though %T was inserted", valObj))
	}
	if severity == 0.0 {
		return nil, nil
	}

	return &st.EventInternal{
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
	}, nil
}

func (dec *decelerationV0) GenerateDrivingCondition(inputs []interface{}) (*st.DrivingConditionInternal, error) {
	return nil, nil
}

func (dec *decelerationV0) Tag() string {
	return dec.tag
}

// First very naive implementation.
type accelerationV0 struct {
	tag      string
	rangeMap util.RangeMap
}

func newAccelerationV0() *accelerationV0 {
	return &accelerationV0{
		tag: "00001",
		rangeMap: util.RangeMap{
			Keys: []util.Range{
				{L: -9999999, U: 7.5},
				{L: 7.5, U: 10},
				{L: 10, U: 15},
				{L: 15, U: 20},
				{L: 20, U: 25},
			},
			Values: []interface{}{
				0.0,
				10.0,
				30.0,
				70.0,
				100.0,
			},
		},
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

	valObj, ok := acc.rangeMap.Get(rawMotion.Motion.AccelerationMphS)
	if !ok {
		return nil, fmt.Errorf("accelerationMPHS of %f for raw motion is impossible", rawMotion.Motion.AccelerationMphS)
	}
	severity, ok := valObj.(float64)
	if !ok {
		return nil, senecaerror.NewDevError(fmt.Errorf("trying to get float64 from rangeMap even though %T was inserted", valObj))
	}
	if severity == 0.0 {
		return nil, nil
	}

	return &st.EventInternal{
		UserId:      rawMotion.UserId,
		EventType:   st.EventType_FAST_ACCELERATION,
		Value:       rawMotion.Motion.AccelerationMphS,
		Severity:    severity,
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
