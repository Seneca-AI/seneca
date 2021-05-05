package cutter

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/util"
	"seneca/internal/util/data"
	"strings"
	"time"
)

// ffmpeg -i <input file name> -ss <start timestamp> -t <cut duration> copy <output file name>.
const ffmpegCommand = "ffmpeg -i %s -ss %s -t %s -c copy %s"

// 	CutRawVideo utilizes ffmpeg to cut the raw video mp4.
func CutRawVideo(cutVideoDur time.Duration, pathToRawVideo string, rawVideo *st.RawVideo, dryRun bool) ([]*st.CutVideo, []string, error) {
	rawVideoFileName, err := util.GetFileNameFromPath(pathToRawVideo)
	if err != nil {
		return nil, nil, fmt.Errorf("error extracting pathToRawVideo %q - err: %v", pathToRawVideo, err)
	}
	rawVideoFileNameParts := strings.Split(rawVideoFileName, ".")
	if len(rawVideoFileNameParts) < 2 {
		return nil, nil, fmt.Errorf("pathToRawVideo %q in invalid format", pathToRawVideo)
	}
	rawVideoFileNameNoSuffix := strings.Join(rawVideoFileNameParts[:len(rawVideoFileNameParts)-1], ".")
	rawVideoFileNameSuffix := rawVideoFileNameParts[len(rawVideoFileNameParts)-1]

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

	cutVideos := data.ConstructCutVideoData(cutVideoDur, rawVideo)
	// If the rawVideo is already less than or equal to the duration we use for cut videos, then don't bother cutting.
	if rawVideoDuration <= cutVideoDur {
		return cutVideos, []string{pathToRawVideo}, nil
	}

	// Create the temp dir for the CutVideos to be staged.
	tempDirName := ""
	if !dryRun {
		tempDirName, err = ioutil.TempDir("", fmt.Sprintf("%s.CutVideos.*", rawVideo.UserId))
		if err != nil {
			return nil, nil, senecaerror.NewBadStateError(fmt.Errorf("error creating default temp dir with pattern %s.CutVideos.* - err: %v", rawVideo.UserId, err))
		}
	}

	cutVideoFileNames := []string{}
	startTime := time.Duration(0)
	for i, cutVideo := range cutVideos {
		// Add the index to the rawVideoFile name. e.g. user.RAW_VIDEO.123.mp4 => user.RAW_VIDEO.123.0.mp4
		cutVideoFileName := fmt.Sprintf("%s/%s.%d.%s", tempDirName, rawVideoFileNameNoSuffix, i, rawVideoFileNameSuffix)
		cutVideoFileNames = append(cutVideoFileNames, cutVideoFileName)

		// We use a duration even though it actually signifies a timestamp because that's the format for ffmpeg.
		startTimeString := util.DurationToString(startTime)
		cutLengthString := util.DurationToString(util.MillisecondsToDuration(cutVideo.DurationMs))

		commandString := fmt.Sprintf(ffmpegCommand, pathToRawVideo, startTimeString, cutLengthString, cutVideoFileName)
		commandStringParts := strings.Split(commandString, " ")
		if len(commandStringParts) != 10 {
			return nil, nil, senecaerror.NewBadStateError(fmt.Errorf("malformed command string for ffmpeg: %q", commandString))
		}

		if dryRun {
			continue
		}

		// Strangely, a first string arg is required, then the rest can come.
		cmd := exec.Command(commandStringParts[0], commandStringParts[1:]...)

		if err := cmd.Run(); err != nil {
			return nil, nil, fmt.Errorf("error executing command %q - err: %v", commandString, err)
		}

		startTime += util.MillisecondsToDuration(cutVideo.DurationMs)
	}

	return cutVideos, cutVideoFileNames, nil
}
