package gcpdatastore

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
	"strconv"

	"cloud.google.com/go/datastore"
)

const (
	// "kind" is a Cloud Datstore concept.
	rawVideoKind        = "RawVideo"
	rawVideoDirName     = "RawVideos"
	rawMotionKind       = "RawMotion"
	rawMotionDirName    = "RawMotions"
	rawLocationKind     = "RawLocation"
	rawLocationDirName  = "RawLocations"
	userKind            = "User"
	userDirName         = "Users"
	directoryKind       = "Directory"
	createTimeFieldName = "CreateTimeMs"
	userIDFieldName     = "UserId"
	emailFieldName      = "Email"
)

var (
	rawVideoKey = datastore.Key{
		Kind: rawVideoKind,
		Name: rawVideoDirName,
	}
	rawMotionKey = datastore.Key{
		Kind: rawMotionKind,
		Name: rawMotionDirName,
	}
	rawLocationKey = datastore.Key{
		Kind: rawLocationKind,
		Name: rawLocationDirName,
	}
	userKey = datastore.Key{
		Kind: userKind,
		Name: userDirName,
	}

	tableNameToDatastoreKey = map[constants.TableName]datastore.Key{
		constants.UsersTable:        userKey,
		constants.RawVideosTable:    rawVideoKey,
		constants.RawLocationsTable: rawLocationKey,
		constants.RawMotionsTable:   rawMotionKey,
	}
)

type Service struct {
	client    *datastore.Client
	projectID string
}

func New(ctx context.Context, projectID string) (*Service, error) {
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, senecaerror.NewCloudError(fmt.Errorf("error initializing new gcpdatastore Service - err: %v", err))
	}
	return &Service{
		client:    client,
		projectID: projectID,
	}, nil
}

func (s *Service) ListIDs(tableName constants.TableName, queryParams []*cloud.QueryParam) ([]string, error) {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return nil, fmt.Errorf("no Datstore key found for table %q", tableName)
	}

	query := datastore.NewQuery(key.Kind).KeysOnly()

	for _, qp := range queryParams {
		query = query.Filter(fmt.Sprintf("%s%s", qp.FieldName, qp.Operand), qp.Value)
	}

	keys, err := s.client.GetAll(context.TODO(), query, nil)
	if err != nil {
		return nil, fmt.Errorf("GetAll() for query %v returns err: %w", query, err)
	}

	ids := []string{}
	for _, k := range keys {
		ids = append(ids, fmt.Sprintf("%d", k.ID))
	}

	return ids, nil
}

func (s *Service) GetByID(tableName constants.TableName, id string) (interface{}, error) {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return nil, fmt.Errorf("no Datstore key found for table %q", tableName)
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing id %q into int64", id)
	}

	idKey := datastore.IDKey(key.Kind, idInt, &key)

	switch tableName {
	case constants.UsersTable:
		out := &st.User{}
		if err := s.client.Get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.RawVideosTable:
		out := &st.RawVideo{}
		if err := s.client.Get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.RawLocationsTable:
		out := &st.RawLocation{}
		if err := s.client.Get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.RawMotionsTable:
		out := &st.RawMotion{}
		if err := s.client.Get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	default:
		// TODO(lucaloncar): make a DevError for this
		return nil, fmt.Errorf("getting type not supported for %q", tableName)
	}
}

func (s *Service) Insert(tableName constants.TableName, object interface{}) (string, error) {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return "", fmt.Errorf("no Datstore key found for table %q", tableName)
	}

	fullKey, err := s.client.Put(context.TODO(), &key, object)
	if err != nil {
		return "", fmt.Errorf("error putting %v for table %q", object, tableName)
	}

	return fmt.Sprintf("%d", fullKey.ID), nil
}

func (s *Service) DeleteByID(tableName constants.TableName, id string) error {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return fmt.Errorf("no Datstore key found for table %q", tableName)
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing id %q into int64", id)
	}

	idKey := datastore.IDKey(key.Kind, idInt, &key)
	if err := s.client.Delete(context.TODO(), idKey); err != nil {
		return fmt.Errorf("error deleting object with ID %q from table %q: %w", id, tableName, err)
	}

	return nil
}
