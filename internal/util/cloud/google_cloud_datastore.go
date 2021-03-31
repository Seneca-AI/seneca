package cloud

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"seneca/api/senecaerror"
	"seneca/api/types"
	"seneca/internal/util"

	"cloud.google.com/go/datastore"
)

const (
	// "kind" is a Cloud Datstore concept.
	rawVideoKind        = "RawVideo"
	rawVideoDirName     = "RawVideos"
	cutVideoKind        = "CutVideo"
	cutVideoDirName     = "CutVideos"
	rawMotionKind       = "RawMotion"
	rawMotionDirName    = "RawMotions"
	rawLocationKind     = "RawLocation"
	rawLocationDirName  = "RawLocations"
	directoryKind       = "Directory"
	createTimeFieldName = "CreateTimeMs"
	userIDFieldName     = "UserId"
)

var (
	rawVideoKey = datastore.Key{
		Kind: directoryKind,
		Name: rawVideoDirName,
	}
	cutVideoKey = datastore.Key{
		Kind: directoryKind,
		Name: cutVideoDirName,
	}
	rawMotionKey = datastore.Key{
		Kind: rawMotionKind,
		Name: rawMotionDirName,
	}
	rawLocationKey = datastore.Key{
		Kind: rawLocationKind,
		Name: rawLocationDirName,
	}
)

// GoogleCloudDatastoreClient implements NoSQLDatabaseInterface using the real
// Google Cloud Datastore.
type GoogleCloudDatastoreClient struct {
	client                *datastore.Client
	projectID             string
	createTimeQueryOffset time.Duration
}

// NewGoogleCloudDatastoreClient initializes a new Google datastore.Client with the given parameters.
// Params:
// 		ctx context.Context
// 		projectID string: the project
// Returns:
//		*GoogleCloudDatastoreClient: the client
// 		error
func NewGoogleCloudDatastoreClient(ctx context.Context, projectID string, createTimeQueryOffset time.Duration) (*GoogleCloudDatastoreClient, error) {
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, senecaerror.NewCloudError(fmt.Errorf("error initializing new GoogleCloudDatastoreClient - err: %v", err))
	}
	return &GoogleCloudDatastoreClient{
		client:                client,
		projectID:             projectID,
		createTimeQueryOffset: createTimeQueryOffset,
	}, nil
}

// GetRawVideo gets the *types.RawVideo for the given user around the specified createTime.
// We search datastore for videos +/-createTimeQueryOffset the specified createTime.
func (gcdc *GoogleCloudDatastoreClient) GetRawVideo(userID string, createTime time.Time) (*types.RawVideo, error) {
	beginTimeQuery := createTime.Add(-gcdc.createTimeQueryOffset)
	endTimeQuery := createTime.Add(gcdc.createTimeQueryOffset)

	query := datastore.NewQuery(
		rawVideoKind,
	).Filter(
		fmt.Sprintf("%s%s", createTimeFieldName, ">="), util.TimeToMilliseconds(beginTimeQuery),
	).Filter(
		fmt.Sprintf("%s%s", createTimeFieldName, "<="), util.TimeToMilliseconds(endTimeQuery),
	).Filter(
		fmt.Sprintf("%s%s", userIDFieldName, "="), userID,
	)

	var rawVideoOut []*types.RawVideo

	_, err := gcdc.client.GetAll(context.Background(), query, &rawVideoOut)
	if err != nil {
		return nil, senecaerror.NewCloudError(fmt.Errorf("error querying datastore for RawVideo entity for user ID %q and createTime %v - err: %v", userID, createTime, err))
	}

	if len(rawVideoOut) > 1 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("error querying datastore for RawVideo entity for user ID %q and createTime %v, more than one value returned", userID, createTime))
	}

	if len(rawVideoOut) < 1 {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("rawVideo with userID %q and createTimeMs %d was not found", userID, util.TimeToMilliseconds(createTime)))
	}

	return rawVideoOut[0], nil
}

