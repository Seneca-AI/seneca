package rawlocationdao

import (
	"fmt"
	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
	"seneca/internal/dao"
)

const (
	userIDFieldName    = "UserId"
	timestampFieldName = "TimestampMs"
)

type SQLRawLocationDAO struct {
	sql dao.SQLInterface
}

func NewSQLRawLocationDAO(sqlInterface dao.SQLInterface) *SQLRawLocationDAO {
	return &SQLRawLocationDAO{
		sql: sqlInterface,
	}
}

func (rdao *SQLRawLocationDAO) InsertUniqueRawLocation(rawLocation *st.RawLocation) (*st.RawLocation, error) {
	ids, err := rdao.sql.ListIDs(constants.RawLocationsTable, []*cloud.QueryParam{
		{
			FieldName: userIDFieldName,
			Operand:   "=",
			Value:     rawLocation.UserId,
		},
		{
			FieldName: timestampFieldName,
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
	if err := rdao.sql.Insert(constants.RawLocationsTable, rawLocation.Id, rawLocation); err != nil {
		return nil, fmt.Errorf("error updating rawLocationID for rawLocation %v - err: %w", rawLocation, err)
	}

	rawLocation.Id = newRawLocationID
	return rawLocation, nil
}

func (rdao *SQLRawLocationDAO) GetRawLocationByID(id string) (*st.RawLocation, error) {
	rawLocationObj, err := rdao.sql.GetByID(constants.RawLocationsTable, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rawLocation %q by ID: %w", id, err)
	}

	if rawLocationObj == nil {
		return nil, nil
	}

	rawLocation, ok := rawLocationObj.(*st.RawLocation)
	if !ok {
		return nil, fmt.Errorf("expected RawLocation, got %T", rawLocationObj)
	}

	return rawLocation, nil
}

func (rdao *SQLRawLocationDAO) ListUserRawLocationIDs(userID string) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawLocationsTable, []*cloud.QueryParam{{FieldName: userIDFieldName, Operand: "=", Value: userID}})
}

func (rdao *SQLRawLocationDAO) DeleteRawLocationByID(id string) error {
	return rdao.sql.DeleteByID(constants.RawLocationsTable, id)
}
