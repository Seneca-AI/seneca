package datastore

import (
	"context"
	"errors"
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/database"
	"strconv"

	"cloud.google.com/go/datastore"
)

const (
	// "kind" is a Cloud Datastore concept.
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
		constants.UsersTable:            userKey,
		constants.RawVideosTable:        rawVideoKey,
		constants.RawLocationsTable:     rawLocationKey,
		constants.RawMotionsTable:       rawMotionKey,
		constants.EventTable:            eventKey,
		constants.TripTable:             tripKey,
		constants.DrivingConditionTable: drivingConditionKey,
	}
)

// 	Service defines the datastore service.
type Service struct {
	client    *datastore.Client
	projectID string
}

// 	New returns a new datastore Service object.
//	Params:
//		ctx context.Context
//		projectID string
//	Returns:
//		*gcpdatastore.Service
//		senecaerror.CloudError
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

// 	ListIDs lists the IDs of the objects that satisfy the query.
//	Params:
//		tableName constants.TableName
//		queryParams []*database.QueryParam: the parameters for the query to execute
//	Returns:
//		[]string: the list of IDs
//		senecaerror.DevError, senecaerror.CloudError
func (s *Service) ListIDs(tableName constants.TableName, queryParams []*database.QueryParam) ([]string, error) {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return nil, senecaerror.NewDevError(fmt.Errorf("no Datastore key found for table %q", tableName))
	}

	query := datastore.NewQuery(key.Kind).KeysOnly()

	for _, qp := range queryParams {
		query = query.Filter(fmt.Sprintf("%s%s", qp.FieldName, qp.Operand), qp.Value)
	}

	keys, err := s.client.GetAll(context.TODO(), query, nil)
	if err != nil {
		return nil, senecaerror.NewCloudError(fmt.Errorf("GetAll() for query %v returns err: %w", query, err))
	}

	ids := []string{}
	for _, k := range keys {
		ids = append(ids, fmt.Sprintf("%d", k.ID))
	}

	return ids, nil
}

//	GetByID gets the object with the given id from the table with the given tableName.
//	Params:
//		tableName constants.TableName
//		id string
//	Returns:
//		interface{}: the untyped object
//		senecaerror.DevError, senecaerror.BadStateError, senecaerror.CloudError, senecaerror.NotFoundError
func (s *Service) GetByID(tableName constants.TableName, id string) (interface{}, error) {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return nil, senecaerror.NewDevError(fmt.Errorf("no Datastore key found for table %q", tableName))
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("error parsing id %q into int64", id))
	}

	idKey := datastore.IDKey(key.Kind, idInt, &key)

	switch tableName {
	case constants.UsersTable:
		out := &st.User{}
		if err := s.get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.RawVideosTable:
		out := &st.RawVideo{}
		if err := s.get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.RawLocationsTable:
		out := &st.RawLocation{}
		if err := s.get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.RawMotionsTable:
		out := &st.RawMotion{}
		if err := s.get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.EventTable:
		out := &st.EventInternal{}
		if err := s.get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.DrivingConditionTable:
		out := &st.DrivingConditionInternal{}
		if err := s.get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	case constants.TripTable:
		out := &st.TripInternal{}
		if err := s.get(context.TODO(), idKey, out); err != nil {
			return nil, fmt.Errorf("error getting object with ID %q from table %q: %w", id, tableName, err)
		}
		return out, nil
	default:
		return nil, senecaerror.NewDevError(fmt.Errorf("getting type not supported for %q", tableName))
	}
}

// get adapts datastore errors to senecaerrors.
func (s *Service) get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	// TODO(lucaloncar): better error handling here
	err := s.client.Get(ctx, key, dst)
	if err != nil {
		dneErr := &datastore.ErrNoSuchEntity
		if errors.As(err, dneErr) {
			return senecaerror.NewNotFoundError(fmt.Errorf("object with key %v not found - err: %w", key, err))
		}
		return senecaerror.NewCloudError(fmt.Errorf("error getting object with key %v - err: %w", key, err))
	}
	return nil
}

func (s *Service) Create(tableName constants.TableName, object interface{}) (string, error) {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return "", senecaerror.NewDevError(fmt.Errorf("no Datastore key found for table %q", tableName))
	}

	incompleteKey := datastore.IncompleteKey(key.Kind, &key)

	fullKey, err := s.client.Put(context.TODO(), incompleteKey, object)
	if err != nil {
		return "", senecaerror.NewCloudError(fmt.Errorf("error putting %v for table %q", object, tableName))
	}

	return fmt.Sprintf("%d", fullKey.ID), nil
}

func (s *Service) Insert(tableName constants.TableName, id string, object interface{}) error {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return senecaerror.NewDevError(fmt.Errorf("no Datastore key found for table %q", tableName))
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error parsing id %q into int64", id))
	}

	idKey := datastore.IDKey(key.Kind, idInt, &key)
	if _, err = s.client.Put(context.TODO(), idKey, object); err != nil {
		return senecaerror.NewCloudError(fmt.Errorf("error putting %v with id %q for table %q - err: %w", object, id, tableName, err))
	}

	return nil
}

func (s *Service) DeleteByID(tableName constants.TableName, id string) error {
	key, ok := tableNameToDatastoreKey[tableName]
	if !ok {
		return senecaerror.NewDevError(fmt.Errorf("no Datastore key found for table %q", tableName))
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error parsing id %q into int64", id))
	}

	idKey := datastore.IDKey(key.Kind, idInt, &key)
	if err := s.client.Delete(context.TODO(), idKey); err != nil {
		return senecaerror.NewCloudError(fmt.Errorf("error deleting object with ID %q from table %q: %w", id, tableName, err))
	}

	return nil
}
