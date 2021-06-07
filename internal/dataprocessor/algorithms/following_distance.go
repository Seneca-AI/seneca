package algorithms

import (
	"fmt"
	"math"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/intraseneca"
	"seneca/internal/dataprocessor"
	"seneca/internal/util"
	"sort"
	"time"
)

type followingDistanceV0 struct {
	intraSenecaClient         intraseneca.IntraSenecaInterface
	widthBuckets              util.RangeMap
	speedBucketSize           time.Duration
	minXScreenBoundConsidered float64
	maxXScreenBoundConsidered float64
	minConfidenceLevel        float64
}

func newFollowingDistanceV0(intraSenecaClient intraseneca.IntraSenecaInterface) (*followingDistanceV0, error) {
	// Width * 100
	keys := []util.Range{
		// Too far to care.
		{
			L: 0,
			U: 4,
		},
		{
			L: 4,
			U: 5,
		},
		{
			L: 5,
			U: 6,
		},
		{
			L: 6,
			U: 7,
		},
		{
			L: 7,
			U: 8,
		},
		{
			L: 8,
			U: 10,
		},
		{
			L: 10,
			U: 12,
		},
		// Ignore, it's probably not a car.
		{
			L: 12,
			U: 40,
		},
	}

	// Min speed.
	values := []interface{}{9999.0, 90.0, 80.0, 70.0, 60.0, 50.0, 35.0, 9999.0}

	widthBuckets, err := util.NewRangeMap(keys, values)
	if err != nil {
		return nil, fmt.Errorf("NewRangeMap() returns err: %w", err)
	}

	return &followingDistanceV0{
		intraSenecaClient:         intraSenecaClient,
		widthBuckets:              *widthBuckets,
		speedBucketSize:           time.Second,
		minXScreenBoundConsidered: 0.3,
		maxXScreenBoundConsidered: 0.7,
		minConfidenceLevel:        0.5,
	}, nil
}

type speedEntry struct {
	total      float64
	numEntries int
	sourceIDs  []string
}

type timestampedSeverity struct {
	timestamp int64
	severity  float64
	sourceID  string
}

func (fd *followingDistanceV0) GenerateDrivingConditions(inputs map[string][]interface{}) ([]*st.DrivingConditionInternal, error) {
	motionObjs := inputs[dataprocessor.RawMotionTypeString]
	if len(motionObjs) == 0 {
		return nil, nil
	}

	frameObjs := inputs[dataprocessor.RawFrameTypeString]
	if len(frameObjs) == 0 {
		return nil, nil
	}

	// Construct time buckets with their average speeds.
	speeds, err := util.NewRangeMap([]util.Range{}, []interface{}{})
	if err != nil {
		return nil, senecaerror.NewDevError(fmt.Errorf("NewRangeMap(nil, nil) returns err: %w", err))
	}
	userID := ""
	for _, mObj := range motionObjs {
		motion, ok := mObj.(*st.RawMotion)
		if !ok {
			return nil, senecaerror.NewDevError(fmt.Errorf("found a %T in map entry for %s", mObj, dataprocessor.RawMotionTypeString))
		}
		userID = motion.UserId

		secondsFloor := motion.TimestampMs - (motion.TimestampMs % time.Second.Milliseconds())
		secondsCeiling := secondsFloor + time.Second.Milliseconds()

		keyRange := util.Range{
			L: secondsFloor,
			U: secondsCeiling,
		}

		speedEntryObj, ok := speeds.Get(motion.TimestampMs)
		if !ok {
			speedEntryObj = speedEntry{}
		}

		speedEntryVal := speedEntryObj.(speedEntry)

		speedEntryVal.numEntries++
		speedEntryVal.total += motion.Motion.VelocityMph
		speedEntryVal.sourceIDs = append(speedEntryVal.sourceIDs, motion.Id)

		speeds.Insert(keyRange, speedEntryVal)
	}

	followingDistanceEntries := []timestampedSeverity{}
	for _, frameObj := range frameObjs {
		frame, ok := frameObj.(*st.RawFrame)
		if !ok {
			return nil, senecaerror.NewDevError(fmt.Errorf("found a %T in map entry for %s", frameObj, dataprocessor.RawFrameTypeString))
		}

		speedEntryAtFrameObj, ok := speeds.Get(frame.TimestampMs)
		if !ok {
			// TODO(lucaloncar): log this somehow
			continue
		}

		speedEntryAtFrame := speedEntryAtFrameObj.(speedEntry)

		severity := fd.followingDistanceSeverity((speedEntryAtFrame.total / float64(speedEntryAtFrame.numEntries)), frame)

		if severity > 0 {
			followingDistanceEntries = append(followingDistanceEntries, timestampedSeverity{timestamp: frame.TimestampMs, severity: severity, sourceID: frame.Id})
		}
	}

	sort.Slice(followingDistanceEntries, func(i, j int) bool {
		return followingDistanceEntries[i].timestamp < followingDistanceEntries[j].timestamp
	})

	drivingConditions := []*st.DrivingConditionInternal{}

	totalSeverity := float64(0)
	numValues := 0
	firstSourceID := ""
	firstTimeStamp := util.TimeToMilliseconds(time.Now())
	lastTimestamp := util.TimeToMilliseconds(time.Now())
	for i, fde := range followingDistanceEntries {
		if numValues == 0 {
			firstSourceID = fde.sourceID
			firstTimeStamp = fde.timestamp
			lastTimestamp = fde.timestamp
		}

		numValues++
		totalSeverity += fde.severity

		if util.MillisecondsToDuration(fde.timestamp-lastTimestamp) > time.Second*5 || i == len(followingDistanceEntries)-1 {
			dc := &st.DrivingConditionInternal{
				UserId:        userID,
				ConditionType: st.ConditionType_CLOSE_FOLLOWING_DISTANCE,
				Severity:      totalSeverity / float64(numValues),
				StartTimeMs:   firstTimeStamp,
				EndTimeMs:     lastTimestamp,
				Source: &st.Source{
					SourceType: st.Source_RAW_FRAME,
					SourceId:   firstSourceID,
				},
				AlgoTag: fd.Tag(),
			}
			drivingConditions = append(drivingConditions, dc)
			totalSeverity = float64(0)
			numValues = 0
			firstSourceID = ""
		}

		lastTimestamp = fde.timestamp
	}

	return drivingConditions, nil
}