// InsertRawVideo inserts the given *types.RawVideo into the RawVideos Directory.
func (gcdc *GoogleCloudDatastoreClient) InsertRawVideo(rawVideo *types.RawVideo) (string, error) {
	key := datastore.IncompleteKey(rawVideoKind, &rawVideoKey)
	completeKey, err := gcdc.client.Put(context.Background(), key, rawVideo)
	if err != nil {
		return "", senecaerror.NewCloudError(fmt.Errorf("error putting RawVideo entity for user ID %q - err: %v", rawVideo.UserId, err))
	}
	return fmt.Sprintf("%d", completeKey.ID), nil
}

// InsertUniqueRawVideo inserts the given *types.RawVideo into the RawVideos Directory if a
// similar RawVideo doesn't already exist.
func (gcdc *GoogleCloudDatastoreClient) InsertUniqueRawVideo(rawVideo *types.RawVideo) (string, error) {
	existingRawVideo, err := gcdc.GetRawVideo(rawVideo.UserId, util.MillisecondsToTime(rawVideo.CreateTimeMs))

	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		return "", fmt.Errorf("error checking if raw video already exists - err: %w", err)
	}

	if existingRawVideo != nil {
		return "", senecaerror.NewUserError(rawVideo.UserId, fmt.Errorf("rawVideo with CreateTimeMs %d already exists", rawVideo.CreateTimeMs), fmt.Sprintf("Video at time %v already exists.", util.MillisecondsToTime(rawVideo.CreateTimeMs)))
	}
	return gcdc.InsertRawVideo(rawVideo)
}

// DeleteRawVideoByID deletes the rawVideoo with the given ID from the datastore.
func (gcdc *GoogleCloudDatastoreClient) DeleteRawVideoByID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
	}

	key := datastore.IDKey(rawVideoKind, idInt, &rawVideoKey)
	if err := gcdc.client.Delete(context.Background(), key); err != nil {
		return senecaerror.NewCloudError(fmt.Errorf("error deleting raw video by key - err: %v", err))
	}
	return nil
}

// GetCutVideo gets the *types.CutVideo for the given user around the specified createTime.
func (gcdc *GoogleCloudDatastoreClient) GetCutVideo(userID string, createTime time.Time) (*types.CutVideo, error) {
	query := datastore.NewQuery(cutVideoKind).Filter(fmt.Sprintf("%s%s", userIDFieldName, "="), userID)

	query = addTimeOffsetFilter(createTime, gcdc.createTimeQueryOffset, query)

	fmt.Printf("query: %v\n", query)

	var cutVideoOut []*types.CutVideo

	_, err := gcdc.client.GetAll(context.Background(), query, &cutVideoOut)
	if err != nil {
		return nil, senecaerror.NewCloudError(fmt.Errorf("error querying datastore for CutVideo entity for user ID %q and createTime %v - err: %v", userID, createTime, err))
	}

	if len(cutVideoOut) > 1 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("error querying datastore for CutVideo entity for user ID %q and createTime %v, more than one value returned", userID, createTime))
	}

	if len(cutVideoOut) < 1 {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("cutVideo with userID %q and createTimeMs %d was not found", userID, util.TimeToMilliseconds(createTime)))
	}

	return cutVideoOut[0], nil
}

// DeleteCutVideoByID deletes the raw video with the given ID from the datastore.
func (gcdc *GoogleCloudDatastoreClient) DeleteCutVideoByID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
	}

	key := datastore.IDKey(cutVideoKind, idInt, &cutVideoKey)
	if err := gcdc.client.Delete(context.Background(), key); err != nil {
		return senecaerror.NewCloudError(fmt.Errorf("error deleting cut video by key - err: %v", err))
	}
	return nil
}

