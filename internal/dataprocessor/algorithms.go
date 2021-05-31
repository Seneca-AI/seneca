package dataprocessor

import (
	"fmt"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/weather"
	"seneca/internal/client/weather/client"
	"seneca/internal/util"
	"sort"
	"time"
)

const (
	weatherRadius = 3
)

type weatherV0 struct {
	tag           string
	weatherClient *client.WeatherClient
}

func newWeatherV0(weatherService weather.WeatherServiceInterface) *weatherV0 {
	return &weatherV0{
		tag:           "00003",
		weatherClient: client.New(weatherService, weatherRadius),
	}
}

func (wthr *weatherV0) GenerateEvents(inputs []interface{}) ([]*st.EventInternal, error) {
	return nil, nil
}

type drivingConditionAndSeverity struct {
	dc       st.ConditionType
	severity float64
}

// TODO(lucaloncar): also make this let bootleg...the TimstampedWeatherCondition object
// needs to return more data and then we'll make better inferences here.
var weatherStackCodeToDrivingConidtionAndSeverity = map[int]drivingConditionAndSeverity{
	// 	Blizzard
	230: {dc: st.ConditionType_SNOW, severity: 100},
	// 	Heavy snow
	338: {dc: st.ConditionType_SNOW, severity: 80},
	// 	Blowing snow
	227: {dc: st.ConditionType_SNOW, severity: 70},
	// 	Patchy heavy snow
	335: {dc: st.ConditionType_SNOW, severity: 60},
	// 	Moderate or heavy snow in area with thunder
	395: {dc: st.ConditionType_SNOW, severity: 55},
	// 	Moderate snow
	332: {dc: st.ConditionType_SNOW, severity: 50},
	// 	Moderate or heavy snow showers
	371: {dc: st.ConditionType_SNOW, severity: 45},
	// 	Patchy moderate snow
	329: {dc: st.ConditionType_SNOW, severity: 40},
	// 	Light snow
	326: {dc: st.ConditionType_SNOW, severity: 30},
	// 	Light snow showers
	368: {dc: st.ConditionType_SNOW, severity: 20},
	// 	Patchy light snow
	323: {dc: st.ConditionType_SNOW, severity: 10},
	// 	Patchy light snow in area with thunder
	392: {dc: st.ConditionType_SNOW, severity: 10},
	// 	Patchy snow nearby
	179: {dc: st.ConditionType_SNOW, severity: 0},

	// 	Torrential rain shower
	359: {dc: st.ConditionType_RAIN, severity: 100},
	// 	Heavy rain
	308: {dc: st.ConditionType_RAIN, severity: 80},
	// 	Heavy rain at times
	305: {dc: st.ConditionType_RAIN, severity: 70},
	// 	Moderate or heavy rain in area with thunder
	389: {dc: st.ConditionType_RAIN, severity: 60},
	// 	Moderate or heavy rain shower
	356: {dc: st.ConditionType_RAIN, severity: 60},
	// 	Moderate rain
	302: {dc: st.ConditionType_RAIN, severity: 50},
	// 	Moderate rain at times
	299: {dc: st.ConditionType_RAIN, severity: 40},
	// 	Light rain
	296: {dc: st.ConditionType_RAIN, severity: 30},
	// 	Patchy light rain in area with thunder
	386: {dc: st.ConditionType_RAIN, severity: 20},
	// 	Patchy light rain
	293: {dc: st.ConditionType_RAIN, severity: 20},
	// 	Light rain shower
	353: {dc: st.ConditionType_RAIN, severity: 15},
	// 	Light drizzle
	266: {dc: st.ConditionType_RAIN, severity: 10},
	// 	Patchy light drizzle
	263: {dc: st.ConditionType_RAIN, severity: 5},
	// 	Patchy rain nearby
	176: {dc: st.ConditionType_RAIN, severity: 0},

	// 	Ice pellets
	350: {dc: st.ConditionType_HAIL, severity: 100},
	//	Moderate or heavy showers of ice pellets
	377: {dc: st.ConditionType_HAIL, severity: 60},
	//	Light showers of ice pellets
	374: {dc: st.ConditionType_HAIL, severity: 30},

	// 	Moderate or heavy sleet
	320: {dc: st.ConditionType_SLEET, severity: 75},
	// 	Moderate or heavy sleet showers
	365: {dc: st.ConditionType_SLEET, severity: 60},
	// 	Light sleet
	317: {dc: st.ConditionType_SLEET, severity: 45},
	// 	Light sleet showers
	362: {dc: st.ConditionType_SLEET, severity: 30},
	// 	Patchy sleet nearby
	182: {dc: st.ConditionType_SLEET, severity: 0},

	// 	Fog
	248: {dc: st.ConditionType_FOG, severity: 70},
	// 	Mist
	143: {dc: st.ConditionType_FOG, severity: 30},

	// 	Moderate or Heavy freezing rain
	314: {dc: st.ConditionType_FREEZING_RAIN, severity: 100},
	// 	Light freezing rain
	311: {dc: st.ConditionType_FREEZING_RAIN, severity: 80},
	// 	Heavy freezing drizzle
	284: {dc: st.ConditionType_FREEZING_RAIN, severity: 60},
	// 	Freezing fog
	260: {dc: st.ConditionType_FREEZING_RAIN, severity: 40},
	// 	Freezing drizzle
	281: {dc: st.ConditionType_FREEZING_RAIN, severity: 20},
	// 	Patchy freezing drizzle nearby
	185: {dc: st.ConditionType_FREEZING_RAIN, severity: 0},

	// 	Thundery outbreaks in nearby
	200: {dc: st.ConditionType_NONE_CONDITION_TYPE, severity: 0},
	// 	Overcast
	122: {dc: st.ConditionType_NONE_CONDITION_TYPE, severity: 0},
	// 	Cloudy
	119: {dc: st.ConditionType_NONE_CONDITION_TYPE, severity: 0},
	// 	Partly Cloudy
	116: {dc: st.ConditionType_NONE_CONDITION_TYPE, severity: 0},
	// 	Clear/Sunny
	113: {dc: st.ConditionType_NONE_CONDITION_TYPE, severity: 0},
}

