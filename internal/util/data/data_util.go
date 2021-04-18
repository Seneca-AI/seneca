package data

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/util"
	"time"
)

// 	ConstructCutVideoData data generates the st.CutVideo list based on the input raw video and how long the cut videos should be.
//	Params:
//		cutVideoDuration time.Duration: how long each st.CutVideo should be
//		rawVideo *st.RawVideo: the raw video to cut
//	Returns:
//		[]*st.CutVideo
func ConstructCutVideoData(cutVideoDuration time.Duration, rawVideo *st.RawVideo) []*st.CutVideo {
	cutVideos := []*st.CutVideo{}

	remainingRawVideoDuration := util.MillisecondsToDuration(rawVideo.DurationMs)
	newCutVideoCreateTimeMs := rawVideo.CreateTimeMs
	for remainingRawVideoDuration > 0 {
		cutVidDur := cutVideoDuration
		if remainingRawVideoDuration < cutVideoDuration {
			cutVidDur = remainingRawVideoDuration
		}

		cutVideos = append(cutVideos, &st.CutVideo{
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

//	ConstructRawLocationDatas construct a list of st.RawLocation from a list of st.Location and time.Time for the given userID.
//	Params:
//		userID string
//		locations []*st.Location
//		times	[]time.Time
//	Returns:
//		[]*st.RawLocation
//		error
func ConstructRawLocationDatas(userID string, locations []*st.Location, times []time.Time) ([]*st.RawLocation, error) {
	if len(locations) != len(times) {
		return nil, fmt.Errorf("locations has length %d, but times has legth %d", len(locations), len(times))
	}
	rawLocations := []*st.RawLocation{}
	for i := range locations {
		rawLocations = append(rawLocations, &st.RawLocation{
			UserId:      userID,
			Location:    locations[i],
			TimestampMs: util.TimeToMilliseconds(times[i]),
		})
	}
	return rawLocations, nil
}

//	ConstructRawMotionDatas construct a list of st.RawMotion from a list of st.Motion and time.Time for the given userID.
//	Params:
//		userID string
//		motions []*st.Motion
//		times	[]time.Time
//	Returns:
//		[]*st.RawMotion
//		error
func ConstructRawMotionDatas(userID string, motions []*st.Motion, times []time.Time) ([]*st.RawMotion, error) {
	if len(motions) != len(times) {
		return nil, fmt.Errorf("motions has length %d, but times has legth %d", len(motions), len(times))
	}
	rawMotions := []*st.RawMotion{}
	for i := range motions {
		rawMotions = append(rawMotions, &st.RawMotion{
			UserId:      userID,
			Motion:      motions[i],
			TimestampMs: util.TimeToMilliseconds(times[i]),
		})
	}
	return rawMotions, nil
}