// InsertCutVideo inserts the given *types.CutVideo into the CutVideos directory of the datastore.
func (gcdc *GoogleCloudDatastoreClient) InsertCutVideo(cutVideo *types.CutVideo) (string, error) {
	key := datastore.IncompleteKey(cutVideoKind, &cutVideoKey)
	completeKey, err := gcdc.client.Put(context.Background(), key, cutVideo)
	if err != nil {
		return "", senecaerror.NewCloudError(fmt.Errorf("error putting CutVideo entity for user ID %q - err: %v", cutVideo.UserId, err))
	}
	return fmt.Sprintf("%d", completeKey.ID), nil
}

// InsertUniqueCutVideo inserts the given *types.CutVideo if a CutVideo with a similar creation time doesn't already exist.
func (gcdc *GoogleCloudDatastoreClient) InsertUniqueCutVideo(cutVideo *types.CutVideo) (string, error) {
	existingCutVideo, err := gcdc.GetCutVideo(cutVideo.UserId, util.MillisecondsToTime(cutVideo.CreateTimeMs))
	var nfe *senecaerror.NotFoundError

	if !errors.As(err, &nfe) {
		return "", fmt.Errorf("error checking if cut video already exists - err: %w", err)
	}

	if existingCutVideo != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("cutVideo with CreateTimeMs %d for user %s already exists", cutVideo.CreateTimeMs, cutVideo.UserId))
	}
	return gcdc.InsertCutVideo(cutVideo)
}

// GetRawMotion gets the *types.RawMotion for the given user at the given timestamp.
func (gcdc *GoogleCloudDatastoreClient) GetRawMotion(userID string, timestamp time.Time) (*types.RawMotion, error) {
	query := datastore.NewQuery(rawMotionKind).Filter(fmt.Sprintf("%s%s", userIDFieldName, "="), userID).Filter("TimestampMs=", util.TimeToMilliseconds(timestamp))

	var rawMotionOut []*types.RawMotion

	_, err := gcdc.client.GetAll(context.Background(), query, &rawMotionOut)
	if err != nil {
		return nil, senecaerror.NewCloudError(fmt.Errorf("error querying datastore for RawMotion entity for user ID %q and timestamp %v - err: %v", userID, timestamp, err))
	}

	if len(rawMotionOut) > 1 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("error querying datastore for RawMotion entity for user ID %q and timestamp %v, more than one value returned", userID, timestamp))
	}

	if len(rawMotionOut) < 1 {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("types.RawMotion with userID %q and TimestampMs %d was not found", userID, util.TimeToMilliseconds(timestamp)))
	}

	return rawMotionOut[0], nil
}

// DeleteRawMotionByID deletes the raw motion with the given ID from the datastore.
func (gcdc *GoogleCloudDatastoreClient) DeleteRawMotionByID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
	}

	key := datastore.IDKey(rawMotionKind, idInt, &rawMotionKey)
	if err := gcdc.client.Delete(context.Background(), key); err != nil {
		return senecaerror.NewCloudError(fmt.Errorf("error deleting raw motion by key - err: %v", err))
	}
	return nil
}

// InsertRawMotion inserts the given *types.RawMotion into the RawMotions directory in the datastore.
func (gcdc *GoogleCloudDatastoreClient) InsertRawMotion(rawMotion *types.RawMotion) (string, error) {
	key := datastore.IncompleteKey(rawMotionKind, &rawMotionKey)
	completeKey, err := gcdc.client.Put(context.Background(), key, rawMotion)
	if err != nil {
		return "", senecaerror.NewCloudError(fmt.Errorf("error putting RawMotion entity for user ID %q - err: %v", rawMotion.UserId, err))
	}

	return fmt.Sprintf("%d", completeKey.ID), nil
}

