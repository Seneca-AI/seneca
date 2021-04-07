package cloud

import (
	"fmt"
)

// FakeSimpleStorageClient implements a fake SimpleStorageInterface for testing.
// No file values are actually stored, simple whether or not they exist.
type FakeSimpleStorageClient struct {
	// key is bucket, value is files
	files                map[BucketName]map[string][]byte
	CreateBucketMock     func(bucketName BucketName) error
	BucketExistsMock     func(bucketName BucketName) (bool, error)
	BucketFileExistsMock func(bucketName BucketName, bucketFileName string) (bool, error)
	WriteBucketFileMock  func(bucketName BucketName, localFileNameAndPath, bucketFileName string) error
	GetBucketFileMock    func(bucketName BucketName, bucketFileName string) (string, error)
}

// NewFakeSimpleStorageClient returns an instance of FakeSimpleStorageClient.
func NewFakeSimpleStorageClient() *FakeSimpleStorageClient {
	return &FakeSimpleStorageClient{
		files: make(map[BucketName]map[string][]byte),
	}
}

// CreateBucket creates a bucket with the given name in the internal bucket and files map.
func (fssc *FakeSimpleStorageClient) CreateBucket(bucketName BucketName) error {
	if fssc.CreateBucketMock == nil {
		return fmt.Errorf("CreateBucketMock not set")
	}
	return fssc.CreateBucketMock(bucketName)
}

// BucketExists checks if a bucket with the given name exists in the internal map.
func (fssc *FakeSimpleStorageClient) BucketExists(bucketName BucketName) (bool, error) {
	if fssc.BucketExistsMock == nil {
		return false, fmt.Errorf("BucketExistsMock not set")
	}
	return fssc.BucketExists(bucketName)
}

// BucketFileExists checks if the given bucket exists, and if it holds the given file.
func (fssc *FakeSimpleStorageClient) BucketFileExists(bucketName BucketName, bucketFileName string) (bool, error) {
	if fssc.BucketFileExistsMock == nil {
		return false, fmt.Errorf("BucketFileExistsMock not set")
	}
	return fssc.BucketFileExistsMock(bucketName, bucketFileName)
}

// WriteBucketFile sets the internal map value for the given bucket at the given bucketFileName to the bytes,
// stored at the file path.
func (fssc *FakeSimpleStorageClient) WriteBucketFile(bucketName BucketName, localFileNameAndPath, bucketFileName string) error {
	if fssc.WriteBucketFileMock == nil {
		return fmt.Errorf("WriteBucketFileMock not set")
	}
	return fssc.WriteBucketFileMock(bucketName, localFileNameAndPath, bucketFileName)
}

// GetBucketFile writes the bytes stored in the map to a temp file and returns the path to the file.
func (fssc *FakeSimpleStorageClient) GetBucketFile(bucketName BucketName, bucketFileName string) (string, error) {
	if fssc.GetBucketFileMock == nil {
		return "", fmt.Errorf("GetBucketFileMock not set")
	}
	return fssc.GetBucketFileMock(bucketName, bucketFileName)
}
