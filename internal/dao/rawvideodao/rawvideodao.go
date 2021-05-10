package rawvideodao

import (
	"fmt"
	"seneca/api/constants"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
	"seneca/internal/dao"
	"time"
)

const (
	userIDFieldName     = "UserId"
	createTimeFieldName = "CreateTimeMs"
)

type SQLRawVideoDAO struct {
	sql                   dao.SQLInterface
	createTimeQueryOffset time.Duration
}

func NewSQLRawVideoDAO(sqlInterface dao.SQLInterface, createTimeQueryOffset time.Duration) *SQLRawVideoDAO {
	return &SQLRawVideoDAO{
		sql:                   sqlInterface,
		createTimeQueryOffset: createTimeQueryOffset,
	}
}

func (rdao *SQLRawVideoDAO) InsertUniqueRawVideo(rawVideo *st.RawVideo) (*st.RawVideo, error) {
	params := append(cloud.GenerateTimeOffsetParams(createTimeFieldName, rawVideo.CreateTimeMs, rdao.createTimeQueryOffset), &cloud.QueryParam{FieldName: userIDFieldName, Operand: "=", Value: rawVideo.UserId})

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
	if err := rdao.sql.Insert(constants.RawVideosTable, rawVideo.Id, rawVideo); err != nil {
		return nil, fmt.Errorf("error updating rawVideoID for rawVideo %v - err: %w", rawVideo, err)
	}

	rawVideo.Id = newRawVideoID
	return rawVideo, nil
}

func (rdao *SQLRawVideoDAO) GetRawVideoByID(id string) (*st.RawVideo, error) {
	rawVideoObj, err := rdao.sql.GetByID(constants.RawVideosTable, id)
	if err != nil {
		return nil, fmt.Errorf("error getting rawVideo %q by ID: %w", id, err)
	}

	if rawVideoObj == nil {
		return nil, nil
	}

	rawVideo, ok := rawVideoObj.(*st.RawVideo)
	if !ok {
		return nil, fmt.Errorf("expected RawVideo, got %T", rawVideoObj)
	}

	return rawVideo, nil
}

func (rdao *SQLRawVideoDAO) ListUserRawVideoIDs(userID string) ([]string, error) {
	return rdao.sql.ListIDs(constants.RawVideosTable, []*cloud.QueryParam{{FieldName: userIDFieldName, Operand: "=", Value: userID}})
}

func (rdao *SQLRawVideoDAO) DeleteRawVideoByID(id string) error {
	return rdao.sql.DeleteByID(constants.RawVideosTable, id)
}
