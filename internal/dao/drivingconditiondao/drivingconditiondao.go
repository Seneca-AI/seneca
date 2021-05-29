package drivingconditiondao

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/dao"
	"seneca/internal/util"
	"sort"
	"time"
)

type SQLDrivingConditionDAO struct {
	sql      database.SQLInterface
	tripDAO  dao.TripDAO
	eventDAO dao.EventDAO
}

func NewSQLDrivingConditionDAO(sql database.SQLInterface, tripDAO dao.TripDAO, eventDAO dao.EventDAO) *SQLDrivingConditionDAO {
	return &SQLDrivingConditionDAO{
		sql:      sql,
		tripDAO:  tripDAO,
		eventDAO: eventDAO,
	}
}

func (ddao *SQLDrivingConditionDAO) CreateDrivingCondition(ctx context.Context, drivingCondition *st.DrivingConditionInternal) (*st.DrivingConditionInternal, error) {
	// Check if there're a existing trips.
	tripIDs, err := ddao.tripDAO.ListUserTripIDsByTime(drivingCondition.UserId, util.MillisecondsToTime(drivingCondition.StartTimeMs), util.MillisecondsToTime(drivingCondition.EndTimeMs))
	if err != nil {
		return nil, fmt.Errorf("error checking for existing trip: %w", err)
	}

	tripID, err := ddao.mergeTrips(drivingCondition, tripIDs)
	if err != nil {
		return nil, fmt.Errorf("error trying to merge parent trips: %w", err)
	}

	drivingCondition.TripId = tripID

	drivingConditionID, err := ddao.sql.Create(constants.DrivingConditionTable, drivingCondition)
	if err != nil {
		return nil, fmt.Errorf("error creating drivingCondition %v - err: %w", drivingCondition, err)
	}

	drivingCondition.Id = drivingConditionID
	if err := ddao.sql.Insert(constants.DrivingConditionTable, drivingCondition.Id, drivingCondition); err != nil {
		return nil, fmt.Errorf("error updating drivingConditionID %q: %w", drivingCondition.Id, err)
	}

	return drivingCondition, nil
}

func (ddao *SQLDrivingConditionDAO) GetDrivingConditionByID(userID, tripID, drivingConditionID string) (*st.DrivingConditionInternal, error) {
	drivingConditionObj, err := ddao.sql.GetByID(constants.DrivingConditionTable, drivingConditionID)
	if err != nil {
		return nil, fmt.Errorf("error getting event by ID: %w", err)
	}
	if drivingConditionObj == nil {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("drivingCondition with ID %q not found in the store", drivingConditionID))
	}

	drivingCondition, ok := drivingConditionObj.(*st.DrivingConditionInternal)
	if !ok {
		return nil, fmt.Errorf("want type DrivingConditionInternal, got %T", drivingConditionObj)
	}

	if drivingCondition.UserId != userID || drivingCondition.TripId != tripID {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("mismatch between user or trip IDs (want, got) tripIDs(%s, %s) userIDs(%s, %s)", tripID, drivingCondition.TripId, userID, drivingCondition.UserId))
	}

	return drivingCondition, nil
}

func (ddao *SQLDrivingConditionDAO) ListTripDrivingConditionIDs(userID, tripID string) ([]string, error) {
	return ddao.sql.ListIDs(constants.DrivingConditionTable, []*database.QueryParam{{FieldName: constants.TripIDFieldName, Operand: "=", Value: tripID}})
}
func (ddao *SQLDrivingConditionDAO) DeleteDrivingConditionByID(ctx context.Context, userID, tripID, drivingConditionID string) error {
	return ddao.sql.DeleteByID(constants.DrivingConditionTable, drivingConditionID)
}

