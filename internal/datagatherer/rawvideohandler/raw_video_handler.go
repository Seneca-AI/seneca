// !!! listen, response should not include error code. it should just return raw video ID, and error should handle error code. like in our APIT	 !!!

package rawvideohandler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"seneca/api/constants"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/cloud"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
	"seneca/internal/util"
	"seneca/internal/util/data"
	"seneca/internal/util/mp4"
	"seneca/internal/util/mp4/cutter"
	mp4util "seneca/internal/util/mp4/util"
	"strings"
	"time"
)

const (
	mp4FormKey                       = "mp4"
	rawVideoBucketFileNameIdentifier = "RAW_VIDEO"
	userIDPostFormKey                = "user_id"
	// rawVideoExtractedFramesPerSecond * ~7.5 is how long the extraction will take, at least on an M1 Mac.
	rawVideoExtractedFramesPerSecond = 1
)

// RawVideoHandler implements all logic for handling raw video requests.
type RawVideoHandler struct {
	simpleStorage cloud.SimpleStorageInterface
	mp4Tool       mp4.MP4ToolInterface
	logger        logging.LoggingInterface
	projectID     string

	rawVideoDAO    dao.RawVideoDAO
	rawLocationDAO dao.RawLocationDAO
	rawMotionDAO   dao.RawMotionDAO
	rawFrameDAO    dao.RawFrameDAO
}

// NewRawVideoHandler initializes a new RawVideoHandler with the given parameters.
//
// Params:
//		simpleStorageInterface cloud.SimpleStorageInterface: client for storing mp4 files
//		TODO(lucaloncar): fix paramas
//		mp4ToolInterface mp4.MP4ToolInterface: tool for parsing and manipulating mp4 data
//		logger logging.LoggingInterface
// 		projectID string
// Returns:
//		*RawVideoHandler: the handler
//		error
func NewRawVideoHandler(
	simpleStorageInterface cloud.SimpleStorageInterface,
	mp4ToolInterface mp4.MP4ToolInterface,
	rawVideoDAO dao.RawVideoDAO,
	rawLocationDAO dao.RawLocationDAO,
	rawMotionDAO dao.RawMotionDAO,
	rawFrameDAO dao.RawFrameDAO,
	logger logging.LoggingInterface,
	projectID string,
) (*RawVideoHandler, error) {
	return &RawVideoHandler{
		simpleStorage:  simpleStorageInterface,
		mp4Tool:        mp4ToolInterface,
		logger:         logger,
		projectID:      projectID,
		rawVideoDAO:    rawVideoDAO,
		rawLocationDAO: rawLocationDAO,
		rawMotionDAO:   rawMotionDAO,
		rawFrameDAO:    rawFrameDAO,
	}, nil
}

// HandleRawVideoPostRequest handles the raw video post request and writes the response.
// Params:
//		w http.ResponseWriter the response
// 		*http.Request r: the request
// Returns:
//		none
func (rvh *RawVideoHandler) HandleRawVideoHTTPRequest(w http.ResponseWriter, r *http.Request) {
	rawVideoProcessRequest, err := rvh.convertHTTPRequestToRawVideoProcessRequest(r)
	if err != nil {
		senecaerror.WriteErrorToHTTPResponse(w, err)
		return
	}

	_, err = rvh.HandleRawVideoProcessRequest(rawVideoProcessRequest)
	if err != nil {
		senecaerror.WriteErrorToHTTPResponse(w, err)
		logging.LogSenecaError(rvh.logger, err)
	}
	// TODO(lucaloncar): use the RawVideoProcessResponse object
	w.WriteHeader(200)
}

