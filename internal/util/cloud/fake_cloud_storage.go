package cloud

import (
	"fmt"
	"os"
	"seneca/api/senecaerror"
)

// FakeSimpleStorageClient implements a fake SimpleStorageInterface for testing.
// No file values are actually stored, simple whether or not they exist.
type FakeSimpleStorageClient struct {
	// key is bucket, value is files
	files map[string]map[string]bool
}

// NewFakeSimpleStorageClient returns an instance of FakeSimpleStorageClient.
func NewFakeSimpleStorageClient() *FakeSimpleStorageClient {
	return &FakeSimpleStorageClient{
		files: make(map[string]map[string]bool),
	}
}

// CreateBucket creates a bucket with the given name in the internal bucket and files map.
func (fssc *FakeSimpleStorageClient) CreateBucket(bucketName string) error {
	fssc.files[bucketName] = make(map[string]bool)
	return nil
}

// BucketExists checks if a bucket with the given name exists in the internal map.
func (fssc *FakeSimpleStorageClient) BucketExists(bucketName string) (bool, error) {
	_, ok := fssc.files[bucketName]
	if !ok {
		return false, nil
	}
	return true, nil
}

// BucketFileExists checks if the given bucket exists, and if it holds the given file.
func (fssc *FakeSimpleStorageClient) BucketFileExists(bucketName, bucketFileName string) (bool, error) {
	bucketMap, ok := fssc.files[bucketName]
	if !ok {
		return false, nil
	}
	_, ok = bucketMap[bucketFileName]
	if !ok {
		return false, nil
	}
	return true, nil
}

// WriteBucketFile sets the internal map value for the given bucket at the given bucketFileName to true,
// and simple checks whether the localFileNameAndPath is not "".
func (fssc *FakeSimpleStorageClient) WriteBucketFile(bucketName, localFileNameAndPath, bucketFileName string) error {
	if localFileNameAndPath == "" {
		return senecaerror.NewBadStateError(fmt.Errorf("received empty localFileName"))
	}

	if bucketFileName == "" {
		return senecaerror.NewBadStateError(fmt.Errorf("received empty bucketFileName"))
	}

	f, err := os.Open(localFileNameAndPath)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error opening local file %q - err: %v", localFileNameAndPath, err))
	}
	defer f.Close()

	bucketExists, err := fssc.BucketExists(bucketName)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error checking if bucket exists - err: %v", err))
	}
	if !bucketExists {
		return senecaerror.NewBadStateError(fmt.Errorf("cannot insert file into non-existant bucket %q", bucketName))
	}
	fssc.files[bucketName][bucketFileName] = true
	return nil
}
