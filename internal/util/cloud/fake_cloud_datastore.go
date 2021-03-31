package cloud

import (
	"errors"
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
	cutVideos             map[string]*types.CutVideo
	rawLocations          map[string]*types.RawLocation
	rawMotions            map[string]*types.RawMotion
	createTimeQueryOffset time.Duration
}

// NewFakeNoSQLDatabaseClient returns an instance of FakeNoSQLDatabaseClient.
func NewFakeNoSQLDatabaseClient(createTimeQueryOffset time.Duration) *FakeNoSQLDatabaseClient {
	return &FakeNoSQLDatabaseClient{
		rawVideos:             make(map[string]*types.RawVideo),
		cutVideos:             make(map[string]*types.CutVideo),
		rawLocations:          make(map[string]*types.RawLocation),
		rawMotions:            make(map[string]*types.RawMotion),
		createTimeQueryOffset: createTimeQueryOffset,
	}
}

// InsertRawVideo inserts the given *types.RawVideo into the internal rawVideos map.
func (fnsdc *FakeNoSQLDatabaseClient) InsertRawVideo(rawVideo *types.RawVideo) (string, error) {
	rawVideo.Id = generateUniqueID()
	fnsdc.rawVideos[rawVideo.Id] = rawVideo
	return rawVideo.Id, nil
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
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("raw video with userID %q and createTime %v not found in the store", userID, createTime))
	}

	return rawVideos[0], nil
}

// DeleteRawVideoByID deletes the rawVideo with the given ID from the internal rawVideos map.
func (fnsdc *FakeNoSQLDatabaseClient) DeleteRawVideoByID(id string) error {
	delete(fnsdc.rawVideos, id)
	return nil
}

// InsertUniqueRawVideo inserts the given *types.RawVideo into the internal rawVideos map, making
// sure it doesn't already exist.
func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueRawVideo(rawVideo *types.RawVideo) (string, error) {
	createTime := util.MillisecondsToTime(rawVideo.CreateTimeMs)

	existingRawVideo, err := fnsdc.GetRawVideo(rawVideo.UserId, createTime)

	var nfe *senecaerror.NotFoundError
	if err != nil && !errors.As(err, &nfe) {
		return "", fmt.Errorf("error checking if rawVideo alreadying exists - err: %w", err)
	}

	if existingRawVideo != nil {
		return "", senecaerror.NewUserError(rawVideo.UserId, fmt.Errorf("raw video with createTimeMs %d already exists", rawVideo.CreateTimeMs), fmt.Sprintf("Video at time %v already exists.", util.MillisecondsToTime(rawVideo.CreateTimeMs)))
	}

	return fnsdc.InsertRawVideo(rawVideo)
}

// GetCutVideo returns the *types.CutVideo for the given user at the given create time.
func (fnsdc *FakeNoSQLDatabaseClient) GetCutVideo(userID string, createTime time.Time) (*types.CutVideo, error) {
	beginTimeQuery := createTime.Add(-fnsdc.createTimeQueryOffset)
	endTimeQuery := createTime.Add(fnsdc.createTimeQueryOffset)
	beginTimeQueryMs := util.TimeToMilliseconds(beginTimeQuery)
	endTimeQueryMs := util.TimeToMilliseconds(endTimeQuery)

	cutVideos := []*types.CutVideo{}
	for _, v := range fnsdc.cutVideos {
		if v.CreateTimeMs >= beginTimeQueryMs && v.CreateTimeMs <= endTimeQueryMs {
			cutVideos = append(cutVideos, v)
		}
	}

	if len(cutVideos) > 1 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("more than one value for cutVideo for user ID %q and createTime %v", userID, createTime))
	}

	if len(cutVideos) == 0 {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("cut video with userID %q and createTime %v not found in the store", userID, createTime))
	}

	return cutVideos[0], nil
}

// DeleteCutVideoByID deletes the cutVideo with the given ID from the internal cutVideos map.
func (fnsdc *FakeNoSQLDatabaseClient) DeleteCutVideoByID(id string) error {
	delete(fnsdc.cutVideos, id)
	return nil
}

// InsertCutVideo inserts the given *types.CutVideo into the internal cutVideos map.
func (fnsdc *FakeNoSQLDatabaseClient) InsertCutVideo(cutVideo *types.CutVideo) (string, error) {
	cutVideo.Id = generateUniqueID()
	fnsdc.cutVideos[cutVideo.Id] = cutVideo
	return cutVideo.Id, nil
}

// InsertUniqueCutVideo inserts the given *types.CutVideo into the internal rawVideos map, making
// sure it doesn't already exist.
func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueCutVideo(cutVideo *types.CutVideo) (string, error) {
	createTime := util.MillisecondsToTime(cutVideo.CreateTimeMs)

	existingCutVideo, err := fnsdc.GetCutVideo(cutVideo.UserId, createTime)

	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		return "", fmt.Errorf("error checking if cutVideo alreadying exists - err: %w", err)
	}

	if existingCutVideo != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("cut video %v already exists", cutVideo))
	}

	return fnsdc.InsertCutVideo(cutVideo)
}