// 	HandleRawVideoProcessRequest implements the logic for handling a RawVideoProcessRequest and returning
// 	a RawVideoProcessResponse
// 	Params:
// 		*st.RawVideoProcessRequest req
//	Returns:
//		*st.RawVideoProcessResponse
func (rvh *RawVideoHandler) HandleRawVideoProcessRequest(req *st.RawVideoProcessRequest) (*st.RawVideoProcessResponse, error) {
	nowTime := time.Now()
	defer func(startTime time.Time) {
		rvh.logger.Log(fmt.Sprintf("Handling RawVideoRequest took %s", time.Since(startTime)))
	}(nowTime)

	cleanUp := newCleanUp(rvh.rawVideoDAO, rvh.rawLocationDAO, rvh.rawMotionDAO, rvh.rawFrameDAO, rvh.logger)
	defer cleanUp.cleanUpFailure()

	if req.UserId == "" {
		return nil, senecaerror.NewUserError("", fmt.Errorf("UserID must not be \"\" in RawVideoProcessRequest"), "UserID not specified in request.")
	}

	// TODO(lucaloncar): check to see if the user actually exists

	rvh.logger.Log(fmt.Sprintf("Handling RawVideoRequest for user %q", req.UserId))

	// Stage mp4 locally.
	mp4Path := ""
	if req.LocalPath == "" {
		var mp4File *os.File
		defer mp4File.Close()
		mp4File, err := mp4util.CreateTempMP4File(req.VideoName)
		if err != nil {
			return nil, senecaerror.NewServerError(fmt.Errorf("error creating temp mp4 file %v - err: %w", req, err))
		}
		mp4Path = mp4File.Name()

		if err := ioutil.WriteFile(mp4File.Name(), req.VideoBytes, 0644); err != nil {
			return nil, senecaerror.NewServerError(fmt.Errorf("error writing mp4 file - err: %w", err))
		}
	} else {
		mp4Path = req.LocalPath
	}

	// Extract metadata.
	rawVideo, locations, motions, times, err := rvh.mp4Tool.ParseVideoMetadata(mp4Path)
	if err != nil {
		var devErr *senecaerror.DevError
		if errors.As(err, &devErr) {
			rvh.logger.Critical(fmt.Sprintf("ParseVideoMetadata(%s) returns DevError: %v", mp4Path, err))
		}

		return nil, fmt.Errorf("mp4Tool.ParseOutRawVideoMetadata(%s) returns - err: %w", mp4Path, err)
	}
	rawVideo.OriginalFileName = req.VideoName
	rawVideo.UserId = req.UserId

	if rawVideo.DurationMs > constants.MaxInputVideoDuration.Milliseconds() {
		return nil, senecaerror.NewUserError(req.UserId, fmt.Errorf("error handling RawVideoProcessRequest - duration %v is longer than maxVideoDuration %v", util.MillisecondsToDuration(rawVideo.DurationMs), constants.MaxInputVideoDuration), fmt.Sprintf("Max video duration is %v", constants.MaxInputVideoDuration))
	}

	// Upload firestore data.
	rawVideo.CloudStorageFileName = fmt.Sprintf("gs://%s/%s", cloud.RawVideoBucketName, fmt.Sprintf("%s.%d.%s.mp4", req.UserId, rawVideo.CreateTimeMs, rawVideoBucketFileNameIdentifier))
	rawVideo, err = rvh.rawVideoDAO.InsertUniqueRawVideo(rawVideo)
	if err != nil {
		cleanUp.clean = true
		return nil, fmt.Errorf("error writing to datastore: %w", err)
	}
	cleanUp.rawVideoID = rawVideo.Id

	// Extract frames.
	rawFrames, framesDirPath, framesFilePaths, err := cutter.RawVideoToFrames(rawVideoExtractedFramesPerSecond, mp4Path, rawVideo)
	defer os.RemoveAll(framesDirPath)
	if err != nil {
		cleanUp.clean = true
		return nil, fmt.Errorf("RawVideoToFrames() returns err: %w", err)
	}

	// Upload cloud storage data.
	if err := rvh.writeMP4ToGCS(mp4Path, rawVideo); err != nil {
		cleanUp.clean = true
		return nil, fmt.Errorf("error writing mp4 to cloud storage: %w", err)
	}

	rawFrames, err = rvh.writeFramesToGCSAndCloudStorageFileNames(rawFrames, framesFilePaths)
	if err != nil {
		cleanUp.clean = true
		return nil, fmt.Errorf("error writing frames to cloud storage: %w", err)
	}

	source := &st.Source{
		SourceId:   rawVideo.Id,
		SourceType: st.Source_RAW_VIDEO,
	}

	rawLocations, err := data.ConstructRawLocationDatas(req.UserId, source, locations, times)
	if err != nil {
		cleanUp.clean = true
		return nil, fmt.Errorf("ConstructRawLocationDatas() returns err: $%w", err)
	}
	rawMotions, err := data.ConstructRawMotionDatas(req.UserId, source, motions, times)
	if err != nil {
		cleanUp.clean = true
		return nil, fmt.Errorf("ConstructRawMotionDatas() returns err: $%w", err)
	}

	if len(rawLocations) != len(rawMotions) {
		cleanUp.clean = true
		return nil, senecaerror.NewBadStateError(fmt.Errorf("len(rawLocations) %d != len(rawMotions) %d", len(rawLocations), len(rawMotions)))
	}
	for i := range rawLocations {
		cleanUp.rawLocationIDs = append(cleanUp.rawLocationIDs, rawLocations[i].Id)
		if _, err := rvh.rawLocationDAO.InsertUniqueRawLocation(rawLocations[i]); err != nil {
			cleanUp.clean = true
			return nil, fmt.Errorf("InsertUniqueRawLocation(%v) returns err: %w", rawLocations[i], err)
		}
		cleanUp.rawMotionIDs = append(cleanUp.rawMotionIDs, rawMotions[i].Id)
		if _, err := rvh.rawMotionDAO.InsertUniqueRawMotion(rawMotions[i]); err != nil {
			cleanUp.clean = true
			return nil, fmt.Errorf("InsertUniqueRawMotion(%v) returns err: %w", rawMotions[i], err)
		}
	}

	for i := range rawFrames {
		cleanUp.rawFrameIDs = append(cleanUp.rawFrameIDs, rawFrames[i].Id)
		if _, err := rvh.rawFrameDAO.InsertUniqueRawFrame(rawFrames[i]); err != nil {
			cleanUp.clean = true
			return nil, fmt.Errorf("InsertUniqueRawFrame(%v) returns err: %w", rawFrames[i], err)
		}
	}

	rvh.logger.Log(fmt.Sprintf("Successfully processed video %q for user %q", rawVideo.CloudStorageFileName, req.UserId))
	return &st.RawVideoProcessResponse{
		RawVideoId: rawVideo.Id,
	}, nil
}

