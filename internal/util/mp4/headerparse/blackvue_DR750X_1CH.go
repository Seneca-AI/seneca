package headerparse

import (
	"errors"
	"fmt"
	st "seneca/api/type"
	"seneca/internal/util"
	"time"
)

type blackVueDR750X1CHExifParser struct {
	unprocessedExifData *unprocessedExifData
}

func (prs *blackVueDR750X1CHExifParser) init(unprocessedExifData *unprocessedExifData) {
	prs.unprocessedExifData = unprocessedExifData
}

func (prs *blackVueDR750X1CHExifParser) parseOutRawVideoMetadata() (*st.RawVideo, error) {
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

func (prs *blackVueDR750X1CHExifParser) getVideoCreationTime(timeString string) (int64, error) {
	t, err := time.Parse("2006:01:02 15:04:05.000", timeString)
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

func (prs *blackVueDR750X1CHExifParser) parseOutGPSMetadata(rawVideo *st.RawVideo) ([]*st.Location, []*st.Motion, []time.Time, error) {
	locations, motions, times, err := getLocationsMotionsTimes("2006:01:02 15:04:05.00Z", prs.unprocessedExifData)
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

	// For some reason it's still off an hour for BlackVue.
	tzOffset -= time.Hour

	// For some reason BlackVue is also always just a bit off in seconds.
	rawVideoStartTime := util.MillisecondsToTime(rawVideo.CreateTimeMs)
	firstLocationTimeTimeZoned := times[0].Add(tzOffset)
	startTimeOffset := rawVideoStartTime.Sub(firstLocationTimeTimeZoned)
	if startTimeOffset > time.Second*30 {
		return nil, nil, nil, fmt.Errorf("RawVideo create time is %v, but first location data time is %v", rawVideoStartTime, firstLocationTimeTimeZoned)
	}

	newTimes := []time.Time{}
	for _, t := range times {
		newTime := t.Add(tzOffset)

		newTime = newTime.Add(startTimeOffset)

		newTimes = append(newTimes, newTime)
	}

	return locations, motions, newTimes, err
}
