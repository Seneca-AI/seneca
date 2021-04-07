package data

import (
	"fmt"
	"seneca/api/types"
	"seneca/internal/util"
	"time"
)

// 	ConstructCutVideoData data generates the types.CutVideo list based on the input raw video and how long the cut videos should be.
//	Params:
//		cutVideoDuration time.Duration: how long each types.CutVideo should be
//		rawVideo *types.RawVideo: the raw video to cut
//	Returns:
//		[]*types.CutVideo
func ConstructCutVideoData(cutVideoDuration time.Duration, rawVideo *types.RawVideo) []*types.CutVideo {
	cutVideos := []*types.CutVideo{}

	remainingRawVideoDuration := util.MillisecondsToDuration(rawVideo.DurationMs)
	newCutVideoCreateTimeMs := rawVideo.CreateTimeMs
	for remainingRawVideoDuration > 0 {
		cutVidDur := cutVideoDuration
		if remainingRawVideoDuration < cutVideoDuration {
			cutVidDur = remainingRawVideoDuration
		}

		cutVideos = append(cutVideos, &types.CutVideo{
			UserId:               rawVideo.GetUserId(),
			CreateTimeMs:         newCutVideoCreateTimeMs,
			DurationMs:           cutVidDur.Milliseconds(),
			RawVideoId:           rawVideo.Id,
			CloudStorageFileName: fmt.Sprintf("%s.%d.CUT_VIDEO.mp4", rawVideo.UserId, newCutVideoCreateTimeMs),
		})

		newCutVideoCreateTimeMs += cutVidDur.Milliseconds()
		remainingRawVideoDuration -= cutVidDur
	}
	return cutVideos
}

//	ConstructRawLocationDatas construct a list of types.RawLocation from a list of types.Location and time.Time for the given userID.
//	Params:
//		userID string
//		locations []*types.Location
//		times	[]time.Time
//	Returns:
//		[]*types.RawLocation
//		error
func ConstructRawLocationDatas(userID string, locations []*types.Location, times []time.Time) ([]*types.RawLocation, error) {
	if len(locations) != len(times) {
		return nil, fmt.Errorf("locations has length %d, but times has legth %d", len(locations), len(times))
	}
	rawLocations := []*types.RawLocation{}
	for i := range locations {
		rawLocations = append(rawLocations, &types.RawLocation{
			UserId:      userID,
			Location:    locations[i],
			TimestampMs: util.TimeToMilliseconds(times[i]),
		})
	}
	return rawLocations, nil
}

//	ConstructRawMotionDatas construct a list of types.RawMotion from a list of types.Motion and time.Time for the given userID.
//	Params:
//		userID string
//		motions []*types.Motion
//		times	[]time.Time
//	Returns:
//		[]*types.RawMotion
//		error
func ConstructRawMotionDatas(userID string, motions []*types.Motion, times []time.Time) ([]*types.RawMotion, error) {
	if len(motions) != len(times) {
		return nil, fmt.Errorf("motions has length %d, but times has legth %d", len(motions), len(times))
	}
	rawMotions := []*types.RawMotion{}
	for i := range motions {
		rawMotions = append(rawMotions, &types.RawMotion{
			UserId:      userID,
			Motion:      motions[i],
			TimestampMs: util.TimeToMilliseconds(times[i]),
		})
	}
	return rawMotions, nil
}