func (ddao SQLDrivingConditionDAO) mergeTrips(drivingCondition *st.DrivingConditionInternal, tripIDs []string) (string, error) {
	// Merge and sort.
	trips := []*st.TripInternal{
		{
			UserId:      drivingCondition.UserId,
			StartTimeMs: drivingCondition.StartTimeMs,
			EndTimeMs:   drivingCondition.EndTimeMs,
		},
	}
	if len(tripIDs) > 0 {
		for _, tid := range tripIDs {
			trip, err := ddao.tripDAO.GetTripByID(drivingCondition.UserId, tid)
			if err != nil {
				return "", fmt.Errorf("error getting trip by ID %q - err: %w", tid, err)
			}
			trips = append(trips, trip)
		}
	}
	sort.Slice(trips, func(i, j int) bool { return trips[i].StartTimeMs < trips[j].StartTimeMs })

	var winnerTrip *st.TripInternal

	// Existing one keeps ID.
	for _, tp := range trips {
		if tp.Id != "" {
			winnerTrip = tp
			break
		}
	}

	if winnerTrip == nil {
		newTrip, err := ddao.tripDAO.CreateUniqueTrip(context.TODO(), trips[0])
		if err != nil {
			return "", fmt.Errorf("erroring creating new trip: %w", err)
		}
		return newTrip.Id, nil
	}

	newStartTime := util.TimeToMilliseconds(time.Now())
	newEndTime := int64(0)
	for _, tp := range trips {
		if tp.StartTimeMs < newStartTime {
			newStartTime = tp.StartTimeMs
		}
		if tp.EndTimeMs > newEndTime {
			newEndTime = tp.EndTimeMs
		}

		if tp.Id == winnerTrip.Id || tp.Id == "" {
			continue
		}

		// Re-assign children.
		eventIDs, err := ddao.eventDAO.ListTripEventIDs(drivingCondition.UserId, tp.Id)
		if err != nil {
			return "", fmt.Errorf("error listing eventIDs for tripID %q - err: %w", tp.Id, err)
		}
		for _, eid := range eventIDs {
			event, err := ddao.eventDAO.GetEventByID(drivingCondition.UserId, tp.Id, eid)
			if err != nil {
				return "", fmt.Errorf("error getting event by ID %q - err: %w", eid, err)
			}
			event.TripId = winnerTrip.Id
			if err := ddao.eventDAO.PutEventByID(context.TODO(), event.UserId, event.TripId, event.Id, event); err != nil {
				return "", fmt.Errorf("error putting event %v - err: %w", event, err)
			}
		}

		existingDrivingConditionIDs, err := ddao.ListTripDrivingConditionIDs(drivingCondition.UserId, tp.Id)
		if err != nil {
			return "", fmt.Errorf("error listing existingDrivingConditionIDs for tripID %q - err: %w", tp.Id, err)
		}
		for _, edci := range existingDrivingConditionIDs {
			condition, err := ddao.GetDrivingConditionByID(drivingCondition.UserId, tp.Id, edci)
			if err != nil {
				return "", fmt.Errorf("error getting existingDrivingCondition with userID %q, tripID %q, ID %q  - err: %w", tp.UserId, tp.Id, edci, err)
			}
			condition.TripId = winnerTrip.Id
			if err := ddao.sql.Insert(constants.DrivingConditionTable, condition.Id, condition); err != nil {
				return "", fmt.Errorf("error putting existingDrivingCondition %v - err: %w", condition, err)
			}
		}
	}

	winnerTrip = &st.TripInternal{
		UserId:      drivingCondition.UserId,
		Id:          winnerTrip.Id,
		StartTimeMs: newStartTime,
		EndTimeMs:   newEndTime,
	}

	if err := ddao.tripDAO.PutTripByID(context.TODO(), winnerTrip.Id, winnerTrip); err != nil {
		return "", fmt.Errorf("error updating new trip with ID %q: %w", winnerTrip.Id, err)
	}

	// Delete old trips.
	for _, tp := range trips {
		if tp.Id != winnerTrip.Id && tp.Id != "" {
			if err := ddao.tripDAO.DeleteTripByID(context.TODO(), tp.Id); err != nil {
				return "", fmt.Errorf("error deleting trip %q by ID: %w", tp.Id, err)
			}
		}
	}

	return winnerTrip.Id, nil
}
