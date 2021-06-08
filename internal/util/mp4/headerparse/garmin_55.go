package headerparse

import (
	"errors"
	"fmt"
	st "seneca/api/type"
	"seneca/internal/util"
	"time"
)

type Garmin55ExifParser struct {
	unprocessedExifData *unprocessedExifData
}

func (prs *Garmin55ExifParser) init(unprocessedExifData *unprocessedExifData) {
	prs.unprocessedExifData = unprocessedExifData
}

func (prs *Garmin55ExifParser) parseOutRawVideoMetadata() (*st.RawVideo, error) {
	rawVideo := &st.RawVideo{}

	creationTimeMs, err := prs.getVideoCreationTime(prs.unprocessedExifData.startTime)
	if err != nil {
		return nil, fmt.Errorf("error getting creationTimeMs: %w", err)
	}
	rawVideo.CreateTimeMs = creationTimeMs

	durationMs, err := getDurationMs(prs.unprocessedExifData.duration)
	if err != nil {
		return nil, fmt.Errorf("error getting durationMs: %w", err)
	}
	rawVideo.DurationMs = durationMs

	return rawVideo, nil
}

func (prs *Garmin55ExifParser) getVideoCreationTime(timeString string) (int64, error) {
	t, err := time.Parse("2006:01:02 15:04:05", timeString)
	if err != nil {
		return 0, fmt.Errorf("error parsing CreationTime - err: %v", err)
	}
	t = t.In(time.UTC).Round(time.Second)

	if t.Equal(time.Unix(0, 0)) {
		return 0, errors.New("creationTime of 0 is not allowed")
	}
	t = t.In(time.UTC)

	return util.TimeToMilliseconds(t), nil
}

func (prs *Garmin55ExifParser) parseOutGPSMetadata(rawVideo *st.RawVideo) ([]*st.Location, []*st.Motion, []time.Time, error) {
	locations, motions, times, err := getLocationsMotionsTimes("2006:01:02 15:04:05.000Z", prs.unprocessedExifData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getLocationsMotionsTimes() returns err: %w", err)
	}

	if len(times) == 0 {
		return nil, nil, nil, fmt.Errorf("got 0 times")
	}

	tzOffset, err := getTZOffset(times[0], locations[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getTZOffset(%v, %v) returns err: %w", times[0], locations[0], err)
	}

	newTimes := []time.Time{}
	for _, t := range times {
		newTime := t.Add(tzOffset)
		newTimes = append(newTimes, newTime)
	}

	return locations, motions, newTimes, err
}
