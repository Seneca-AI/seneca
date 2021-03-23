package gcp_util

import (
	"context"
	"fmt"
	"time"

	"seneca/source/api/types"
	"seneca/source/util"

	"cloud.google.com/go/datastore"
)

const (
	// The "kind" of raw videos in Cloud Datstore.
	rawVideoKind    = "RawVideo"
	rawVideoDirName = "RawVideos"
	directoryKind   = "Directory"
	// queryOffset is the offset we allow for querying the CreatTime of the video.
	queryOffset         = time.Second
	createTimeFieldName = "CreateTimeMs"
	userIDFieldName     = "UserId"
)

var (
	RawVideoKey = datastore.Key{
		Kind: directoryKind,
		Name: rawVideoDirName,
	}
)

// GoogleCloudStorageClient is the client used for interfacing with
// Google Cloud Storage across Seneca.
type GoogleCloudDatastoreClient struct {
	client    *datastore.Client
	projectID string
}

// GoogleCloudDatastoreClient initializes a new Google datastore.Client with the given parameters.
// Params:
// 		ctx context.Context
// 		projectID string: the project
// Returns:
//		*GoogleCloudDatastoreClient: the client
// 		error
func NewGoogleCloudDatastoreClient(ctx context.Context, projectID string) (*GoogleCloudDatastoreClient, error) {
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error initializing new GoogleCloudDatastoreClient - err: %v", err)
	}
	return &GoogleCloudDatastoreClient{
		client:    client,
		projectID: projectID,
	}, nil
}

// InsertRawVideo inserts the given *types.RawVideo into the RawVideos Directory.
// Params:
// 		rawVideo *types.RawVideo: the rawVideo
// Returns:
//		int64: the newly generated datastore ID for the rawVideo
//		error
func (gcdc *GoogleCloudDatastoreClient) InsertRawVideo(rawVideo *types.RawVideo) (int64, error) {
	key := datastore.IncompleteKey(rawVideoKind, &RawVideoKey)
	completeKey, err := gcdc.client.Put(context.Background(), key, rawVideo)
	if err != nil {
		return 0, fmt.Errorf("error putting RawVideo entity for user ID %q - err: %v", rawVideo.UserId, err)
	}
	return completeKey.ID, nil
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
	beginTimeQuery := createTime.Add(-queryOffset)
	endTimeQuery := createTime.Add(queryOffset)

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
//		int64: the newly generated datastore ID for the rawVideo
//		error
func (gcdc *GoogleCloudDatastoreClient) InsertUniqueRawVideo(rawVideo *types.RawVideo) (int64, error) {
	existingRawVideo, err := gcdc.GetRawVideo(rawVideo.UserId, util.MillisecondsToTime(rawVideo.CreateTimeMs))
	if err != nil {
		return 0, fmt.Errorf("error checking if raw video already exists - err: %v", err)
	}
	if existingRawVideo != nil {
		return 0, fmt.Errorf("raw video for user %q with CreateTimeMs %d already exists", rawVideo.UserId, rawVideo.CreateTimeMs)
	}
	return gcdc.InsertRawVideo(rawVideo)
}

func (gcdc *GoogleCloudDatastoreClient) DeleteRawVideoByID(id int64) error {
	key := datastore.IDKey(rawVideoKind, id, &RawVideoKey)
	if err := gcdc.client.Delete(context.Background(), key); err != nil {
		return fmt.Errorf("error deleting raw video by key - err: %v", err)
	}
	return nil
}