func (rvh *RawVideoHandler) writeMP4ToGCS(mp4Path string, rawVideo *st.RawVideo) error {
	bucketName, fileName, err := data.GCSURLToBucketNameAndFileName(rawVideo.CloudStorageFileName)
	if err != nil {
		return fmt.Errorf("GCSURLToBucketNameAndFileName(%s) returns err: %w", rawVideo.CloudStorageFileName, err)
	}

	if bucketExists, err := rvh.simpleStorage.BucketExists(bucketName); err != nil {
		return fmt.Errorf("bucketExists(_, %s, %s) returned err: %v", rvh.projectID, bucketName, err)
	} else if !bucketExists {
		if err := rvh.simpleStorage.CreateBucket(bucketName); err != nil {
			return fmt.Errorf("CreateBucket(%s) returns err: %w", bucketName, err)
		}
	}

	bucketFileExists, err := rvh.simpleStorage.BucketFileExists(bucketName, fileName)
	if err != nil {
		return fmt.Errorf("error checking if file %q in bucket %q exists: %w", fileName, bucketName, err)
	}
	if !bucketFileExists {
		if err := rvh.simpleStorage.WriteBucketFile(bucketName, mp4Path, fileName); err != nil {
			return fmt.Errorf("writeBucketFile(%s, %s, %s) returns err: %v", bucketName, fileName, mp4Path, err)
		}
	} else {
		return senecaerror.NewBadStateError(fmt.Errorf("attempting to overwrite existing file %q", fileName))
	}

	return nil
}

func (rvh *RawVideoHandler) writeFramesToGCSAndCloudStorageFileNames(rawFrames []*st.RawFrame, localFilePaths []string) ([]*st.RawFrame, error) {
	if bucketExists, err := rvh.simpleStorage.BucketExists(cloud.RawFrameBucketName); err != nil {
		return nil, fmt.Errorf("bucketExists(_, %s, %s) returned err: %v", rvh.projectID, cloud.RawFrameBucketName, err)
	} else if !bucketExists {
		if err := rvh.simpleStorage.CreateBucket(cloud.RawFrameBucketName); err != nil {
			return nil, fmt.Errorf("CreateBucket(%s) returns err: %w", cloud.RawFrameBucketName, err)
		}
	}

	if len(rawFrames) != len(localFilePaths) {
		return nil, fmt.Errorf("have %d rawFrames but %d actual files", len(rawFrames), len(localFilePaths))
	}

	for i, path := range localFilePaths {
		bucketFileExists, err := rvh.simpleStorage.BucketFileExists(cloud.RawFrameBucketName, rawFrames[i].CloudStorageFileName)
		if err != nil {
			return nil, fmt.Errorf("error checking if file %q in bucket %q exists: %w", rawFrames[i].CloudStorageFileName, cloud.RawFrameBucketName, err)
		}
		if !bucketFileExists {
			if err := rvh.simpleStorage.WriteBucketFile(cloud.RawFrameBucketName, path, rawFrames[i].CloudStorageFileName); err != nil {
				return nil, fmt.Errorf("writeBucketFile(%s, %s, %s) returns err: %v", cloud.RawFrameBucketName, path, rawFrames[i].CloudStorageFileName, err)
			}
		} else {
			return nil, senecaerror.NewBadStateError(fmt.Errorf("attempting to overwrite existing file %q", rawFrames[i].CloudStorageFileName))
		}
		rawFrames[i].CloudStorageFileName = fmt.Sprintf("gs://%s/%s", cloud.RawFrameBucketName.RealName(rvh.projectID), rawFrames[i].CloudStorageFileName)
	}

	return rawFrames, nil
}

