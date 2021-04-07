package constants

import "time"

// MaxInputVideoDuration dictates the maximum duration of single videos Seneca will process.
const MaxInputVideoDuration = time.Minute * 30

// CutVideoDuration dictates the duration of videos after being cut.
const CutVideoDuration = time.Minute

// KilometersToMiles defines the ratio from kilometers to miles.
const KilometersToMiles = float64(1.60934)

// MaxVideoFileSizeMB dictates the maximum size of video files Seneca will accept.
const MaxVideoFileSizeMB int64 = 250

// PathToOAuthCredentials is the path to the application's credentials for handling oauth.
const PathToOAuthCredentials = ""
