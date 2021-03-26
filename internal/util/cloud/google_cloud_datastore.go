package cloud

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"seneca/api/types"
	"seneca/internal/util"

	"cloud.google.com/go/datastore"
)

const (
	// The "kind" of raw videos in Cloud Datstore.
	rawVideoKind        = "RawVideo"
	rawVideoDirName     = "RawVideos"
	directoryKind       = "Directory"
	createTimeFieldName = "CreateTimeMs"
	userIDFieldName     = "UserId"
)

var (
	rawVideoKey = datastore.Key{
		Kind: directoryKind,
		Name: rawVideoDirName,
	}
)

// GoogleCloudStorageClient implements NoSQLDatabaseInterface using the real
// Google Cloud Datastore.
type GoogleCloudDatastoreClient struct {
	client                *datastore.Client
	projectID             string
	createTimeQueryOffset time.Duration
}

// GoogleCloudDatastoreClient initializes a new Google datastore.Client with the given parameters.
// Params:
// 		ctx context.Context
// 		projectID string: the project
// Returns:
//		*GoogleCloudDatastoreClient: the client
// 		error
func NewGoogleCloudDatastoreClient(ctx context.Context, projectID string, createTimeQueryOffset time.Duration) (*GoogleCloudDatastoreClient, error) {
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error initializing new GoogleCloudDatastoreClient - err: %v", err)
	}
	return &GoogleCloudDatastoreClient{
		client:                client,
		projectID:             projectID,
		createTimeQueryOffset: createTimeQueryOffset,
	}, nil
}

// InsertRawVideo inserts the given *types.RawVideo into the RawVideos Directory.
// Params:
// 		rawVideo *types.RawVideo: the rawVideo
// Returns:
//		string: the newly generated datastore ID for the rawVideo
//		error
func (gcdc *GoogleCloudDatastoreClient) InsertRawVideo(rawVideo *types.RawVideo) (string, error) {
	key := datastore.IncompleteKey(rawVideoKind, &rawVideoKey)
	completeKey, err := gcdc.client.Put(context.Background(), key, rawVideo)
	if err != nil {
		return "", fmt.Errorf("error putting RawVideo entity for user ID %q - err: %v", rawVideo.UserId, err)
	}
	return strconv.FormatInt(completeKey.ID, 64), nil
}

// GetRawVideo gets the *types.RawVideo for the given user around the specified createTime.
// We search datastore for videos +/-1 second the specified createTime.
// Params:
//		userID string: the userID associated with this video
//		createTime time.Time: the approximate time the video was created
// Returns:
//		*types.RawVideo: the raw video object
//		error
func (gcdc *GoogleCloudDatastoreClient) GetRawVideo(userID string, createTime time.Time) (*types.RawVideo, error) {
	beginTimeQuery := createTime.Add(-gcdc.createTimeQueryOffset)
	endTimeQuery := createTime.Add(gcdc.createTimeQueryOffset)

	query := datastore.NewQuery(
		rawVideoKind,
	).Filter(
		fmt.Sprintf("%s%s", createTimeFieldName, ">="), util.TimeToMilliseconds(&beginTimeQuery),
	).Filter(
		fmt.Sprintf("%s%s", createTimeFieldName, "<="), util.TimeToMilliseconds(&endTimeQuery),
	).Filter(
		fmt.Sprintf("%s%s", userIDFieldName, "="), userID,
	)

	var rawVideoOut []*types.RawVideo

	_, err := gcdc.client.GetAll(context.Background(), query, &rawVideoOut)
	if err != nil {
		return nil, fmt.Errorf("error querying datastore for RawVideo entity for user ID %q and createTime %v - err: %v", userID, createTime, err)
	}

	if len(rawVideoOut) > 1 {
		return nil, fmt.Errorf("error querying datastore for RawVideo entity for user ID %q and createTime %v, more than one value returned", userID, createTime)
	}

	if len(rawVideoOut) < 1 {
		return nil, nil
	}

	return rawVideoOut[0], nil
}

// InsertUniqueRawVideo inserts the given *types.RawVideo into the RawVideos Directory if a
// similar RawVideo doesn't already exist.
// Params:
// 		rawVideo *types.RawVideo: the rawVideo
// Returns:
//		string: the newly generated datastore ID for the rawVideo
//		error
func (gcdc *GoogleCloudDatastoreClient) InsertUniqueRawVideo(rawVideo *types.RawVideo) (string, error) {
	existingRawVideo, err := gcdc.GetRawVideo(rawVideo.UserId, util.MillisecondsToTime(rawVideo.CreateTimeMs))
	if err != nil {
		return "", fmt.Errorf("error checking if raw video already exists - err: %v", err)
	}
	if existingRawVideo != nil {
		return "", fmt.Errorf("raw video for user %q with CreateTimeMs %d already exists", rawVideo.UserId, rawVideo.CreateTimeMs)
	}
	return gcdc.InsertRawVideo(rawVideo)
}

func (gcdc *GoogleCloudDatastoreClient) DeleteRawVideoByID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting id to int64 - err: %v", err)
	}

	key := datastore.IDKey(rawVideoKind, idInt, &rawVideoKey)
	if err := gcdc.client.Delete(context.Background(), key); err != nil {
		return fmt.Errorf("error deleting raw video by key - err: %v", err)
	}
	return nil
}
