package rawframedao

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
)

type SQLRawFrameDAO struct {
	sql database.SQLInterface
}

func NewSQLRawFrameDAO(sqlInterface database.SQLInterface) *SQLRawFrameDAO {
	return &SQLRawFrameDAO{
		sql: sqlInterface,
	}
}

func (rdao *SQLRawFrameDAO) InsertUniqueRawFrame(RawFrame *st.RawFrame) (*st.RawFrame, error) {
	ids, err := rdao.sql.ListIDs(constants.RawFramesTable, []*database.QueryParam{
		{
			FieldName: constants.UserIDFieldName,
			Operand:   "=",
			Value:     RawFrame.UserId,
		},
		{
			FieldName: constants.TimestampFieldName,
			Operand:   "=",
			Value:     RawFrame.TimestampMs,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error checking for existing RawFrame %v - err: %w", RawFrame, err)
	}

	if len(ids) != 0 {
		return nil, fmt.Errorf("RawFrame with timestamp %d already exists for user %q", RawFrame.TimestampMs, RawFrame.UserId)
	}

	newRawFrameID, err := rdao.sql.Create(constants.RawFramesTable, RawFrame)
	if err != nil {
		return nil, fmt.Errorf("error inserting RawFrame %v into store: %w", RawFrame, err)
	}
	RawFrame.Id = newRawFrameID

	// Now set the ID in the datastore object.
	if err := rdao.PutRawFrameByID(context.TODO(), RawFrame.Id, RawFrame); err != nil {
		return nil, fmt.Errorf("error updating RawFrameID for RawFrame %v - err: %w", RawFrame, err)
	}

	RawFrame.Id = newRawFrameID
	return RawFrame, nil
}

func (rdao *SQLRawFrameDAO) PutRawFrameByID(ctx context.Context, RawFrameID string, RawFrame *st.RawFrame) error {
	return rdao.sql.Insert(constants.RawFramesTable, RawFrameID, RawFrame)
}

func (rdao *SQLRawFrameDAO) GetRawFrameByID(id string) (*st.RawFrame, error) {
	RawFrameObj, err := rdao.sql.GetByID(constants.RawFramesTable, id)
	if err != nil {
		return nil, fmt.Errorf("error getting RawFrame %q by ID: %w", id, err)
	}

	if RawFrameObj == nil {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("RawFrame with ID %q not found in the store", id))
	}

	RawFrame, ok := RawFrameObj.(*st.RawFrame)
	if !ok {
		return nil, fmt.Errorf("expected RawFrame, got %T", RawFrameObj)
	}

	return RawFrame, nil
}

func (rdao *SQLRawFrameDAO) ListUserRawFrameIDs(userID string) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawFramesTable, []*database.QueryParam{{FieldName: constants.UserIDFieldName, Operand: "=", Value: userID}})
}

func (rdao *SQLRawFrameDAO) ListUnprocessedRawFramesIDs(userID string, latestVersion float64) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawFramesTable, []*database.QueryParam{{FieldName: constants.UserIDFieldName, Operand: "=", Value: userID}, {FieldName: constants.AlgosVersionFieldName, Operand: "<", Value: latestVersion}})
}

func (rdao *SQLRawFrameDAO) DeleteRawFrameByID(id string) error {
	return rdao.sql.DeleteByID(constants.RawFramesTable, id)
}
