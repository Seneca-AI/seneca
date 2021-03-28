package mp4

import (
	"fmt"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	"seneca/api/types"
	"seneca/internal/util"
)

const ffmpegCommand = "ffmpeg -i %s -ss "

func CutVideo(pathToRawVideo string, rawVideo *types.RawVideo) ([]*types.CutVideo, []string, error) {
	if rawVideo.UserId == "" {
		return nil, nil, senecaerror.NewBadStateError(fmt.Errorf("rawVideo %v has no userID set", rawVideo))
	}

	if rawVideo.DurationMs == 0 {
		return nil, nil, senecaerror.NewBadStateError(fmt.Errorf("rawVideo %v has no duration set", rawVideo))
	}

	rawVideoDuration := util.MillisecondsToDuration(rawVideo.DurationMs)
	if rawVideoDuration > constants.MaxInputVideoDuration {
		return nil, nil, senecaerror.NewBadStateError(fmt.Errorf("error cutting video, rawVideo.Duration %v is longer than MaxInputVideoDuration %v", rawVideoDuration, constants.MaxInputVideoDuration))
	}

	// If the rawVideo is already less than or equal to the duration we use for cut videos, then don't bother cutting.
	if rawVideoDuration <= constants.CutVideoDuration {

	}

	// Create the temp dir for the CutVideos to be staged.
	// tempDirName, err := ioutil.TempDir("", fmt.Sprintf("%s.CutVideos.*", rawVideo.UserId))
	// if err != nil {
	// 	return nil, nil, senecaerror.NewBadStateError(fmt.Errorf("error creating default temp dir with pattern %s.CutVideos.* - err: %v", rawVideo.UserId, err))
	// }

	// extra := int64(0)
	// if rawVideoDuration.Round(constants.CutVideoDuration) > rawVideoDuration {
	// 	extra = 1
	// }
	// numLoops := int64(rawVideoDuration/constants.CutVideoDuration) + extra
	// for i := int64(0); i < numLoops; i++ {

	// }

	return nil, nil, nil
}
