package rawmotiondao

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

type SQLRawMotionDAO struct {
	sql dao.SQLInterface
}

func NewSQLRawMotionDAO(sqlInterface dao.SQLInterface) *SQLRawMotionDAO {
	return &SQLRawMotionDAO{
		sql: sqlInterface,
	}
}

func (rdao *SQLRawMotionDAO) InsertUniqueRawMotion(rawMotion *st.RawMotion) (*st.RawMotion, error) {
	ids, err := rdao.sql.ListIDs(constants.RawMotionsTable, []*cloud.QueryParam{
		{
			FieldName: userIDFieldName,
			Operand:   "=",
			Value:     rawMotion.UserId,
		},
		{
			FieldName: timestampFieldName,
			Operand:   "=",
			Value:     rawMotion.TimestampMs,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error checking for existing rawMotion %v - err: %w", rawMotion, err)
	}

	if len(ids) != 0 {
		return nil, fmt.Errorf("rawMotion with timestamp %d already exists for user %q", rawMotion.TimestampMs, rawMotion.UserId)
	}

	newRawMotionID, err := rdao.sql.Create(constants.RawMotionsTable, rawMotion)
	if err != nil {
		return nil, fmt.Errorf("error inserting rawMotion %v into store: %w", rawMotion, err)
	}
	rawMotion.Id = newRawMotionID

	// Now set the ID in the datastore object.
	if err := rdao.sql.Insert(constants.RawMotionsTable, rawMotion.Id, rawMotion); err != nil {
		return nil, fmt.Errorf("error updating rawMotionID for rawMotion %v - err: %w", rawMotion, err)
	}

	rawMotion.Id = newRawMotionID
	return rawMotion, nil
}

func (rdao *SQLRawMotionDAO) GetRawMotionByID(id string) (*st.RawMotion, error) {
	rawMotionObj, err := rdao.sql.GetByID(constants.RawMotionsTable, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rawMotion %q by ID: %w", id, err)
	}

	if rawMotionObj == nil {
		return nil, nil
	}

	rawMotion, ok := rawMotionObj.(*st.RawMotion)
	if !ok {
		return nil, fmt.Errorf("expected RawMotion, got %T", rawMotionObj)
	}

	return rawMotion, nil
}

func (rdao *SQLRawMotionDAO) ListUserRawMotionIDs(userID string) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawMotionsTable, []*cloud.QueryParam{{FieldName: userIDFieldName, Operand: "=", Value: userID}})
}

func (rdao *SQLRawMotionDAO) DeleteRawMotionByID(id string) error {
	return rdao.sql.DeleteByID(constants.RawMotionsTable, id)
}
