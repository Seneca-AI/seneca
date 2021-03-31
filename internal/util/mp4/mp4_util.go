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
//		name string
// Returns:
//		*os.File
//		error
func CreateTempMP4File(name string) (*os.File, error) {
	nameParts := strings.Split(name, ".")

	if nameParts[len(nameParts)-1] != "mp4" {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("file name %q does not end in 'mp4'", name))
	}

	tempNameNoSuffix := strings.Join(nameParts[:len(nameParts)-1], ".")
	tempName := strings.Join([]string{tempNameNoSuffix, "*", "mp4"}, ".")

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
