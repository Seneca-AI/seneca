package constants

import "time"

// MaxInputVideoDuration dictates the maximum duration of single videos Seneca will process.
const MaxInputVideoDuration = (time.Minute + time.Second)

// CutVideoDuration dictates the duration of videos after being cut.
const CutVideoDuration = time.Minute

// KilometersToMiles defines the ratio from kilometers to miles.
const KilometersToMiles = float64(1.60934)

// MaxVideoFileSizeMB dictates the maximum size of video files Seneca will accept.
const MaxVideoFileSizeMB int64 = 250

type TableName string

const (
	UsersTable        TableName = "Users"
	RawVideosTable    TableName = "RawVideos"
	RawLocationsTable TableName = "RawLocations"
	RawMotionsTable   TableName = "RawMotions"
)

func (tn TableName) String() string {
	return string(tn)
}