// GetRawLocation returns the *types.RawLocation for the given user at the given timestamp.
func (fnsdc *FakeNoSQLDatabaseClient) GetRawLocation(userID string, timestamp time.Time) (*types.RawLocation, error) {
	timestampMs := util.TimeToMilliseconds(timestamp)

	rawLocations := []*types.RawLocation{}
	for _, v := range fnsdc.rawLocations {
		if v.TimestampMs == timestampMs {
			rawLocations = append(rawLocations, v)
		}
	}

	if len(rawLocations) > 1 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("more than one value for cutVideo for user ID %q and createTime %v", userID, timestamp))
	}

	if len(rawLocations) == 0 {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("raw location with userID %q and createTime %v not found in the store", userID, timestamp))
	}

	return rawLocations[0], nil
}

// DeleteRawLocationByID deletes the raw location with the given ID from the internal rawLocations map.
func (fnsdc *FakeNoSQLDatabaseClient) DeleteRawLocationByID(id string) error {
	delete(fnsdc.rawLocations, id)
	return nil
}

// InsertRawLocation inserts the given *types.RawLocation into the internal rawLocations map.
func (fnsdc *FakeNoSQLDatabaseClient) InsertRawLocation(rawLocation *types.RawLocation) (string, error) {
	rawLocation.Id = generateUniqueID()
	fnsdc.rawLocations[rawLocation.Id] = rawLocation
	return rawLocation.Id, nil
}

// InsertUniqueRawLocation inserts the given *types.RawLocation into the internal rawLocations map, making
// sure it doesn't already exist.
func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueRawLocation(rawLocation *types.RawLocation) (string, error) {
	timestamp := util.MillisecondsToTime(rawLocation.TimestampMs)

	existingCutVideo, err := fnsdc.GetCutVideo(rawLocation.UserId, timestamp)

	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		return "", fmt.Errorf("error checking if rawLocation alreadying exists - err: %w", err)
	}

	if existingCutVideo != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("raw location %v already exists", rawLocation))
	}

	return fnsdc.InsertRawLocation(rawLocation)
}

// GetRawMotion returns the *types.RawMotion for the given user at the given timestamp.
func (fnsdc *FakeNoSQLDatabaseClient) GetRawMotion(userID string, timestamp time.Time) (*types.RawMotion, error) {
	timestampMs := util.TimeToMilliseconds(timestamp)

	rawMotions := []*types.RawMotion{}
	for _, v := range fnsdc.rawMotions {
		if v.TimestampMs == timestampMs {
			rawMotions = append(rawMotions, v)
		}
	}

	if len(rawMotions) > 1 {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("more than one value for cutVideo for user ID %q and createTime %v", userID, timestamp))
	}

	if len(rawMotions) == 0 {
		return nil, senecaerror.NewNotFoundError(fmt.Errorf("raw motion with userID %q and createTime %v not found in the store", userID, timestamp))
	}

	return rawMotions[0], nil
}

// DeleteRawMotionByID deletes the raw motion with the given ID from the internal rawMotions map.
func (fnsdc *FakeNoSQLDatabaseClient) DeleteRawMotionByID(id string) error {
	delete(fnsdc.rawMotions, id)
	return nil
}

// InsertRawMotion inserts the given *types.RawMotion into the internal rawMotions map.
func (fnsdc *FakeNoSQLDatabaseClient) InsertRawMotion(rawMotion *types.RawMotion) (string, error) {
	rawMotion.Id = generateUniqueID()
	fnsdc.rawMotions[rawMotion.Id] = rawMotion
	return rawMotion.Id, nil
}

// InsertUniqueRawMotion inserts the given *types.RawMotion into the internal rawMotions map, making
// sure it doesn't already exist.
func (fnsdc *FakeNoSQLDatabaseClient) InsertUniqueRawMotion(rawMotion *types.RawMotion) (string, error) {
	timestamp := util.MillisecondsToTime(rawMotion.TimestampMs)

	existingCutVideo, err := fnsdc.GetCutVideo(rawMotion.UserId, timestamp)

	var nfe *senecaerror.NotFoundError
	if !errors.As(err, &nfe) {
		return "", fmt.Errorf("error checking if rawMotion alreadying exists - err: %w", err)
	}

	if existingCutVideo != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("raw motion %v already exists", rawMotion))
	}

	return fnsdc.InsertRawMotion(rawMotion)
}

func generateUniqueID() string {
	id := rand.Int63()
	return strconv.FormatInt(id, 10)
}
