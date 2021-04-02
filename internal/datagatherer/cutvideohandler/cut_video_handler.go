package cutvideohandler

import (
	"fmt"
	"net/http"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	"seneca/internal/util/cloud"
	"seneca/internal/util/data"
	"seneca/internal/util/logging"
	"seneca/internal/util/mp4"
)

const (
	cutVideoBucketFileNameIdentifier = "CUT_VIDEO"
	rawVideoIDPostFormKey            = "raw_video_id"
)

// CutVideoHandler implements all logic for handling cut video requests.
type CutVideoHandler struct {
	simpleStorage cloud.SimpleStorageInterface
	noSQLDB       cloud.NoSQLDatabaseInterface
	mp4Tool       mp4.MP4ToolInterface

	logger    logging.LoggingInterface
	projectID string
}

// NewCutVideoHandler initializes a new CutVideoHandler with the given parameters.
//
// Params:
//		simpleStorageInterface cloud.SimpleStorageInterface: client for storing mp4 files
//		noSQLDatabaseInterface cloud.NoSQLDatabaseInterface: client for storing documents
//		mp4ToolInterface mp4.MP4ToolInterface: tool for parsing and manipulating mp4 data
//		logger logging.LoggingInterface
// 		projectID string
// Returns:
//		*CutVideoHandler: the handler
//		error
func NewCutVideoHandler(simpleStorageInterface cloud.SimpleStorageInterface, noSQLDatabaseInterface cloud.NoSQLDatabaseInterface, mp4ToolInterface mp4.MP4ToolInterface, logger logging.LoggingInterface, projectID string) (*CutVideoHandler, error) {
	return &CutVideoHandler{
		simpleStorage: simpleStorageInterface,
		noSQLDB:       noSQLDatabaseInterface,
		mp4Tool:       mp4ToolInterface,
		logger:        logger,
		projectID:     projectID,
	}, nil
}

// HandleCutVideoPostRequest handles the cut video post request and writes the response.
// Params:
//		w http.ResponseWriter" the response
// 		*http.Request r: the request
// Returns:
//		none
func (cvh *CutVideoHandler) HandleRawVideoPostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		userError := senecaerror.NewUserError("", fmt.Errorf("error handling CutVideoRequest, method %q not supported", r.Method), fmt.Sprintf("Error: %q requests are not supported at this endpoint. Supported methods are: [POST]", r.Method))
		cvh.logger.Log(userError.Error())
		w.WriteHeader(400)
		return
	}

	rawVideoID := r.PostFormValue(rawVideoIDPostFormKey)
	if rawVideoID == "" {
		userError := senecaerror.NewUserError("", fmt.Errorf("error handling CutVideoRequest, no raw_video_id specified"), "No raw_video_id specified in request.")
		cvh.logger.Log(userError.Error())
		w.WriteHeader(400)
		return
	}

	if err := cvh.ProcessAndCutRawVideo(rawVideoID); err != nil {
		cvh.logger.Error(err.Error())
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
}

