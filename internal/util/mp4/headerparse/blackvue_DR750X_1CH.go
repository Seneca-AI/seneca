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

	if len(times) == 0 {
		return nil, nil, nil, fmt.Errorf("got 0 times")
	}

	diff := util.MillisecondsToTime(rawVideo.CreateTimeMs).Sub(times[0])

	newTimes := []time.Time{}
	for _, t := range times {
		// Adjust timezone to UTC.
		newTime := t.Add(time.Duration(diff.Hours() * -1))
		// TODO(lucaloncar): here, and elsewhere, use the gps coordinates to figure out time zone and adjust
		newTime = t.Add((time.Second * 3) + (time.Hour * -5))

		newTimes = append(newTimes, newTime)
	}

	return locations, motions, newTimes, err
}
