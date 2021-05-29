package eventdao

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
	"seneca/internal/util"
)

type SQLEventDAO struct {
	sql     database.SQLInterface
	tripDAO dao.TripDAO
	logger  logging.LoggingInterface
}

func NewSQLEventDAO(sql database.SQLInterface, tripDAO dao.TripDAO, logger logging.LoggingInterface) *SQLEventDAO {
	return &SQLEventDAO{
		sql:     sql,
		tripDAO: tripDAO,
		logger:  logger,
	}
}

func (edao *SQLEventDAO) CreateEvent(ctx context.Context, event *st.EventInternal) (*st.EventInternal, error) {
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
	if err := edao.PutEventByID(context.TODO(), event.UserId, event.TripId, event.Id, event); err != nil {
		return nil, fmt.Errorf("error updating eventID %q: %w", event.Id, err)
	}

	return event, nil
}

func (edao *SQLEventDAO) PutEventByID(ctx context.Context, userID, tripID, eventID string, event *st.EventInternal) error {
	err := edao.sql.Insert(constants.EventTable, event.Id, event)
	if err != nil {
		edao.logger.Log(fmt.Sprintf("Put eventInternal for user %s for trip %s at %v", userID, tripID, util.MillisecondsToTime(event.TimestampMs)))
	}
	return err
}

func (edao *SQLEventDAO) GetEventByID(userID, tripID, eventID string) (*st.EventInternal, error) {
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

func (edao *SQLEventDAO) ListTripEventIDs(userID, tripID string) ([]string, error) {
	return edao.sql.ListIDs(constants.EventTable, []*database.QueryParam{{FieldName: constants.TripIDFieldName, Operand: "=", Value: tripID}})
}

func (edao *SQLEventDAO) DeleteEventByID(ctx context.Context, userID, tripID, eventID string) error {
	return edao.sql.DeleteByID(constants.EventTable, eventID)
}