func (cvh *CutVideoHandler) ProcessAndCutRawVideo(rawVideoID string) error {

	rawVideo, err := cvh.noSQLDB.GetRawVideoByID(rawVideoID)
	if err != nil {
		return fmt.Errorf("error getting raw video by ID - err: %w", err)
	}

	if len(rawVideo.CutVideoId) > 0 {
		cvh.logger.Log(fmt.Sprintf("RawVideo %v already cut.", rawVideo))
		return nil
	}

	// Get raw video from simple storage.
	filePath, err := cvh.simpleStorage.GetBucketFile(cloud.RawVideoBucketName, rawVideo.CloudStorageFileName)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("GetBucketFile(%q, %q) for RawVideo %v returns err: %w", cloud.RawVideoBucketName, rawVideo.CloudStorageFileName, rawVideo, err))
	}

	locations, motions, times, err := cvh.mp4Tool.ParseOutGPSMetadata(filePath)
	if err != nil {
		return fmt.Errorf("ParseOutGPSMetadata(%s) returns err: %w", filePath, err)
	}

	rawLocations, err := data.ConstructRawLocationDatas(rawVideo.UserId, locations, times)
	if err != nil {
		return fmt.Errorf("ConstructRawLocationDatas(%s, %d, %d) returns err: %w", rawVideo.UserId, len(locations), len(times), err)
	}

	rawMotions, err := data.ConstructRawMotionDatas(rawVideo.UserId, motions, times)
	if err != nil {
		return fmt.Errorf("ConstructRawMotionDatas(%s, %d, %d) returns err: %w", rawVideo.UserId, len(motions), len(times), err)
	}

	cutVideos, cutVideoPaths, err := mp4.CutRawVideo(constants.CutVideoDuration, filePath, rawVideo, false /* dryRun */)
	if err != nil {
		return fmt.Errorf("CutRawVideo(%v, %s, %v, false) returns err: %w", rawVideo.UserId, filePath, rawVideo, err)
	}

	cutVideoIDs := []string{}
	rawLocationIDs := []string{}
	rawMotionIDs := []string{}
	for _, cutVideo := range cutVideos {
		cvid, err := cvh.noSQLDB.InsertUniqueCutVideo(cutVideo)
		if err != nil {
			errorString := fmt.Sprintf("InsertUniqueCutVideo(%v) returned err: %v", cutVideo, err)
			cvh.logger.Error(errorString)
			cvh.cleanUpFailure(cutVideoIDs, rawLocationIDs, rawMotionIDs)
			return fmt.Errorf(errorString)
		}
		cutVideoIDs = append(cutVideoIDs, cvid)
	}

	for _, rawLocation := range rawLocations {
		rlid, err := cvh.noSQLDB.InsertUniqueRawLocation(rawLocation)
		if err != nil {
			errorString := fmt.Sprintf("InsertUniqueRawLocation(%v) returned err: %v", rawLocation, err)
			cvh.logger.Error(errorString)
			cvh.cleanUpFailure(cutVideoIDs, rawLocationIDs, rawMotionIDs)
			return fmt.Errorf(errorString)
		}
		rawLocationIDs = append(rawLocationIDs, rlid)
	}

	for _, rawMotion := range rawMotions {
		rmid, err := cvh.noSQLDB.InsertUniqueRawMotion(rawMotion)
		if err != nil {
			errorString := fmt.Sprintf("InsertUniqueRawMotion(%v) returned err: %v", rawMotion, err)
			cvh.logger.Error(errorString)
			cvh.cleanUpFailure(cutVideoIDs, rawMotionIDs, rawMotionIDs)
			return fmt.Errorf(errorString)
		}
		rawMotionIDs = append(rawMotionIDs, rmid)
	}

	for i, cvp := range cutVideoPaths {
		if err := cvh.simpleStorage.WriteBucketFile(cloud.CutVideoBucketName, cvp, fmt.Sprintf("%s.%d.%s.mp4", rawVideo.UserId, cutVideos[i].CreateTimeMs, cutVideoBucketFileNameIdentifier)); err != nil {
			errorString := fmt.Sprintf("WriteBucketFile(%s, %s, %s) returned err: %v", cloud.CutVideoBucketName, cvp, fmt.Sprintf("%s.%d.%s.mp4", rawVideo.UserId, cutVideos[i].CreateTimeMs, cutVideoBucketFileNameIdentifier), err)
			cvh.logger.Error(errorString)
			cvh.cleanUpFailure(cutVideoIDs, rawMotionIDs, rawMotionIDs)
			return fmt.Errorf(errorString)
		}
	}

	cvh.logger.Log(fmt.Sprintf("Successfully processed rawVideo with ID %q for user %q", rawVideo.Id, rawVideo.UserId))
	return nil
}

func (cvh *CutVideoHandler) cleanUpFailure(cutVideoIds []string, rawLocationIds []string, rawMotionsIds []string) {
	for _, cvid := range cutVideoIds {
		if err := cvh.noSQLDB.DeleteCutVideoByID(cvid); err != nil {
			cvh.logger.Error(fmt.Sprintf("failed to clean up CutVideo with ID %q", cvid))
		}
	}

	for _, rlid := range rawLocationIds {
		if err := cvh.noSQLDB.DeleteRawLocationByID(rlid); err != nil {
			cvh.logger.Error(fmt.Sprintf("failed to clean up RawLocation with ID %q", rlid))
		}
	}

	for _, rmid := range rawMotionsIds {
		if err := cvh.noSQLDB.DeleteRawMotionByID(rmid); err != nil {
			cvh.logger.Error(fmt.Sprintf("failed to clean up RawMotion with ID %q", rmid))
		}
	}
}
