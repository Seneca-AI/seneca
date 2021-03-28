package cloud

import (
	"fmt"
	"math/rand"
	"seneca/api/senecaerror"
	"seneca/api/types"
	"seneca/internal/util"
	"strconv"
	"time"
)

// FakeNoSQLDatabaseClient implements a fake version of the NoSQLDatabaseInterface.
type FakeNoSQLDatabaseClient struct {
	// keyed by an ID
	rawVideos             map[string]*types.RawVideo
	createTimeQueryOffset time.Duration
}

// NewFakeNoSQLDatabaseClient returns an instance of FakeNoSQLDatabaseClient.
func NewFakeNoSQLDatabaseClient(createTimeQueryOffset time.Duration) *FakeNoSQLDatabaseClient {
	return &FakeNoSQLDatabaseClient{
		rawVideos:             make(map[string]*types.RawVideo),
		createTimeQueryOffset: createTimeQueryOffset,
	}
}

// InsertRawVideo inserts the given *types.RawVideo into the internal rawVideos map.
func (fnsdc *FakeNoSQLDatabaseClient) InsertRawVideo(rawVideo *types.RawVideo) (string, error) {
	id := rand.Int63()
	stringID := strconv.FormatInt(id, 10)

	rawVideo.Id = stringID

	fnsdc.rawVideos[rawVideo.Id] = rawVideo
	return stringID, nil
}

// GetRawVideo returns the *types.RawVideo for the given user at the given create time.
func (fnsdc *FakeNoSQLDatabaseClient) GetRawVideo(userID string, createTime time.Time) (*types.RawVideo, error) {
	beginTimeQuery := createTime.Add(-fnsdc.createTimeQueryOffset)
	endTimeQuery := createTime.Add(fnsdc.createTimeQueryOffset)
	beginTimeQueryMs := util.TimeToMilliseconds(beginTimeQuery)
	endTimeQueryMs := util.TimeToMilliseconds(endTimeQuery)

	rawVideos := []*types.RawVideo{}
	for _, v := range fnsdc.rawVideos {
		if v.CreateTimeMs >= beginTimeQueryMs && v.CreateTimeMs <= endTimeQueryMs {
			rawVideos = append(rawVideos, v)
		}
	}

	if len(rawVideos) > 1 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("more than one value for rawVideo for user ID %q and createTime %v", userID, createTime))
	}

	if len(rawVideos) == 0 {
		return nil, nil
	}

	return rawVideos[0], nil
}

// InsertUniqueRawVideo inserts the given *types.RawVideo into the internal rawVideos map, making
// sure it doesn't already exist.
func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueRawVideo(rawVideo *types.RawVideo) (string, error) {
	createTime := util.MillisecondsToTime(rawVideo.CreateTimeMs)

	existingRawVideo, err := fnsdc.GetRawVideo(rawVideo.UserId, createTime)
	if err != nil {
		return "", fmt.Errorf("error checking if rawVideo alreadying exists - err: %w", err)
	}

	if existingRawVideo != nil {
		return "", senecaerror.NewUserError(rawVideo.UserId, fmt.Errorf("raw video with createTimeMs %d already exists", rawVideo.CreateTimeMs), fmt.Sprintf("Video at time %v already exists.", util.MillisecondsToTime(rawVideo.CreateTimeMs)))
	}

	return fnsdc.InsertRawVideo(rawVideo)
}

// DeleteRawVideoByID deletes the rawVideo with the given ID from the internal rawVideos map.
func (fnsdc *FakeNoSQLDatabaseClient) DeleteRawVideoByID(id string) error {
	delete(fnsdc.rawVideos, id)
	return nil
}
