package eventdao

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
	"seneca/internal/dao"
	"seneca/internal/util"
)

const (
	eventIDFieldName = "EventId"
	tripIDFieldName  = "TripId"
)

type SQLEventDao struct {
	sql     dao.SQLInterface
	tripDAO dao.TripDAO
}

func (edao *SQLEventDao) CreateEvent(ctx context.Context, event *st.EventInternal) (*st.EventInternal, error) {
	// Check if there's an existing trip.
	tripIDs, err := edao.tripDAO.ListUserTripIDsByTime(event.UserId, util.MillisecondsToTime(event.TimestampMs), util.MillisecondsToTime(event.TimestampMs))
	if err != nil {
		return nil, fmt.Errorf("error checking for existing trip: %w", err)
	}

	if len(tripIDs) > 1 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("more than one tripID at timestamp %q: %v", util.MillisecondsToTime(event.TimestampMs), tripIDs))
	}

	if len(tripIDs) == 0 {
		newTripNoID := &st.TripInternal{
			UserId:      event.UserId,
			StartTimeMs: event.TimestampMs,
			EndTimeMs:   event.TimestampMs,
		}

		newTrip, err := edao.tripDAO.CreateUniqueTrip(context.TODO(), newTripNoID)
		if err != nil {
			return nil, fmt.Errorf("error creating new trip for event: %w", err)
		}
		event.TripId = newTrip.Id
	} else {
		event.TripId = tripIDs[0]
	}

	eventID, err := edao.sql.Create(constants.EventTable, event)
	if err != nil {
		return nil, fmt.Errorf("error creating event %v - err: %w", event, err)
	}

	event.Id = eventID
	if err := edao.PutEventByID(context.TODO(), event.Id, event); err != nil {
		return nil, fmt.Errorf("error updating eventID %q: %w", event.Id, err)
	}

	return event, nil
}

func (edao *SQLEventDao) PutEventByID(ctx context.Context, eventID string, event *st.EventInternal) error {
	return edao.sql.Insert(constants.EventTable, event.Id, event)
}

func (edao *SQLEventDao) GetEventByID(userID, tripID, eventID string) (*st.EventInternal, error) {
	eventObj, err := edao.sql.GetByID(constants.EventTable, eventID)
	if err != nil {
		return nil, fmt.Errorf("error getting event by ID: %w", err)
	}

	event, ok := eventObj.(*st.EventInternal)
	if !ok {
		return nil, fmt.Errorf("want type EventInternal, got %T", eventObj)
	}

	if event.UserId != userID || event.TripId != tripID {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("mismatch between user or trip IDs (want, got) tripIDs(%s, %s) userIDs(%s, %s)", tripID, event.TripId, userID, event.UserId))
	}

	return event, nil
}

func (edao *SQLEventDao) ListTripEventIDs(tripID string) ([]string, error) {
	return edao.sql.ListIDs(constants.EventTable, []*cloud.QueryParam{{FieldName: tripIDFieldName, Operand: "=", Value: tripID}})
}

func (edao *SQLEventDao) DeleteEventByID(ctx context.Context, eventID string) error {
	return edao.sql.DeleteByID(constants.EventTable, eventID)
}
