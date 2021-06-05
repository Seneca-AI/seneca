package cutter

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/util"
	"seneca/internal/util/data"
	"strings"
	"time"
)

const (
	// ffmpeg -i <input file name> -ss <start timestamp> -t <cut duration> copy <output file name>.
	cutVideoCommand = "ffmpeg -i %s -ss %s -t %s -c copy %s"

	// ffmpeg -i <intput file name> -vf fps=<frames per second> <prefix>%05.png.
	videoToFramesCommand = "ffmpeg -i %s -vf fps=%s %s/%%05d.png"
)

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

		commandString := fmt.Sprintf(cutVideoCommand, pathToRawVideo, startTimeString, cutLengthString, cutVideoFileName)
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

// 	RawVideoToFrames converts a rawVideo to constituent frames.
// 	Params:
//
//	Returns:
//		[]*st.RawFrame
//		[]string: ordered filenames
//		error
func RawVideoToFrames(fps float64, pathToRawVideo string, rawVideo *st.RawVideo) ([]*st.RawFrame, string, []string, error) {
	rawVideoFileName, err := util.GetFileNameFromPath(pathToRawVideo)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error extracting pathToRawVideo %q - err: %v", pathToRawVideo, err)
	}

	rawVideoFileNameParts := strings.Split(rawVideoFileName, ".")
	if len(rawVideoFileNameParts) < 2 {
		return nil, "", nil, fmt.Errorf("pathToRawVideo %q in invalid format", pathToRawVideo)
	}

	if rawVideo.Id == "" {
		return nil, "", nil, senecaerror.NewBadStateError(fmt.Errorf("rawVideo %v has no ID set", rawVideo))
	}

	if rawVideo.UserId == "" {
		return nil, "", nil, senecaerror.NewBadStateError(fmt.Errorf("rawVideo %v has no userID set", rawVideo))
	}

	if rawVideo.DurationMs == 0 {
		return nil, "", nil, senecaerror.NewBadStateError(fmt.Errorf("rawVideo %v has no duration set", rawVideo))
	}

	rawFrames := []*st.RawFrame{}
	rawVideoDurationSeconds := util.MillisecondsToDuration(rawVideo.DurationMs).Seconds()
	for i := float64(0); i < float64(rawVideoDurationSeconds)*fps; i++ {
		timestampMS := rawVideo.CreateTimeMs + (time.Second * time.Duration(i*(1/fps))).Milliseconds()
		rf := &st.RawFrame{
			UserId:               rawVideo.UserId,
			TimestampMs:          timestampMS,
			CloudStorageFileName: fmt.Sprintf("%s.%s.%d.png", rawVideo.UserId, rawVideo.Id, timestampMS),
			Source: &st.Source{
				SourceId:   rawVideo.Id,
				SourceType: st.Source_RAW_VIDEO,
			},
		}
		rawFrames = append(rawFrames, rf)
	}

	// Create the temp dir for the CutVideos to be staged.
	tempDirName := ""
	tempDirName, err = ioutil.TempDir("", fmt.Sprintf("RawFrames.%s*", rawVideo.Id))
	if err != nil {
		return nil, "", nil, senecaerror.NewBadStateError(fmt.Errorf("error creating default temp dir with pattern RawFrames.%s* - err: %v", rawVideo.Id, err))
	}

	commandString := fmt.Sprintf(videoToFramesCommand, pathToRawVideo, decimalToFractionString(fps), tempDirName)
	commandStringParts := strings.Split(commandString, " ")
	if len(commandStringParts) != 6 {
		return nil, "", nil, senecaerror.NewBadStateError(fmt.Errorf("malformed command string for ffmpeg: %q", commandString))
	}

	// Strangely, a first string arg is required, then the rest can come.
	cmd := exec.Command(commandStringParts[0], commandStringParts[1:]...)

	if err := cmd.Run(); err != nil {
		return nil, "", nil, fmt.Errorf("error executing command %q - err: %v", commandString, err)
	}

	files, err := os.ReadDir(tempDirName)
	if err != nil {
		os.RemoveAll(tempDirName)
		return nil, "", nil, fmt.Errorf("os.ReadDir(tempDir - %s) returns err: %w", tempDirName, err)
	}

	fileNames := []string{}
	for _, f := range files {
		fileNames = append(fileNames, path.Join(tempDirName, f.Name()))
	}

	fileNamesSorted, err := util.SortStringsAlphaNumerically(fileNames, func(s string) string {
		sParts := strings.Split(s, "/")
		lastPart := sParts[len(sParts)-1]
		leadingZeroesTrimmed := strings.TrimLeft(lastPart, "0")
		return strings.TrimSuffix(leadingZeroesTrimmed, ".png")
	})
	if err != nil {
		os.RemoveAll(tempDirName)
		return nil, "", nil, fmt.Errorf("SortStringsAlphaNumerically() returns err: %w", err)
	}

	return rawFrames, tempDirName, fileNamesSorted, nil
}

func decimalToFractionString(decimal float64) string {
	if decimal >= 1 {
		return fmt.Sprintf("%d", int64(decimal))
	}
	return fmt.Sprintf("1/%d", int(1/decimal))
}
