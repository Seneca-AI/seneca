package rawmotiondao

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
)

type SQLRawMotionDAO struct {
	sql    database.SQLInterface
	logger logging.LoggingInterface
}

func NewSQLRawMotionDAO(sqlInterface database.SQLInterface, logger logging.LoggingInterface) *SQLRawMotionDAO {
	return &SQLRawMotionDAO{
		sql:    sqlInterface,
		logger: logger,
	}
}

func (rdao *SQLRawMotionDAO) InsertUniqueRawMotion(rawMotion *st.RawMotion) (*st.RawMotion, error) {
	ids, err := rdao.sql.ListIDs(constants.RawMotionsTable, []*database.QueryParam{
		{
			FieldName: constants.UserIDFieldName,
			Operand:   "=",
			Value:     rawMotion.UserId,
		},
		{
			FieldName: constants.TimestampFieldName,
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
	if err := rdao.PutRawMotionByID(context.TODO(), rawMotion.Id, rawMotion); err != nil {
		return nil, fmt.Errorf("error putting rawMotion %v - err: %w", rawMotion, err)
	}

	return rawMotion, nil
}

func (rdao *SQLRawMotionDAO) PutRawMotionByID(ctx context.Context, rawMotionID string, rawMotion *st.RawMotion) error {
	return rdao.sql.Insert(constants.RawMotionsTable, rawMotionID, rawMotion)
}

func (rdao *SQLRawMotionDAO) ListUnprocessedRawMotionIDs(userID string, latestVersion float64) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawMotionsTable, []*database.QueryParam{{FieldName: constants.UserIDFieldName, Operand: "=", Value: userID}, {FieldName: constants.AlgosVersionFieldName, Operand: "<", Value: latestVersion}})
}

func (rdao *SQLRawMotionDAO) GetRawMotionByID(id string) (*st.RawMotion, error) {
	rawMotionObj, err := rdao.sql.GetByID(constants.RawMotionsTable, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rawMotion %q by ID: %w", id, err)
	}

	if rawMotionObj == nil {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("rawLocation with ID %q not found in the store", id))
	}

	rawMotion, ok := rawMotionObj.(*st.RawMotion)
	if !ok {
		return nil, fmt.Errorf("expected RawMotion, got %T", rawMotionObj)
	}

	return rawMotion, nil
}

func (rdao *SQLRawMotionDAO) ListUserRawMotionIDs(userID string) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawMotionsTable, []*database.QueryParam{{FieldName: constants.UserIDFieldName, Operand: "=", Value: userID}})
}

func (rdao *SQLRawMotionDAO) DeleteRawMotionByID(id string) error {
	return rdao.sql.DeleteByID(constants.RawMotionsTable, id)
}
