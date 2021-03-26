package cloud

import (
	"fmt"
	"math/rand"
	"seneca/api/types"
	"seneca/internal/util"
	"strconv"
	"time"
)

type FakeNoSQLDatabaseClient struct {
	// keyed by an ID
	rawVideos             map[string]*types.RawVideo
	createTimeQueryOffset time.Duration
}

func NewFakeNoSQLDatabaseClient(createTimeQueryOffset time.Duration) *FakeNoSQLDatabaseClient {
	return &FakeNoSQLDatabaseClient{
		rawVideos:             make(map[string]*types.RawVideo),
		createTimeQueryOffset: createTimeQueryOffset,
	}
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertRawVideo(rawVideo *types.RawVideo) (string, error) {
	id := rand.Int63()
	stringID := strconv.FormatInt(id, 10)

	rawVideo.Id = stringID

	fnsdc.rawVideos[rawVideo.Id] = rawVideo
	return stringID, nil
}

func (fnsdc *FakeNoSQLDatabaseClient) GetRawVideo(userID string, createTime time.Time) (*types.RawVideo, error) {
	beginTimeQuery := createTime.Add(-fnsdc.createTimeQueryOffset)
	endTimeQuery := createTime.Add(fnsdc.createTimeQueryOffset)
	beginTimeQueryMs := util.TimeToMilliseconds(&beginTimeQuery)
	endTimeQueryMs := util.TimeToMilliseconds(&endTimeQuery)

	rawVideos := []*types.RawVideo{}
	for _, v := range fnsdc.rawVideos {
		if v.CreateTimeMs >= beginTimeQueryMs && v.CreateTimeMs <= endTimeQueryMs {
			rawVideos = append(rawVideos, v)
		}
	}

	if len(rawVideos) > 1 {
		return nil, fmt.Errorf("more than one value for rawVideo for user ID %q and createTime %v", userID, createTime)
	}

	if len(rawVideos) == 0 {
		return nil, nil
	}

	return rawVideos[0], nil
}

func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueRawVideo(rawVideo *types.RawVideo) (string, error) {
	createTime := util.MillisecondsToTime(rawVideo.CreateTimeMs)

	existingRawVideo, err := fnsdc.GetRawVideo(rawVideo.UserId, createTime)
	if err != nil {
		return "", fmt.Errorf("error checking if rawVideo alreadying exists - err: %v", err)
	}

	if existingRawVideo != nil {
		return "", fmt.Errorf("raw video with createTimeMs %d and user ID %q already exists", rawVideo.CreateTimeMs, rawVideo.UserId)
	}

	return fnsdc.InsertRawVideo(rawVideo)
}

func (fnsdc *FakeNoSQLDatabaseClient) DeleteRawVideoByID(id string) error {
	delete(fnsdc.rawVideos, id)
	return nil
}
