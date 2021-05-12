package apiserver

import (
	"fmt"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/dao"
	"sort"
	"time"
)

type Sanitizer struct {
	tripDAO             dao.TripDAO
	eventDAO            dao.EventDAO
	drivingConditionDAO dao.DrivingConditionDAO
}

func NewSanitizer(tripDAO dao.TripDAO, eventDAO dao.EventDAO, drivingConditionDAO dao.DrivingConditionDAO) *Sanitizer {
	return &Sanitizer{
		tripDAO:             tripDAO,
		eventDAO:            eventDAO,
		drivingConditionDAO: drivingConditionDAO,
	}
}

func (san *Sanitizer) ListTrips(userID string, startTime time.Time, endTime time.Time) ([]*st.Trip, error) {

	trips := []*st.Trip{}

	tripIDs, err := san.tripDAO.ListUserTripIDsByTime(userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("ListUserTripIDsByTime(%s, %s, %s) returns err: %w", userID, startTime, endTime, err)
	}

	for _, tid := range tripIDs {
		tripInternal, err := san.tripDAO.GetTripByID(userID, tid)
		if err != nil {
			return nil, fmt.Errorf("GetTripByID(%s) returns err: %w", tid, err)
		}
		tripExternal, err := san.tripInternalToTripExternal(tripInternal)
		if err != nil {
			return nil, fmt.Errorf("error converting internal trip %v to external trip: %w", tripInternal, err)
		}
		trips = append(trips, tripExternal)
	}

	return trips, nil
}

func (san *Sanitizer) tripInternalToTripExternal(tripInternal *st.TripInternal) (*st.Trip, error) {
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
		externalEvent, err := eventInternalToEventExternal(internalEvent)
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
	externalDrivingConditions, err := drivingConditionsInternalToDrivingConditionsExternal(internalDrivingConditions)
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

func eventInternalToEventExternal(eventInternal *st.EventInternal) (*st.Event, error) {
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
	return eventExternal, nil
}

func drivingConditionsInternalToDrivingConditionsExternal(drivingConditionsInternal []*st.DrivingConditionInternal) ([]*st.DrivingCondition, error) {
	drivingConditionsExternal := []*st.DrivingCondition{}

	points := map[int64]map[st.ConditionType]float64{}

	for _, dci := range drivingConditionsInternal {
		points[dci.StartTimeMs] = map[st.ConditionType]float64{}
		points[dci.EndTimeMs] = map[st.ConditionType]float64{}
	}

	timestamps := []int64{}

	for ts := range points {
		timestamps = append(timestamps, ts)
	}

	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	for _, dci := range drivingConditionsInternal {
		for _, ts := range timestamps {
			if ts >= dci.StartTimeMs && ts < dci.EndTimeMs {
				currentMax, ok := points[ts][dci.ConditionType]
				if !ok || dci.Severity > currentMax {
					points[ts][dci.ConditionType] = dci.Severity
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
		for conditionType, severity := range points[timestamps[i]] {
			if conditionType != st.ConditionType_NONE_CONDITION_TYPE {

				conditionAndSeverity := fmt.Sprintf("%s/%f", conditionType, severity)
				currentMap[conditionAndSeverity] = true
				delete(lastMap, conditionAndSeverity)

				dcExternal.ConditionType = append(dcExternal.ConditionType, conditionType)
				dcExternal.Severity = append(dcExternal.Severity, severity)
			}
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

	// Merge time periods that have the same condition type and severities.

	return drivingConditionsExternal, nil
}