type timestampAndSource struct {
	timestamp int64
	source    *st.Source
}

func (wthr *weatherV0) GenerateDrivingConditions(inputs []interface{}) ([]*st.DrivingConditionInternal, error) {
	// Keyed by conditionType and severity.
	conditionTypesMap := map[drivingConditionAndSeverity][]timestampAndSource{}

	foundLocation := true
	userID := ""
	for _, inputObj := range inputs {
		location, ok := inputObj.(*st.RawLocation)
		if !ok {
			if foundLocation {
				return nil, fmt.Errorf("found a %T in the same list as RawLocations", inputObj)
			}
			return nil, nil
		}
		foundLocation = true
		userID = location.UserId

		for _, ranTag := range location.AlgoTag {
			if ranTag == wthr.Tag() {
				continue
			}
		}

		twc, err := wthr.weatherClient.GetHistoricalWeather(util.MillisecondsToTime(location.TimestampMs), location.Location.Lat, location.Location.Long)
		if err != nil {
			return nil, fmt.Errorf("GetHistoricalWeather() returns err: %w", err)
		}

		dcAndSeverity, ok := weatherStackCodeToDrivingConidtionAndSeverity[twc.WeatherCode]
		if !ok {
			return nil, fmt.Errorf("unknown weather code %d", twc.WeatherCode)
		}

		if _, ok := conditionTypesMap[dcAndSeverity]; !ok {
			conditionTypesMap[dcAndSeverity] = []timestampAndSource{}
		}
		conditionTypesMap[dcAndSeverity] = append(
			conditionTypesMap[dcAndSeverity],
			timestampAndSource{
				source: location.Source, timestamp: location.TimestampMs,
			},
		)
	}

	drivingConditions := []*st.DrivingConditionInternal{}
	for k, timestamps := range conditionTypesMap {
		sort.Slice(timestamps, func(i, j int) bool { return timestamps[i].timestamp < timestamps[j].timestamp })

		// Merge driving conditions less than an hour apart.
		var nextDrivingCondition *st.DrivingConditionInternal
		lastTimeStamp := timestamps[0].timestamp
		for _, ts := range timestamps {
			if nextDrivingCondition == nil {
				nextDrivingCondition = &st.DrivingConditionInternal{
					UserId:        userID,
					ConditionType: k.dc,
					Severity:      k.severity,
					StartTimeMs:   lastTimeStamp,
					Source:        ts.source,
					AlgoTag:       wthr.Tag(),
				}
				lastTimeStamp = ts.timestamp
			}
			if ts.timestamp-lastTimeStamp > time.Hour.Milliseconds() {
				nextDrivingCondition.EndTimeMs = lastTimeStamp
				drivingConditions = append(drivingConditions, nextDrivingCondition)
				nextDrivingCondition = nil
			}
			lastTimeStamp = ts.timestamp
		}
		if nextDrivingCondition != nil {
			nextDrivingCondition.EndTimeMs = timestamps[len(timestamps)-1].timestamp
			drivingConditions = append(drivingConditions, nextDrivingCondition)
			nextDrivingCondition = nil
		}
	}

	return drivingConditions, nil
}

func (wthr *weatherV0) Tag() string {
	return wthr.tag
}

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

func (dec *decelerationV0) GenerateEvents(inputs []interface{}) ([]*st.EventInternal, error) {
	if len(inputs) != 1 {
		return nil, nil
	}

	rawMotion, ok := inputs[0].(*st.RawMotion)
	if !ok {
		return nil, nil
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
		return nil, nil
	}

	return []*st.EventInternal{
		{
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
		},
	}, nil
}

func (dec *decelerationV0) GenerateDrivingConditions(inputs []interface{}) ([]*st.DrivingConditionInternal, error) {
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

func (acc *accelerationV0) GenerateEvents(inputs []interface{}) ([]*st.EventInternal, error) {
	if len(inputs) != 1 {
		return nil, nil
	}

	rawMotion, ok := inputs[0].(*st.RawMotion)
	if !ok {
		return nil, nil
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
		return nil, nil
	}

	return []*st.EventInternal{
		{
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
		},
	}, nil
}

func (acc *accelerationV0) GenerateDrivingConditions(inputs []interface{}) ([]*st.DrivingConditionInternal, error) {
	return nil, nil
}

func (acc *accelerationV0) Tag() string {
	return acc.tag
}
