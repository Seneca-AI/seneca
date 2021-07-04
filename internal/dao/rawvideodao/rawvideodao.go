package rawvideodao

// TODO(): known risks:
//	- No write failure handling (e.g. , if Insert fails Create is not cleaned up)
//	- No context

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/util"
	"time"
)

type SQLRawVideoDAO struct {
	sql                   database.SQLInterface
	logger                logging.LoggingInterface
	createTimeQueryOffset time.Duration
}

func NewSQLRawVideoDAO(sqlInterface database.SQLInterface, logger logging.LoggingInterface, createTimeQueryOffset time.Duration) *SQLRawVideoDAO {
	return &SQLRawVideoDAO{
		sql:                   sqlInterface,
		logger:                logger,
		createTimeQueryOffset: createTimeQueryOffset,
	}
}

func (rdao *SQLRawVideoDAO) InsertUniqueRawVideo(rawVideo *st.RawVideo) (*st.RawVideo, error) {
	params := append(database.GenerateTimeOffsetParams(constants.CreateTimeFieldName, rawVideo.CreateTimeMs, rdao.createTimeQueryOffset), &database.QueryParam{FieldName: constants.UserIDFieldName, Operand: "=", Value: rawVideo.UserId})

	ids, err := rdao.sql.ListIDs(constants.RawVideosTable, params)
	if err != nil {
		return nil, fmt.Errorf("error checking for existing rawVideo %v - err: %w", rawVideo, err)
	}

	if len(ids) != 0 {
		return nil, fmt.Errorf("rawVideo with timestamp %d already exists for user %q", rawVideo.CreateTimeMs, rawVideo.UserId)
	}

	newRawVideoID, err := rdao.sql.Create(constants.RawVideosTable, rawVideo)
	if err != nil {
		return nil, fmt.Errorf("error inserting rawVideo %v into store: %w", rawVideo, err)
	}
	rawVideo.Id = newRawVideoID
	// Now set the ID in the datastore object.
	if err := rdao.PutRawVideoByID(context.TODO(), rawVideo.Id, rawVideo); err != nil {
		return nil, fmt.Errorf("error updating rawVideoID for rawVideo %v - err: %w", rawVideo, err)
	}

	rawVideo.Id = newRawVideoID
	return rawVideo, nil
}

func (rdao *SQLRawVideoDAO) PutRawVideoByID(ctx context.Context, rawVideoID string, rawVideo *st.RawVideo) error {
	err := rdao.sql.Insert(constants.RawVideosTable, rawVideo.Id, rawVideo)
	if err == nil {
		rdao.logger.Log(
			fmt.Sprintf(
				"Put rawVideo for user %s between %v and %v",
				rawVideo.UserId,
				util.MillisecondsToTime(rawVideo.CreateTimeMs),
				util.MillisecondsToTime(rawVideo.CreateTimeMs+rawVideo.DurationMs),
			),
		)
	}
	return err
}

func (rdao *SQLRawVideoDAO) ListUnprocessedRawVideoIDs(userID string, latestVersion float64) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawVideosTable, []*database.QueryParam{{FieldName: constants.UserIDFieldName, Operand: "=", Value: userID}, {FieldName: constants.AlgosVersionFieldName, Operand: "<", Value: latestVersion}})
}

func (rdao *SQLRawVideoDAO) GetRawVideoByID(id string) (*st.RawVideo, error) {
	rawVideoObj, err := rdao.sql.GetByID(constants.RawVideosTable, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rawVideo %q by ID: %w", id, err)
	}

	if rawVideoObj == nil {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("rawVideo with ID %q not found in the store", id))
	}

	rawVideo, ok := rawVideoObj.(*st.RawVideo)
	if !ok {
		return nil, fmt.Errorf("expected RawVideo, got %T", rawVideoObj)
	}

	return rawVideo, nil
}

func (rdao *SQLRawVideoDAO) ListUserRawVideoIDs(userID string) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawVideosTable, []*database.QueryParam{{FieldName: constants.UserIDFieldName, Operand: "=", Value: userID}})
}

func (rdao *SQLRawVideoDAO) DeleteRawVideoByID(id string) error {
	return rdao.sql.DeleteByID(constants.RawVideosTable, id)
}
