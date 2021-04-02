package cloud

import (
	"fmt"
	"os"
	"seneca/api/senecaerror"
	"seneca/internal/util/mp4"
)

// FakeSimpleStorageClient implements a fake SimpleStorageInterface for testing.
// No file values are actually stored, simple whether or not they exist.
type FakeSimpleStorageClient struct {
	// key is bucket, value is files
	files map[BucketName]map[string][]byte
}

// NewFakeSimpleStorageClient returns an instance of FakeSimpleStorageClient.
func NewFakeSimpleStorageClient() *FakeSimpleStorageClient {
	return &FakeSimpleStorageClient{
		files: make(map[BucketName]map[string][]byte),
	}
}

// CreateBucket creates a bucket with the given name in the internal bucket and files map.
func (fssc *FakeSimpleStorageClient) CreateBucket(bucketName BucketName) error {
	fssc.files[bucketName] = make(map[string][]byte)
	return nil
}

// BucketExists checks if a bucket with the given name exists in the internal map.
func (fssc *FakeSimpleStorageClient) BucketExists(bucketName BucketName) (bool, error) {
	_, ok := fssc.files[bucketName]
	if !ok {
		return false, nil
	}
	return true, nil
}

// BucketFileExists checks if the given bucket exists, and if it holds the given file.
func (fssc *FakeSimpleStorageClient) BucketFileExists(bucketName BucketName, bucketFileName string) (bool, error) {
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

// WriteBucketFile sets the internal map value for the given bucket at the given bucketFileName to the bytes,
// stored at the file path.
func (fssc *FakeSimpleStorageClient) WriteBucketFile(bucketName BucketName, localFileNameAndPath, bucketFileName string) error {
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

	var bytes []byte
	if _, err := f.Read(bytes); err != nil {
		if err == nil {
			return senecaerror.NewBadStateError(fmt.Errorf("error reading bytes from file %q", bucketName))
		}
	}

	fssc.files[bucketName][bucketFileName] = bytes
	return nil
}

// GetBucketFile writes the bytes stored in the map to a temp file and returns the path to the file.
func (fssc *FakeSimpleStorageClient) GetBucketFile(bucketName BucketName, bucketFileName string) (string, error) {
	bucketMap, ok := fssc.files[bucketName]
	if !ok {
		return "", fmt.Errorf("bucket %q does not exist", bucketName)
	}

	data, ok := bucketMap[bucketFileName]
	if !ok {
		return "", fmt.Errorf("file %q in bucket %q does not exist", bucketFileName, bucketName)
	}

	tempFile, err := mp4.CreateTempMP4File(fmt.Sprintf("%s/%s", bucketName, bucketFileName))
	if err != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("error creating temp file for file %q in bucket %q - err: %w", bucketFileName, bucketName, err))
	}

	if _, err := tempFile.Write(data); err != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("error writing bytes to temp file for file %q in bucket %q - err: %w", bucketFileName, bucketName, err))
	}
	if err := tempFile.Close(); err != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("error closing temp file for file %q in bucket %q - err: %w", bucketFileName, bucketName, err))
	}

	return tempFile.Name(), nil
}
