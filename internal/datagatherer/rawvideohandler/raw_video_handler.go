// !!! listen, response should not include error code. it should just return raw video ID, and error should handle error code. like in our APIT	 !!!

package rawvideohandler

import (
	"bytes"
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
	"seneca/internal/util"
	"seneca/internal/util/mp4"
	"strings"
)

const (
	mp4FormKey                       = "mp4"
	rawVideoBucketFileNameIdentifier = "RAW_VIDEO"
	userIDPostFormKey                = "user_id"
)

// RawVideoHandler implements all logic for handling raw video requests.
type RawVideoHandler struct {
	simpleStorage cloud.SimpleStorageInterface
	noSQLDB       cloud.NoSQLDatabaseInterface
	mp4Tool       mp4.MP4ToolInterface
	logger        logging.LoggingInterface
	projectID     string
}

// NewRawVideoHandler initializes a new RawVideoHandler with the given parameters.
//
// Params:
//		simpleStorageInterface cloud.SimpleStorageInterface: client for storing mp4 files
//		noSQLDatabaseInterface cloud.NoSQLDatabaseInterface: client for storing documents
//		mp4ToolInterface mp4.MP4ToolInterface: tool for parsing and manipulating mp4 data
//		logger logging.LoggingInterface
// 		projectID string
// Returns:
//		*RawVideoHandler: the handler
//		error
func NewRawVideoHandler(simpleStorageInterface cloud.SimpleStorageInterface, noSQLDatabaseInterface cloud.NoSQLDatabaseInterface, mp4ToolInterface mp4.MP4ToolInterface, logger logging.LoggingInterface, projectID string) (*RawVideoHandler, error) {
	return &RawVideoHandler{
		simpleStorage: simpleStorageInterface,
		noSQLDB:       noSQLDatabaseInterface,
		mp4Tool:       mp4ToolInterface,
		logger:        logger,
		projectID:     projectID,
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

	_, err = rvh.InsertRawVideoFromRequest(rawVideoProcessRequest)
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
func (rvh *RawVideoHandler) InsertRawVideoFromRequest(req *st.RawVideoProcessRequest) (*st.RawVideoProcessResponse, error) {
	if req.UserId == "" {
		return nil, senecaerror.NewUserError("", fmt.Errorf("UserID must not be \"\" in RawVideoProcessRequest"), fmt.Sprintf("UserID not specified in request."))
	}

	// Stage mp4 locally.
	var mp4File *os.File
	mp4Path := ""
	defer mp4File.Close()
	mp4File, err := mp4.CreateTempMP4File(req.VideoName)
	if err != nil {
		return nil, senecaerror.NewServerError(fmt.Errorf("error creating temp mp4 file %v - err: %w", req, err))
	}
	mp4Path = mp4File.Name()

	if err := ioutil.WriteFile(mp4File.Name(), req.VideoBytes, 0644); err != nil {
		return nil, senecaerror.NewServerError(fmt.Errorf("error writing mp4 file - err: %w", err))
	}

	// Extract metadata.
	rawVideo, err := rvh.mp4Tool.ParseOutRawVideoMetadata(mp4Path)
	if err != nil {
		userError := senecaerror.NewUserError(req.UserId, fmt.Errorf("mp4Tool.ParseOutRawVideoMetadata(%s) returns - err: %w", mp4Path, err), "Error handling RawVideoProcessRequest.  MP4 may not have headers.")
		return nil, userError
	}

	if rawVideo.DurationMs > constants.MaxInputVideoDuration.Milliseconds() {
		return nil, senecaerror.NewUserError(req.UserId, fmt.Errorf("error handling RawVideoProcessRequest - duration %v is longer than maxVideoDuration %v", util.MillisecondsToDuration(rawVideo.DurationMs), constants.MaxInputVideoDuration), fmt.Sprintf("Max video duration is %v", constants.MaxInputVideoDuration))
	}

	// Upload firestore data.
	bucketFileName := fmt.Sprintf("%s.%d.%s.mp4", req.UserId, rawVideo.CreateTimeMs, rawVideoBucketFileNameIdentifier)
	err = rvh.writePartialRawVideoToGCD(req.UserId, bucketFileName, rawVideo)
	if err != nil {
		return nil, fmt.Errorf("error writing to datastore: %w", err)
	}

	// Upload cloud storage data.
	if err := rvh.writeMP4ToGCS(mp4Path, bucketFileName); err != nil {
		if err := rvh.noSQLDB.DeleteRawVideoByID(rawVideo.Id); err != nil {
			rvh.logger.Warning(fmt.Sprintf("Error cleaning up rawVideo with ID %q on failed request", rawVideo.Id))
		}
		return nil, fmt.Errorf("error writing mp4 to cloud storage: %w", err)
	}

	rvh.logger.Log(fmt.Sprintf("Successfully processed video %q for user %q", bucketFileName, req.UserId))
	return &st.RawVideoProcessResponse{
		RawVideoId: rawVideo.Id,
	}, nil
}

func (rvh *RawVideoHandler) writePartialRawVideoToGCD(userID, bucketFileName string, rawVideo *st.RawVideo) error {
	rawVideo.UserId = userID
	rawVideo.CloudStorageFileName = bucketFileName

	id, err := rvh.noSQLDB.InsertUniqueRawVideo(rawVideo)
	if err != nil {
		return fmt.Errorf("error inserting MP4 metadata into Google Cloud Datastore - err: %w", err)
	}

	rawVideo.Id = id
	return nil
}

func (rvh *RawVideoHandler) writeMP4ToGCS(mp4Path, bucketFileName string) error {
	if bucketExists, err := rvh.simpleStorage.BucketExists(cloud.RawVideoBucketName); err != nil {
		return fmt.Errorf("bucketExists(_, %s, %s) returned err: %v", rvh.projectID, cloud.RawVideoBucketName, err)
	} else if !bucketExists {
		if err := rvh.simpleStorage.CreateBucket(cloud.RawVideoBucketName); err != nil {
			return fmt.Errorf("CreateBucket(%s) returns err: %w", cloud.RawVideoBucketName, err)
		}
	}

	if bucketFileExists, err := rvh.simpleStorage.BucketFileExists(cloud.RawVideoBucketName, bucketFileName); !bucketFileExists {
		fmt.Printf("bucketFileExists(_, %s, %s) returns err %v, assuming bucket file does not exist\n", cloud.RawVideoBucketName, bucketFileName, err)
		if err := rvh.simpleStorage.WriteBucketFile(cloud.RawVideoBucketName, mp4Path, bucketFileName); err != nil {
			return fmt.Errorf("writeBucketFile(%s, %s, %s) returns err: %v", cloud.RawVideoBucketName, bucketFileName, mp4Path, err)
		}
	}

	return nil
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
