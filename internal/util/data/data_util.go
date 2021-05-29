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
func ConstructRawLocationDatas(userID string, source *st.Source, locations []*st.Location, times []time.Time) ([]*st.RawLocation, error) {
	if len(locations) != len(times) {
		return nil, fmt.Errorf("locations has length %d, but times has legth %d", len(locations), len(times))
	}
	rawLocations := []*st.RawLocation{}
	for i := range locations {
		rawLocations = append(rawLocations, &st.RawLocation{
			UserId:      userID,
			Location:    locations[i],
			TimestampMs: util.TimeToMilliseconds(times[i]),
			Source:      source,
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
func ConstructRawMotionDatas(userID string, source *st.Source, motions []*st.Motion, times []time.Time) ([]*st.RawMotion, error) {
	// TODO(lucaloncar): look for RawMotion right before and right after this list since beginning motions will always have
	// zero acceleration
	if len(motions) != len(times) {
		return nil, fmt.Errorf("motions has length %d, but times has legth %d", len(motions), len(times))
	}
	rawMotions := []*st.RawMotion{}
	for i := range motions {
		rawMotions = append(rawMotions, &st.RawMotion{
			UserId:      userID,
			Motion:      motions[i],
			TimestampMs: util.TimeToMilliseconds(times[i]),
			Source:      source,
		})
	}
	return rawMotions, nil
}

func RawVideosEqual(lhs *st.RawVideo, rhs *st.RawVideo) error {
	errorString := ""
	if lhs.UserId != rhs.UserId {
		errorString = errorString + fmt.Sprintf("UserID: %q != UserID: %q", lhs.UserId, rhs.UserId)
	}

	if lhs.CreateTimeMs != rhs.CreateTimeMs {
		errorString = errorString + fmt.Sprintf(" CreateTimeMs: %d != CreateTimeMs: %d", lhs.CreateTimeMs, rhs.CreateTimeMs)
	}

	if lhs.DurationMs != rhs.DurationMs {
		errorString = errorString + fmt.Sprintf(" DurationMs: %d != DurationMs: %d", lhs.DurationMs, rhs.DurationMs)
	}

	if lhs.OriginalFileName != rhs.OriginalFileName {
		errorString = errorString + fmt.Sprintf(" OriginalFileName: %q != OriginalFileName: %q", lhs.OriginalFileName, rhs.OriginalFileName)
	}

	if errorString != "" {
		return fmt.Errorf(errorString)
	}

	return nil
}

// LocationsEqual compares the degrees and direction of the locations.
// Params:
//		l1 *st.Location
//		l2 *st.Location
// Returns:
//		bool
func LocationsEqual(l1 *st.Location, l2 *st.Location) bool {
	if l1 == nil || l2 == nil {
		return l1 == l2
	}
	if l1.Lat == nil || l2.Lat == nil {
		return l1.Lat == l2.Lat
	}
	if l1.Long == nil || l2.Long == nil {
		return l1.Long == l2.Long
	}
	return l1.Lat.Degrees == l2.Lat.Degrees && l1.Lat.DegreeMinutes == l2.Lat.DegreeMinutes && l1.Lat.DegreeSeconds == l2.Lat.DegreeSeconds && l1.Lat.LatDirection == l2.Lat.LatDirection &&
		l1.Long.Degrees == l2.Long.Degrees && l1.Long.DegreeMinutes == l2.Long.DegreeMinutes && l1.Long.DegreeSeconds == l2.Long.DegreeSeconds && l1.Long.LongDirection == l2.Long.LongDirection
}

func MotionsEqual(m1 *st.Motion, m2 *st.Motion) bool {
	if m1 == nil || m2 == nil {
		return m1 == m2
	}
	return m1.VelocityMph == m2.VelocityMph && m1.AccelerationMphS == m2.AccelerationMphS
}
