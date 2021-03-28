package mp4

import (
	"fmt"
	"io/ioutil"
	"os"
	"seneca/api/senecaerror"
	"strings"
)

// TODO: fix these to be more generic

// CreateTempMP4File creates a temp MP4 file with the given name and returns it.
// Params:
//		userID string: used for returnign NewUserError
//		name string
// Returns:
//		*os.File
//		error
func CreateTempMP4File(userID, name string) (*os.File, error) {
	nameParts := strings.Split(name, ".")
	if len(nameParts) != 2 {
		return nil, senecaerror.NewUserError(userID, fmt.Errorf("error parsing form for mp4 - no file name not in the form (name.mp4)"), "MP4 file not in format (name.mp4).")
	}
	tempName := strings.Join([]string{nameParts[0], "*", nameParts[1]}, ".")

	tempFile, err := ioutil.TempFile("", tempName)
	if err != nil {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("error creating temp file - err: %v", err))
	}

	return tempFile, nil
}

// CreateLocalMP4File creates a a file at the given path with the given name.
// Params:
//		name string
//		path string
// Returns:
//		*os.File
//		error
func CreateLocalMP4File(name, path string) (*os.File, error) {
	mp4File, err := os.Create(fmt.Sprintf("%s/%s", path, name))
	if err != nil {
		return nil, fmt.Errorf("error creating local file - err: %v", err)
	}

	return mp4File, nil
}
