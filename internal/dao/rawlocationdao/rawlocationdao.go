package rawlocationdao

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
)

type SQLRawLocationDAO struct {
	sql database.SQLInterface
}

func NewSQLRawLocationDAO(sqlInterface database.SQLInterface) *SQLRawLocationDAO {
	return &SQLRawLocationDAO{
		sql: sqlInterface,
	}
}

func (rdao *SQLRawLocationDAO) InsertUniqueRawLocation(rawLocation *st.RawLocation) (*st.RawLocation, error) {
	ids, err := rdao.sql.ListIDs(constants.RawLocationsTable, []*database.QueryParam{
		{
			FieldName: constants.UserIDFieldName,
			Operand:   "=",
			Value:     rawLocation.UserId,
		},
		{
			FieldName: constants.TimestampFieldName,
			Operand:   "=",
			Value:     rawLocation.TimestampMs,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error checking for existing rawLocation %v - err: %w", rawLocation, err)
	}

	if len(ids) != 0 {
		return nil, fmt.Errorf("rawLocation with timestamp %d already exists for user %q", rawLocation.TimestampMs, rawLocation.UserId)
	}

	newRawLocationID, err := rdao.sql.Create(constants.RawLocationsTable, rawLocation)
	if err != nil {
		return nil, fmt.Errorf("error inserting rawLocation %v into store: %w", rawLocation, err)
	}
	rawLocation.Id = newRawLocationID

	// Now set the ID in the datastore object.
	if err := rdao.PutRawLocationByID(context.TODO(), rawLocation.Id, rawLocation); err != nil {
		return nil, fmt.Errorf("error updating rawLocationID for rawLocation %v - err: %w", rawLocation, err)
	}

	rawLocation.Id = newRawLocationID
	return rawLocation, nil
}

func (rdao *SQLRawLocationDAO) PutRawLocationByID(ctx context.Context, rawLocationID string, rawLocation *st.RawLocation) error {
	return rdao.sql.Insert(constants.RawLocationsTable, rawLocationID, rawLocation)
}

func (rdao *SQLRawLocationDAO) GetRawLocationByID(id string) (*st.RawLocation, error) {
	rawLocationObj, err := rdao.sql.GetByID(constants.RawLocationsTable, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rawLocation %q by ID: %w", id, err)
	}

	if rawLocationObj == nil {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("rawLocation with ID %q not found in the store", id))
	}

	rawLocation, ok := rawLocationObj.(*st.RawLocation)
	if !ok {
		return nil, fmt.Errorf("expected RawLocation, got %T", rawLocationObj)
	}

	return rawLocation, nil
}

func (rdao *SQLRawLocationDAO) ListUserRawLocationIDs(userID string) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawLocationsTable, []*database.QueryParam{{FieldName: constants.UserIDFieldName, Operand: "=", Value: userID}})
}

func (rdao *SQLRawLocationDAO) ListUnprocessedRawLocationsIDs(userID string, latestVersion float64) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawLocationsTable, []*database.QueryParam{{FieldName: constants.UserIDFieldName, Operand: "=", Value: userID}, {FieldName: constants.AlgosVersionFieldName, Operand: "<", Value: latestVersion}})
}

func (rdao *SQLRawLocationDAO) DeleteRawLocationByID(id string) error {
	return rdao.sql.DeleteByID(constants.RawLocationsTable, id)
}
