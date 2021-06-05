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

// HeartbeatEndpoint defines the HTTP endpoint that the server will answer heartbeat requests on.
const HeartbeatEndpoint = "heartbeat"

// SenecaAPIKey defines the key HTTP requests must have in their header to be accepted by all Seneca servers.  This will soon be replaced by reasonable
// auth methods.
const SenecaAPIKey = "lSfsjS3nebraYqbzbpFS"

type TableName string

const (
	UsersTable            TableName = "Users"
	RawVideosTable        TableName = "RawVideos"
	RawLocationsTable     TableName = "RawLocations"
	RawMotionsTable       TableName = "RawMotions"
	RawFramesTable        TableName = "RawFrames"
	EventTable            TableName = "Events"
	DrivingConditionTable TableName = "DrivingConditions"
	TripTable             TableName = "Trips"
)

func (tn TableName) String() string {
	return string(tn)
}

var DataTableNames = []TableName{UsersTable, RawVideosTable, RawLocationsTable, RawMotionsTable, EventTable, DrivingConditionTable, TripTable, RawFramesTable}

type SenecaTypeFieldName string

const (
	UserIDFieldName       SenecaTypeFieldName = "UserId"
	CreateTimeFieldName   SenecaTypeFieldName = "CreateTimeMs"
	TimestampFieldName    SenecaTypeFieldName = "TimestampMs"
	EmailFieldName        SenecaTypeFieldName = "Email"
	StartTimeFieldName    SenecaTypeFieldName = "StartTimeMs"
	EndTimeFieldName      SenecaTypeFieldName = "EndTimeMs"
	TripIDFieldName       SenecaTypeFieldName = "TripId"
	AlgosVersionFieldName SenecaTypeFieldName = "AlgosVersion"
)

func (stfn SenecaTypeFieldName) String() string {
	return string(stfn)
}
