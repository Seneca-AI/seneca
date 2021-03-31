package rawvideohandler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"seneca/api/constants"
	"seneca/api/senecaerror"
	"seneca/api/types"
	"seneca/internal/util"
	"seneca/internal/util/cloud"
	"seneca/internal/util/logging"
	"seneca/internal/util/mp4"
)

const (
	mp4FormKey                             = "mp4"
	rawVideoBucketFileNameIdentifier       = "RAW_VIDEO"
	userIDPostFormKey                      = "user_id"
	maxFileSizeMB                    int64 = 250
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
//		error if localStoragePath does not exist
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
//		w http.ResponseWriter" the response
// 		*http.Request r: the request
// Returns:
//		none
func (rvh *RawVideoHandler) HandleRawVideoPostRequest(w http.ResponseWriter, r *http.Request) {
	if err := rvh.InsertRawVideoFromRequest(r); err != nil {
		senecaerror.WriteErrorToHTTPResponse(w, err)
		return
	}
	w.WriteHeader(200)
}

// HandleRawVideoPostRequest accepts POST requests that include mp4 data and
// parses the metadata.  The metadata is stored in NoSQL DB and the video is stored in simple storage.
// Params:
// 		*http.Request r: the request
// Returns:
//		error
func (rvh *RawVideoHandler) InsertRawVideoFromRequest(r *http.Request) error {
	// Extract request data.
	if r.Method != "POST" {
		userError := senecaerror.NewUserError("", fmt.Errorf("error handling RawVideoRequest, method %q not supported", r.Method), fmt.Sprintf("Error: %q requests are not supported at this endpoint. Supported methods are: [POST]", r.Method))
		rvh.logger.Log(userError.Error())
		return userError
	}

	userID := r.PostFormValue(userIDPostFormKey)
	if userID == "" {
		userError := senecaerror.NewUserError("", fmt.Errorf("error handling RawVideoRequest, no user_id specified"), "No user_id specified in request.")
		rvh.logger.Log(userError.Error())
		return userError
	}
	var err error
	mp4Buffer, mp4Name, err := getMP4BytesFromForm(userID, r)
	if err != nil {
		rvh.logger.Warning(fmt.Sprintf("Error handling RawVideoRequest %v - err: %v", r, err))
		return err
	}
	mp4Name = mp4Name + ".mp4"

	// Stage mp4 locally.
	var mp4File *os.File
	mp4Path := ""
	defer mp4File.Close()
	mp4File, err = mp4.CreateTempMP4File(mp4Name)
	if err != nil {
		rvh.logger.Error(fmt.Sprintf("Error handling RawVideoRequest %v - err: %v", r, err))
		return err
	}
	mp4Path = mp4File.Name()
	if _, err := mp4File.Write(mp4Buffer); err != nil {
		rvh.logger.Warning(fmt.Sprintf("Error handling RawVideoRequest %v - err: %v", r, err))
		return err
	}

	// Extract metadata.
	rawVideo, err := rvh.mp4Tool.ParseOutRawVideoMetadata(mp4Path)
	if err != nil {
		rvh.logger.Warning(fmt.Sprintf("Error handling RawVideoRequest %v - err: %v", r, err))
		return err
	}

	if rawVideo.DurationMs > util.DurationToMilliseconds(constants.MaxInputVideoDuration) {
		userError := senecaerror.NewUserError(userID, fmt.Errorf("error handling RawVideoRequest - duration %v is longer than maxVideoDuration %v", util.MillisecondsToDuration(rawVideo.DurationMs), constants.MaxInputVideoDuration), fmt.Sprintf("Max video duration is %v", constants.MaxInputVideoDuration))
		rvh.logger.Log(userError.Error())
		return err
	}

	// Upload firestore data.
	bucketFileName := fmt.Sprintf("%s.%d.%s.mp4", userID, rawVideo.CreateTimeMs, rawVideoBucketFileNameIdentifier)
	err = rvh.writePartialRawVideoToGCD(userID, bucketFileName, rawVideo)
	if err != nil {
		rvh.logger.Warning(fmt.Sprintf("Error handling RawVideoRequest %v - err: %v", r, err))
		return err
	}

	// Upload cloud storage data.
	if err := rvh.writeMP4ToGCS(mp4Path, bucketFileName); err != nil {
		if err := rvh.noSQLDB.DeleteRawVideoByID(rawVideo.Id); err != nil {
			rvh.logger.Warning(fmt.Sprintf("Error cleaning up rawVideo with ID %q on failed request", rawVideo.Id))
		}
		rvh.logger.Error(fmt.Sprintf("Error handling RawVideoRequest %v - err: %v", r, err))
		return err
	}
	rvh.logger.Log(fmt.Sprintf("Successfully processed video %q for user %q", bucketFileName, userID))
	return nil
}

func (rvh *RawVideoHandler) writePartialRawVideoToGCD(userID, bucketFileName string, rawVideo *types.RawVideo) error {
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
			log.Fatal(err)
		}
		fmt.Printf("created bucket: %q\n", cloud.RawVideoBucketName)
	} else {
		fmt.Printf("bucket %q already exists\n", cloud.RawVideoBucketName)
	}

	if bucketFileExists, err := rvh.simpleStorage.BucketFileExists(cloud.RawVideoBucketName, bucketFileName); !bucketFileExists {
		fmt.Printf("bucketFileExists(_, %s, %s) returns err %v, assuming bucket file does not exist\n", cloud.RawVideoBucketName, bucketFileName, err)
		if err := rvh.simpleStorage.WriteBucketFile(cloud.RawVideoBucketName, mp4Path, bucketFileName); err != nil {
			return fmt.Errorf("writeBucketFile(%s, %s, %s) returns err: %v", cloud.RawVideoBucketName, bucketFileName, mp4Path, err)
		}
	} else {
		fmt.Printf("bucketFile %q already exists, not overwriting\n", bucketFileName)
	}
	return nil
}

func getMP4BytesFromForm(userID string, r *http.Request) ([]byte, string, error) {
	var buf bytes.Buffer
	maxFileSizeBytes := maxFileSizeMB * 1024 * 1024

	mp4, header, err := r.FormFile(mp4FormKey)
	if err != nil {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - err: %v", err), fmt.Sprintf("%q not found in request body.", mp4FormKey))
	}
	defer mp4.Close()

	if header.Size > maxFileSizeBytes {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - file too large"), fmt.Sprintf("File too large. Max file size is %d MB.", maxFileSizeMB))
	}

	if err := r.ParseMultipartForm(maxFileSizeBytes); err != nil {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - err: %v", err), "Malformed mp4 file.")
	}

	mp4NameList := strings.Split(header.Filename, ".")
	if len(mp4NameList) != 2 {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - no file name not in the form (name.mp4)"), "MP4 file not in format (name.mp4).")
	}

	if _, err := io.Copy(&buf, mp4); err != nil {
		return nil, "", senecaerror.NewUserError(userID, fmt.Errorf("error copying bytes - err: %v", err), fmt.Sprintf("Corrupted file: %q.", header.Filename))
	}

	return buf.Bytes(), mp4NameList[0], nil
}
