package rawvideohandler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"seneca/api/types"
	"seneca/internal/util"
	"seneca/internal/util/gcp_util"
)

const (
	mp4FormKey          = "mp4"
	maxFileSizeMB int64 = 250
)

// RawVideoHandler implements all logic for handling raw video requests.
type RawVideoHandler struct {
	ssi              gcp_util.SimpleStorageInterface
	nsdi             gcp_util.NoSQLDatabaseInterface
	localStoragePath string
	projectID        string
}

// NewRawVideoHandler initializes a new RawVideoHandler with the given params.
// Params:
//		storageClient *gcp_util.GoogleCloudStorageClient: client for interfacing with GCP storage
// 		localStoragePath string: 	the path where local files will be staged before
// 									upload to Google Cloud Storage.  If empty, temp directories will be used.
// 									If the directory does not exist, requests will fail.
// Returns:
//		*RawVideoHandler: the handler
//		error if localStoragePath does not exist
func NewRawVideoHandler(simpleStorageInterface gcp_util.SimpleStorageInterface, noSQLDatabaseInterface gcp_util.NoSQLDatabaseInterface, localStoragePath, projectID string) (*RawVideoHandler, error) {
	if localStoragePath != "" {
		if _, err := os.Stat(localStoragePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("localStoragePath %q does not exist", localStoragePath)
		}
	}
	return &RawVideoHandler{
		ssi:              simpleStorageInterface,
		nsdi:             noSQLDatabaseInterface,
		localStoragePath: localStoragePath,
		projectID:        projectID,
	}, nil
}

// HandleRawVideoPostRequest accepts POST requests that include mp4 data and
// parses the metadata to gather timestamp info and, if possible, location info.
// The video itself is stored in simple storage and the metadata is stored in
// firestore.  If the video does not contain timestamp info, the server
// returns a 400 error.
// Params:
// 		http.ResponseWriter w: the response to write to
// 		*http.Request r: the request
// Returns:
//		none
func (rvh *RawVideoHandler) HandleRawVideoPostRequest(w http.ResponseWriter, r *http.Request) {
	// Extract request data.
	userID := r.FormValue("user_id")
	if userID == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest, no user_id specified")
		return
	}
	var err error
	mp4Buffer, mp4Name, err := getMP4Bytes(r)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		return
	}
	mp4Name = mp4Name + ".mp4"

	// Stage mp4 and extract metadata..
	var mp4File *os.File
	mp4Path := ""
	defer mp4File.Close()
	if rvh.localStoragePath != "" {
		mp4Path = strings.Join([]string{rvh.localStoragePath, mp4Name}, "/")
		mp4File, err = createLocalMP4File(mp4Name, rvh.localStoragePath)
		if err != nil {
			fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		}
	} else {
		mp4File, err = createTempMP4File(mp4Name)
		if err != nil {
			fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		}
		mp4Path = mp4File.Name()
	}
	if _, err := mp4File.Write(mp4Buffer); err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		return
	}
	metadata, err := util.GetMetadata(mp4Path)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		return
	}

	// Upload firestore data.
	bucketFileName := fmt.Sprintf("%s.%d.RAW_VIDEO.mp4", userID, util.TimeToMilliseconds(metadata.CreationTime))
	rawVideo, err := rvh.writeMP4MetadataToGCD(userID, bucketFileName, metadata)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		return
	}

	// Upload cloud storage data.
	if err := rvh.writeMP4ToGCS(mp4Path, bucketFileName); err != nil {
		if err := rvh.nsdi.DeleteRawVideoByID(rawVideo.Id); err != nil {
			fmt.Printf("Error cleaning up rawVideo with ID %q on failed request\n", rawVideo.Id)
		}

		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "Stored video %q", rawVideo.String())
}

func (rvh *RawVideoHandler) writeMP4MetadataToGCD(userID, bucketFileName string, metadata *util.VideoMetadata) (*types.RawVideo, error) {
	rawVideo := &types.RawVideo{
		UserId:               userID,
		Id:                   "",
		CloudStorageFileName: bucketFileName,
		CreateTimeMs:         util.TimeToMilliseconds(metadata.CreationTime),
	}

	id, err := rvh.nsdi.InsertUniqueRawVideo(rawVideo)
	if err != nil {
		return nil, fmt.Errorf("error inserting MP4 metadata into Google Cloud Datastore - err: %v", err)
	}

	rawVideo.Id = id
	return rawVideo, nil
}

func (rvh *RawVideoHandler) writeMP4ToGCS(mp4Path, bucketFileName string) error {
	if bucketExists, err := rvh.ssi.BucketExists(gcp_util.RawVideoBucketName); err != nil {
		return fmt.Errorf("bucketExists(_, %s, %s) returned err: %v", rvh.projectID, gcp_util.RawVideoBucketName, err)
	} else if !bucketExists {
		if err := rvh.ssi.CreateBucket(gcp_util.RawVideoBucketName); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("created bucket: %q\n", gcp_util.RawVideoBucketName)
	} else {
		fmt.Printf("bucket %q already exists\n", gcp_util.RawVideoBucketName)
	}

	if bucketFileExists, err := rvh.ssi.BucketFileExists(gcp_util.RawVideoBucketName, bucketFileName); !bucketFileExists {
		fmt.Printf("bucketFileExists(_, %s, %s) returns err %v, assuming bucket file does not exist\n", gcp_util.RawVideoBucketName, bucketFileName, err)
		if err := rvh.ssi.WriteBucketFile(gcp_util.RawVideoBucketName, mp4Path, bucketFileName); err != nil {
			return fmt.Errorf("writeBucketFile(%s, %s, %s) returns err: %v", gcp_util.RawVideoBucketName, bucketFileName, mp4Path, err)
		}
	} else {
		fmt.Printf("bucketFile %q already exists, not overwriting\n", bucketFileName)
	}
	return nil
}

func getMP4Bytes(r *http.Request) ([]byte, string, error) {
	maxFileSizeBytes := maxFileSizeMB * 1024 * 1024

	if err := r.ParseMultipartForm(maxFileSizeBytes); err != nil {
		return nil, "", fmt.Errorf("error parsing form for mp4 - err: %v", err)
	}
	var buf bytes.Buffer

	mp4, header, err := r.FormFile(mp4FormKey)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing form for mp4 - err: %v", err)
	}
	defer mp4.Close()

	if header.Size > maxFileSizeBytes {
		return nil, "", fmt.Errorf("error parsing form for mp4 - file too large")
	}

	mp4NameList := strings.Split(header.Filename, ".")
	if len(mp4NameList) < 1 {
		return nil, "", fmt.Errorf("error parsing form for mp4 - no Filename in header")
	}

	io.Copy(&buf, mp4)

	return buf.Bytes(), mp4NameList[0], nil
}

func createTempMP4File(name string) (*os.File, error) {
	nameParts := strings.Split(name, ".")
	if len(nameParts) != 2 {
		return nil, fmt.Errorf("name %q malformed for mp4 file", name)
	}
	tempName := strings.Join([]string{nameParts[0], "*", nameParts[1]}, ".")

	tempFile, err := ioutil.TempFile("", tempName)
	if err != nil {
		return nil, fmt.Errorf("error creating temp file - err: %v", err)
	}

	return tempFile, nil
}

func createLocalMP4File(name, path string) (*os.File, error) {
	mp4File, err := os.Create(fmt.Sprintf("%s/%s", path, name))
	if err != nil {
		return nil, fmt.Errorf("error creating local file - err: %v", err)
	}

	return mp4File, nil
}
