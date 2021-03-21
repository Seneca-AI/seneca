package rawvideohandler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"seneca/source/util"
)

const (
	mp4FormKey          = "mp4"
	storeLocally        = false
	fileDumpPath        = "../../file_dump"
	maxFileSizeMB int64 = 250
)

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
func HandleRawVideoPostRequest(w http.ResponseWriter, r *http.Request) {
	var err error
	mp4Buffer, mp4Name, err := getMP4Bytes(r)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		return
	}
	mp4Name = mp4Name + ".mp4"

	var mp4File *os.File
	path := ""
	defer mp4File.Close()
	if storeLocally {
		path = strings.Join([]string{fileDumpPath, mp4Name}, "/")
		mp4File, err = createLocalMP4File(mp4Name, fileDumpPath)
		if err != nil {
			fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		}
	} else {
		mp4File, err = createTempMP4File(mp4Name)
		if err != nil {
			fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		}
		path = mp4File.Name()
	}

	if _, err := mp4File.Write(mp4Buffer); err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		return
	}

	metadata, err := util.GetMetadata(path)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error handling RawVideoRequest - err: %v", err)
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "The video's metadata is %q", metadata.String())
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
