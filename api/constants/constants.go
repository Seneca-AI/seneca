package constants

import "time"

// MaxInputVideoDuration dictates the maximum duration of single videos Seneca will process.
const MaxInputVideoDuration = time.Minute * 30

// CutVideoDuration dictates the duration of videos after being cut.
const CutVideoDuration = time.Minute