func (fd *followingDistanceV0) followingDistanceSeverity(speed float64, frame *st.RawFrame) float64 {
	objectsInFrame, err := fd.intraSenecaClient.ProcessObjectsInVideo(&st.ObjectsInFrameRequest{RawFrame: frame})
	// TODO(lucaloncar): log failure
	if err != nil {
		return 0
	}

	if objectsInFrame == nil {
		return 0
	}

	potentialCandidates := []*st.ObjectBox{}
	// Filter out non-candidates.
	for _, objBox := range objectsInFrame.ObjectInFrame.ObjectBox {

		if isOutOfBounds(fd.minXScreenBoundConsidered, fd.maxXScreenBoundConsidered, objBox) {
			continue
		}

		if isLowConfidence(fd.minConfidenceLevel, objBox) {
			continue
		}

		if isNotCarOrTruck(objBox) {
			continue
		}

		potentialCandidates = append(potentialCandidates, objBox)
	}

	mostLikely := findMostLikelyCandidate(potentialCandidates)

	if mostLikely == nil {
		return 0
	}

	sizeOneHundred := int64((mostLikely.XUpper - mostLikely.XLower) * 100)
	minSpeedToViolateObj, ok := fd.widthBuckets.Get(sizeOneHundred)
	// TODO(lucaloncar): this should never happen, consider logging
	if !ok {
		return 0
	}

	minSpeedToViolate, ok := minSpeedToViolateObj.(float64)

	// TODO(lucaloncar): this should never happen, consider logging
	if !ok {
		return 0
	}

	if speed > minSpeedToViolate {
		return 1
	}

	return 0
}

func findMostLikelyCandidate(potentialCandidates []*st.ObjectBox) *st.ObjectBox {
	var mostLikely *st.ObjectBox

	for _, cand := range potentialCandidates {
		if mostLikely == nil {
			mostLikely = cand
			continue
		}

		if isMoreLikely(cand, mostLikely) {
			mostLikely = cand
		}
	}

	return mostLikely
}

func isMoreLikely(lhs *st.ObjectBox, rhs *st.ObjectBox) bool {
	lhsScore := 0.0
	rhsScore := 0.0

	// Which is smaller? With 0.5 of a weight.
	lhsBoxSize := boxSize(lhs)
	rhsBoxSize := boxSize(rhs)
	if lhsBoxSize < rhsBoxSize {
		lhsScore += 0.5
	} else {
		rhsScore += 0.5
	}

	// Which is more central? With 0.25 of a weight.
	lhsCentralScore, rhsCentralScore := centralityScore(lhs, rhs)
	lhsScore += (lhsCentralScore * 0.25)
	rhsCentralScore += (rhsCentralScore * 0.25)

	// Which is more of a square? With a 0.25 weight.
	lhsSquarnessScore, rhsSquarenessScore := squareScore(lhs, rhs)
	lhsScore += (lhsSquarnessScore * 0.25)
	rhsCentralScore += (rhsSquarenessScore * 0.25)

	return lhsScore > rhsScore
}

func squareScore(lhs *st.ObjectBox, rhs *st.ObjectBox) (float64, float64) {
	lhsRatio := (lhs.YUpper - lhs.YLower) / (lhs.XUpper - lhs.XLower)
	rhsRatio := (rhs.YUpper - rhs.YLower) / (rhs.XUpper - rhs.XLower)

	maxRatio := math.Max(lhsRatio, lhsRatio)

	return (1 - (lhsRatio / maxRatio)), (1 - (rhsRatio / maxRatio))
}

func centralityScore(lhs *st.ObjectBox, rhs *st.ObjectBox) (float64, float64) {
	lhsCentralDistance := math.Abs((lhs.XUpper - lhs.XLower) - 0.5)
	rhsCentralDistance := math.Abs((rhs.XUpper - rhs.XLower) - 0.5)

	maxDistance := math.Max(lhsCentralDistance, rhsCentralDistance)

	return (1 - (lhsCentralDistance / maxDistance)), (1 - (rhsCentralDistance / maxDistance))
}

func boxSize(objBox *st.ObjectBox) float64 {
	return (100 * (objBox.YUpper - objBox.YLower)) * (100 * (objBox.XUpper - objBox.XLower))
}

func isNotCarOrTruck(objBox *st.ObjectBox) bool {
	return objBox.ObjectLabel != st.ObjectBox_CAR && objBox.ObjectLabel != st.ObjectBox_TRUCK
}

func isLowConfidence(minConfidenceLevel float64, objBox *st.ObjectBox) bool {
	return objBox.Confidence < minConfidenceLevel
}

func isOutOfBounds(lowerLimit, upperLimit float64, objBox *st.ObjectBox) bool {
	return objBox.XLower < lowerLimit || objBox.XUpper > upperLimit
}

func (fd *followingDistanceV0) GenerateEvents(inputs map[string][]interface{}) ([]*st.EventInternal, error) {
	return nil, nil
}

func (fd *followingDistanceV0) Tag() string {
	return "00004"
}
