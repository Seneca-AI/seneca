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
	drivingConditionKind = "DrivingCondition"
	tripKind             = "Trip"
	eventKind            = "Event"
	rawVideoKind         = "RawVideo"
	rawMotionKind        = "RawMotion"
	rawLocationKind      = "RawLocation"
	userKind             = "User"
)

var (
	tripKey = datastore.Key{
		Kind: tripKind,
		Name: constants.TripTable.String(),
	}
	drivingConditionKey = datastore.Key{
		Kind: drivingConditionKind,
		Name: constants.DrivingConditionTable.String(),
	}
	eventKey = datastore.Key{
		Kind: eventKind,
		Name: constants.EventTable.String(),
	}
	rawVideoKey = datastore.Key{
		Kind: rawVideoKind,
		Name: constants.RawVideosTable.String(),
	}
	rawMotionKey = datastore.Key{
		Kind: rawMotionKind,
		Name: string(constants.RawMotionsTable),
	}
	rawLocationKey = datastore.Key{
		Kind: rawLocationKind,
		Name: constants.RawLocationsTable.String(),
	}
	userKey = datastore.Key{
		Kind: userKind,
		Name: constants.UsersTable.String(),
	}

	tableNameToDatastoreKey = map[constants.TableName]datastore.Key{
		constants.UsersTable:        userKey,
		constants.RawVideosTable:    rawVideoKey,
		constants.RawLocationsTable: rawLocationKey,
		constants.RawMotionsTable:   rawMotionKey,
		constants.EventTable:        eventKey,
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
	case constants.EventTable:
		out := &st.EventInternal{}
		if err := s.client.Get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.DrivingConditionTable:
		out := &st.DrivingConditionInternal{}
		if err := s.client.Get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.TripTable:
		out := &st.TripInternal{}
		if err := s.client.Get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	default:
		// TODO(lucaloncar): make a DevError for this
		return nil, fmt.Errorf("getting type not supported for %q", tableName)
	}
}

func (s *Service) Create(tableName constants.TableName, object interface{}) (string, error) {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return "", fmt.Errorf("no Datstore key found for table %q", tableName)
	}

	incompleteKey := datastore.IncompleteKey(key.Kind, &key)

	fullKey, err := s.client.Put(context.TODO(), incompleteKey, object)
	if err != nil {
		return "", fmt.Errorf("error putting %v for table %q", object, tableName)
	}

	return fmt.Sprintf("%d", fullKey.ID), nil
}

func (s *Service) Insert(tableName constants.TableName, id string, object interface{}) error {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return fmt.Errorf("no Datstore key found for table %q", tableName)
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing id %q - err: %w", id, err)
	}

	idKey := datastore.IDKey(key.Kind, idInt, &key)
	if _, err = s.client.Put(context.TODO(), idKey, object); err != nil {
		return fmt.Errorf("error putting %v with id %q for table %q - err: %w", object, id, tableName, err)
	}

	return nil
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