func (rvh *RawVideoHandler) convertHTTPRequestToRawVideoProcessRequest(r *http.Request) (*st.RawVideoProcessRequest, error) {
	// Extract request data.
	if r.Method != "POST" {
		userError := senecaerror.NewUserError("", fmt.Errorf("error handling RawVideoRequest, method %q not supported", r.Method), fmt.Sprintf("Error: %q requests are not supported at this endpoint. Supported methods are: [POST]", r.Method))
		return nil, userError
	}

	userID := r.PostFormValue(userIDPostFormKey)
	var err error
	mp4Buffer, mp4Name, err := getMP4BytesFromForm(userID, r)
	if err != nil {
		return nil, fmt.Errorf("error getting mp4 bytes from HTTP form body: %w", err)
	}

	return &st.RawVideoProcessRequest{
		UserId:     userID,
		VideoName:  mp4Name,
		VideoBytes: mp4Buffer,
	}, nil
}

func getMP4BytesFromForm(userID string, r *http.Request) ([]byte, string, error) {
	var buf bytes.Buffer
	maxFileSizeBytes := constants.MaxVideoFileSizeMB * 1024 * 1024

	mp4, header, err := r.FormFile(mp4FormKey)
	if err != nil {
		defer mp4.Close()
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - err: %v", err), fmt.Sprintf("%q not found in request body.", mp4FormKey))
	}
	defer mp4.Close()

	if header.Size > maxFileSizeBytes {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - file too large"), fmt.Sprintf("File too large. Max file size is %d MB.", maxFileSizeBytes))
	}
	if err := r.ParseMultipartForm(maxFileSizeBytes); err != nil {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - err: %v", err), "Malformed mp4 file.")
	}
	if header.Filename == "" {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("mp4 file name empty"), "MP4 file name empty.")
	}
	mp4Name, err := util.GetFileNameFromPath(header.Filename)
	if err != nil {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - file name %q not in the form (name.mp4)", header.Filename), "MP4 file not in format (name.mp4).")
	}
	if !strings.HasSuffix(mp4Name, "mp4") {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - file name %q not in the form (name.mp4)", header.Filename), "MP4 file not in format (name.mp4).")
	}
	if _, err := io.Copy(&buf, mp4); err != nil {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error copying bytes - err: %v", err), fmt.Sprintf("Corrupted file: %q.", header.Filename))
	}
	return buf.Bytes(), mp4Name, nil
}

type cleanUp struct {
	rawVideoDAO    dao.RawVideoDAO
	rawLocationDAO dao.RawLocationDAO
	rawMotionDAO   dao.RawMotionDAO
	rawFrameDAO    dao.RawFrameDAO
	logger         logging.LoggingInterface

	clean      bool
	rawVideoID string
	// TODO(lucaloncar): also clean up s3 file so none are dangling
	rawLocationIDs []string
	rawMotionIDs   []string
	rawFrameIDs    []string
}

func newCleanUp(rawVideoDAO dao.RawVideoDAO, rawLocationDAO dao.RawLocationDAO, rawMotionDAO dao.RawMotionDAO, rawFrameDAO dao.RawFrameDAO, logger logging.LoggingInterface) *cleanUp {
	return &cleanUp{
		rawVideoDAO:    rawVideoDAO,
		rawLocationDAO: rawLocationDAO,
		rawMotionDAO:   rawMotionDAO,
		rawFrameDAO:    rawFrameDAO,
		logger:         logger,
	}
}

func (cu *cleanUp) cleanUpFailure() {
	if !cu.clean {
		return
	}

	errs := []error{}

	if cu.rawVideoID != "" {
		errs = append(errs, cu.rawVideoDAO.DeleteRawVideoByID(cu.rawVideoID))
	}

	for _, rlid := range cu.rawLocationIDs {
		errs = append(errs, cu.rawLocationDAO.DeleteRawLocationByID(rlid))
	}

	for _, rmid := range cu.rawMotionIDs {
		errs = append(errs, cu.rawMotionDAO.DeleteRawMotionByID(rmid))
	}

	for _, rfid := range cu.rawFrameIDs {
		errs = append(errs, cu.rawFrameDAO.DeleteRawFrameByID(rfid))
	}

	for _, err := range errs {
		if err != nil {
			cu.logger.Error(fmt.Sprintf("Error cleaning up rawVideo failure: %v", err))
		}
	}

}