// InsertUniqueRawMotion inserts the given *types.RawMotion if a RawMotion with the same creation time doesn't already exist.
func (gcdc *GoogleCloudDatastoreClient) InsertUniqueRawMotion(rawMotion *types.RawMotion) (string, error) {
	existingRawMotion, err := gcdc.GetRawMotion(rawMotion.UserId, util.MillisecondsToTime(rawMotion.TimestampMs))
	if err != nil {
		return "", fmt.Errorf("error checking if RawMotion already exists - err: %w", err)
	}
	if existingRawMotion != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("rawMotion with timestamp %d for user %s already exists", rawMotion.TimestampMs, rawMotion.UserId))
	}
	return gcdc.InsertRawMotion(rawMotion)
}

// GetRawLocation gets the *types.RawLocation for the given user at the specified timestamp.
func (gcdc *GoogleCloudDatastoreClient) GetRawLocation(userID string, timestamp time.Time) (*types.RawLocation, error) {
	query := datastore.NewQuery(rawLocationKind).Filter(fmt.Sprintf("%s %s", userIDFieldName, "="), userID).Filter("TimestampMs =", util.TimeToMilliseconds(timestamp))

	var rawLocationOut []*types.RawLocation

	_, err := gcdc.client.GetAll(context.Background(), query, &rawLocationOut)
	if err != nil {
		return nil, senecaerror.NewCloudError(fmt.Errorf("error querying datastore for RawLocation entity for user ID %q and timestamp %v - err: %v", userID, timestamp, err))
	}

	if len(rawLocationOut) > 1 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("error querying datastore for RawLocation entity for user ID %q and timestamp %v, more than one value returned", userID, timestamp))
	}

	if len(rawLocationOut) < 1 {
		return nil, nil
	}

	return rawLocationOut[0], nil
}

// DeleteRawLocationByID deletes the raw location with the given ID from the datastore.
func (gcdc *GoogleCloudDatastoreClient) DeleteRawLocationByID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error converting id to int64 - err: %v", err))
	}

	key := datastore.IDKey(rawLocationKind, idInt, &rawLocationKey)
	if err := gcdc.client.Delete(context.Background(), key); err != nil {
		return senecaerror.NewCloudError(fmt.Errorf("error deleting raw location by key - err: %v", err))
	}
	return nil
}

// InsertRawLocation inserts the given *types.RawLocation into the RawLocations directory.
func (gcdc *GoogleCloudDatastoreClient) InsertRawLocation(rawLocation *types.RawLocation) (string, error) {
	key := datastore.IncompleteKey(rawLocationKind, &rawLocationKey)
	completeKey, err := gcdc.client.Put(context.Background(), key, rawLocation)
	if err != nil {
		return "", senecaerror.NewCloudError(fmt.Errorf("error putting RawLocation entity for user ID %q - err: %v", rawLocation.UserId, err))
	}
	return fmt.Sprintf("%d", completeKey.ID), nil
}

// InsertUniqueRawLocation inserts the given *types.RawLocation if a RawLocation with the same creation time doesn't already exist.
func (gcdc *GoogleCloudDatastoreClient) InsertUniqueRawLocation(rawLocation *types.RawLocation) (string, error) {
	existingRawLocation, err := gcdc.GetRawLocation(rawLocation.UserId, util.MillisecondsToTime(rawLocation.TimestampMs))
	if err != nil {
		return "", fmt.Errorf("error checking if RawLocation already exists - err: %w", err)
	}
	if existingRawLocation != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("rawLocation with timestamp %d for user %s already exists", rawLocation.TimestampMs, rawLocation.UserId))
	}
	return gcdc.InsertRawLocation(rawLocation)
}

func addTimeOffsetFilter(createTime time.Time, offset time.Duration, query *datastore.Query) *datastore.Query {
	beginTimeQuery := createTime.Add(-offset)
	endTimeQuery := createTime.Add(offset)

	return query.Filter(
		fmt.Sprintf("%s%s", createTimeFieldName, ">="), util.TimeToMilliseconds(beginTimeQuery),
	).Filter(
		fmt.Sprintf("%s%s", createTimeFieldName, "<="), util.TimeToMilliseconds(endTimeQuery),
	)
}
