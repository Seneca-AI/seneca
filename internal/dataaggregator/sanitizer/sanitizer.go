package sanitizer

import (
	"fmt"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/dao"
	"sort"
)

type Sanitizer struct {
	rawMotionDAO        dao.RawMotionDAO
	rawLocationDAO      dao.RawLocationDAO
	rawVideoDAO         dao.RawVideoDAO
	eventDAO            dao.EventDAO
	drivingConditionDAO dao.DrivingConditionDAO
	// Cache the URL of sources.  The source video URL will always be the same if it exists.
	// Keys will be in the form SOURCE_TYPE/SOURCE_ID , eg 'RAW_MOTION/123'.
	videoURLCache map[string]string
}

func New(rawMotionDAO dao.RawMotionDAO, rawLocationDAO dao.RawLocationDAO, rawVideoDAO dao.RawVideoDAO, eventDAO dao.EventDAO, drivingConditionDAO dao.DrivingConditionDAO) *Sanitizer {
	return &Sanitizer{
		rawMotionDAO:        rawMotionDAO,
		rawLocationDAO:      rawLocationDAO,
		rawVideoDAO:         rawVideoDAO,
		eventDAO:            eventDAO,
		drivingConditionDAO: drivingConditionDAO,
		videoURLCache:       map[string]string{},
	}
}

func (san *Sanitizer) TripInternalToTripExternal(tripInternal *st.TripInternal) (*st.Trip, error) {
	externalEvents := []*st.Event{}
	internalEventIDs, err := san.eventDAO.ListTripEventIDs(tripInternal.UserId, tripInternal.Id)
	if err != nil {
		return nil, fmt.Errorf("error listing event IDs: %w", err)
	}

	for _, ieid := range internalEventIDs {
		internalEvent, err := san.eventDAO.GetEventByID(tripInternal.UserId, tripInternal.Id, ieid)
		if err != nil {
			return nil, fmt.Errorf("error getting event by ID %q - err: %w", ieid, err)
		}
		externalEvent, err := san.eventInternalToEventExternal(internalEvent)
		if err != nil {
			return nil, fmt.Errorf("error converting EventInternal %v to Event %v - err: %w", internalEvent, externalEvent, err)
		}
		externalEvents = append(externalEvents, externalEvent)
	}
	sort.Slice(externalEvents, func(i, j int) bool { return externalEvents[i].TimestampMs < externalEvents[j].TimestampMs })

	internalDrivingConditionIDs, err := san.drivingConditionDAO.ListTripDrivingConditionIDs(tripInternal.UserId, tripInternal.Id)
	if err != nil {
		return nil, fmt.Errorf("error listing drivingCondition IDs: %w", err)
	}
	internalDrivingConditions := []*st.DrivingConditionInternal{}
	for _, idcid := range internalDrivingConditionIDs {
		internalDrivingCondition, err := san.drivingConditionDAO.GetDrivingConditionByID(tripInternal.UserId, tripInternal.Id, idcid)
		if err != nil {
			return nil, fmt.Errorf("error getting drivingCondition by ID %q - err: %w", internalDrivingCondition, err)
		}
		internalDrivingConditions = append(internalDrivingConditions, internalDrivingCondition)
	}
	externalDrivingConditions, err := san.drivingConditionsInternalToDrivingConditionsExternal(internalDrivingConditions)
	if err != nil {
		return nil, fmt.Errorf("error converting from internal to external drivingConditions: %w", err)
	}

	return &st.Trip{
		StartTimeMs:      tripInternal.StartTimeMs,
		EndTimeMs:        tripInternal.EndTimeMs,
		Event:            externalEvents,
		DrivingCondition: externalDrivingConditions,
	}, nil
}

// Walk the chain of sources until the RawVideo is found.
func (san *Sanitizer) findVideoLink(source *st.Source) (string, error) {
	if source == nil {
		return "", fmt.Errorf("source is nil")
	}

	// First check the cache.
	if videoURL, ok := san.videoURLCache[fmt.Sprintf("%s/%s", source.SourceType, source.SourceId)]; ok {
		return videoURL, nil
	}

	url, sourceKeys, err := func(maxAttempts int) (string, []string, error) {
		videoURL := ""
		keys := []string{}
		attempts := 0
		// Hop through sources until we find a RawVideoURL.
		for {
			if source == nil || attempts > maxAttempts {
				return "", nil, fmt.Errorf("could not find video URL for source")
			}
			attempts++

			keys = append(keys, fmt.Sprintf("%s/%s", source.SourceType, source.SourceId))
			switch source.SourceType {
			case st.Source_RAW_VIDEO:
				rawVideo, err := san.rawVideoDAO.GetRawVideoByID(source.SourceId)
				if err != nil {
					return "", nil, fmt.Errorf("GetRawVideoByID(%s) returns err: %w", source.SourceId, err)
				}
				videoURL = rawVideo.CloudStorageFileName
				return videoURL, keys, nil
			case st.Source_RAW_MOTION:
				rawMotion, err := san.rawMotionDAO.GetRawMotionByID(source.SourceId)
				if err != nil {
					return "", nil, fmt.Errorf("GetRawMotionByID(%s) returns err: %w", source.SourceId, err)
				}
				source = rawMotion.Source
			case st.Source_RAW_LOCATION:
				rawLocation, err := san.rawLocationDAO.GetRawLocationByID(source.SourceId)
				if err != nil {
					return "", nil, fmt.Errorf("GetRawLocationByID(%s) returns err: %w", source.SourceId, err)
				}
				source = rawLocation.Source
			default:
				return "", nil, fmt.Errorf("unsupported source type %q", source.SourceType)
			}
		}
	}(10)
	if err != nil {
		return "", err
	}

	for _, k := range sourceKeys {
		san.videoURLCache[k] = url
	}

	return url, nil
}

