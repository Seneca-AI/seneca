package tripdao

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/util"
	"time"
)

const (
	userIDFieldName    = "UserId"
	startTimeFieldName = "StartTimeMs"
	endTimeFieldName   = "EndTimeMs"
)

type SQLTripDAO struct {
	sql database.SQLInterface
}

func NewSQLTripDAO(sql database.SQLInterface) *SQLTripDAO {
	return &SQLTripDAO{
		sql: sql,
	}
}

func (tdao *SQLTripDAO) CreateUniqueTrip(ctx context.Context, trip *st.TripInternal) (*st.TripInternal, error) {
	// See if there are any existing overlapping trips.
	tripIDs, err := tdao.ListUserTripIDsByTime(trip.UserId, util.MillisecondsToTime(trip.StartTimeMs), util.MillisecondsToTime(trip.EndTimeMs))
	if err != nil {
		return nil, fmt.Errorf("error checking for existing trips: %w", err)
	}

	if len(tripIDs) > 0 {
		return nil, fmt.Errorf("overlapping trip already exists")
	}

	tripID, err := tdao.sql.Create(constants.TripTable, trip)
	if err != nil {
		return nil, fmt.Errorf("error creating trip %v - err: %w", trip, err)
	}

	trip.Id = tripID
	if err := tdao.PutTripByID(context.TODO(), trip.Id, trip); err != nil {
		return nil, fmt.Errorf("error patching trip's ID %q - err: %w", trip.Id, err)
	}

	fmt.Printf("Created trip %v\n", trip)
	return trip, nil
}

func (tdao *SQLTripDAO) GetTripByID(userID, tripID string) (*st.TripInternal, error) {
	tripObj, err := tdao.sql.GetByID(constants.TripTable, tripID)
	if err != nil {
		return nil, fmt.Errorf("error getting trip from strore: %w", err)
	}

	if tripObj == nil {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("tripInternal with ID %q not found in the store", tripID))
	}

	trip, ok := tripObj.(*st.TripInternal)
	if !ok {
		return nil, fmt.Errorf("expected type TripInternal, got %T", tripObj)
	}

	if trip.UserId != userID {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("mistmatch between UserID and tripID.UserID (UserID, tripID.UserID)(%s, %s)", userID, trip.UserId))
	}

	return trip, nil
}

func (tdao *SQLTripDAO) PutTripByID(ctx context.Context, tripID string, trip *st.TripInternal) error {
	return tdao.sql.Insert(constants.TripTable, tripID, trip)
}

func (tdao *SQLTripDAO) ListUserTripIDs(userID string) ([]string, error) {
	return tdao.sql.ListIDs(constants.TripTable, []*database.QueryParam{{FieldName: userIDFieldName, Operand: "=", Value: userID}})
}

func (tdao *SQLTripDAO) ListUserTripIDsByTime(userID string, startTime time.Time, endTime time.Time) ([]string, error) {
	overLappingStartQuery := []*database.QueryParam{
		{
			FieldName: userIDFieldName,
			Operand:   "=",
			Value:     userID,
		},
		{
			FieldName: endTimeFieldName,
			Operand:   ">=",
			Value:     util.TimeToMilliseconds(startTime),
		},
	}

	overLappingEndQuery := []*database.QueryParam{
		{
			FieldName: userIDFieldName,
			Operand:   "=",
			Value:     userID,
		},
		{
			FieldName: startTimeFieldName,
			Operand:   "<=",
			Value:     util.TimeToMilliseconds(endTime),
		},
	}

	overLappingStartIDs, err := tdao.sql.ListIDs(constants.TripTable, overLappingStartQuery)
	if err != nil {
		return nil, fmt.Errorf("error listing tripIDs between %q and %q for user %q: %w", startTime, endTime, userID, err)
	}
	overLappingEndIDs, err := tdao.sql.ListIDs(constants.TripTable, overLappingEndQuery)
	if err != nil {
		return nil, fmt.Errorf("error listing tripIDs between %q and %q for user %q: %w", startTime, endTime, userID, err)
	}

	uniqueTripIDs := map[string]int{}

	for _, sid := range overLappingStartIDs {
		uniqueTripIDs[sid]++
	}
	for _, eid := range overLappingEndIDs {
		uniqueTripIDs[eid]++
	}

	allTripIDs := []string{}
	for k, v := range uniqueTripIDs {
		if v > 1 {
			allTripIDs = append(allTripIDs, k)
		}
	}

	return allTripIDs, nil
}
func (tdao *SQLTripDAO) DeleteTripByID(ctx context.Context, tripID string) error {
	return tdao.sql.DeleteByID(constants.TripTable, tripID)
}