func (san *Sanitizer) eventInternalToEventExternal(eventInternal *st.EventInternal) (*st.Event, error) {
	eventExternal := &st.Event{}

	if eventInternal.EventType == st.EventType_UNKNOWN_EVENT_TYPE {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("event with ID %q for user %q has EventType as UNKNOWN", eventInternal.Id, eventInternal.UserId))
	}
	eventExternal.EventType = eventInternal.EventType

	eventExternal.Value = eventInternal.Value
	eventExternal.Severity = eventInternal.Severity

	if eventInternal.TimestampMs == 0 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("event with ID %q for user %q has timestamp as 0", eventInternal.Id, eventInternal.UserId))
	}

	eventExternal.TimestampMs = eventInternal.TimestampMs

	sourceVideoLink, err := san.findVideoLink(eventInternal.Source)
	if err != nil {
		return nil, fmt.Errorf("error finding source video link: %w", err)
	}
	eventExternal.ExternalSource = &st.ExternalSource{
		SourceType: st.ExternalSource_DASHCAM_VIDEO,
		VideoUrl:   sourceVideoLink,
	}

	return eventExternal, nil
}

type conditionAndSource struct {
	condition      st.ConditionType
	sourceVideoURL string
}

func (san *Sanitizer) drivingConditionsInternalToDrivingConditionsExternal(drivingConditionsInternal []*st.DrivingConditionInternal) ([]*st.DrivingCondition, error) {
	drivingConditionsExternal := []*st.DrivingCondition{}

	points := map[int64]map[conditionAndSource]float64{}

	for _, dci := range drivingConditionsInternal {
		points[dci.StartTimeMs] = map[conditionAndSource]float64{}
		points[dci.EndTimeMs] = map[conditionAndSource]float64{}
	}

	timestamps := []int64{}

	for ts := range points {
		timestamps = append(timestamps, ts)
	}

	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	for _, dci := range drivingConditionsInternal {
		sourceVideoURL, err := san.findVideoLink(dci.Source)
		if err != nil {
			return nil, fmt.Errorf("error finding source video link: %w", err)
		}
		condAndSrc := conditionAndSource{
			condition:      dci.ConditionType,
			sourceVideoURL: sourceVideoURL,
		}

		for _, ts := range timestamps {
			if ts >= dci.StartTimeMs && ts < dci.EndTimeMs {
				currentMax, ok := points[ts][condAndSrc]
				if !ok || dci.Severity > currentMax {
					points[ts][condAndSrc] = dci.Severity
				}
			}
		}
	}

	// Key is conditionType/severity. lastMap/currentMap are used to 'merge'
	// drivingConditions if they were the same consecutively.
	lastMap := map[string]bool{}
	for i := 0; i < len(timestamps)-1; i++ {
		dcExternal := &st.DrivingCondition{
			StartTimeMs: timestamps[i],
			EndTimeMs:   timestamps[i+1] - 1,
		}

		lastMapSize := len(lastMap)
		currentMap := map[string]bool{}
		for condAndSrc, severity := range points[timestamps[i]] {
			if condAndSrc.condition != st.ConditionType_NONE_CONDITION_TYPE {
				conditionAndSeverity := fmt.Sprintf("%s/%f", condAndSrc.condition, severity)
				currentMap[conditionAndSeverity] = true
				delete(lastMap, conditionAndSeverity)

				dcExternal.ConditionType = append(dcExternal.ConditionType, condAndSrc.condition)
				dcExternal.Severity = append(dcExternal.Severity, severity)
			}
			dcExternal.ExternalSource = append(dcExternal.ExternalSource, &st.ExternalSource{
				SourceType: st.ExternalSource_DASHCAM_VIDEO,
				VideoUrl:   condAndSrc.sourceVideoURL,
			})
		}

		if i != 0 && len(lastMap) == 0 && lastMapSize == len(currentMap) {
			drivingConditionsExternal[len(drivingConditionsExternal)-1].EndTimeMs = timestamps[i+1] - 1
		} else {
			drivingConditionsExternal = append(drivingConditionsExternal, dcExternal)
		}

		for lm := range lastMap {
			delete(lastMap, lm)
		}
		for cm := range currentMap {
			lastMap[cm] = true
		}
	}

	sort.Slice(drivingConditionsExternal, func(i, j int) bool {
		return drivingConditionsExternal[i].StartTimeMs < drivingConditionsExternal[j].StartTimeMs
	})

	return drivingConditionsExternal, nil
}
