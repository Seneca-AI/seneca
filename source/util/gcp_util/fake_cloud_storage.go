package gcp_util

import (
	"fmt"
	"os"
)

// FakeSimpleStorageClient implements a fake SimpleStorageInterface for testing.
type FakeSimpleStorageClient struct {
	// key is bucket, value is files
	files map[string]map[string]bool
}

func NewFakeSimpleStorageClient() *FakeSimpleStorageClient {
	return &FakeSimpleStorageClient{
		files: make(map[string]map[string]bool),
	}
}

func (fssc *FakeSimpleStorageClient) CreateBucket(bucketName string) error {
	fssc.files[bucketName] = make(map[string]bool)
	return nil
}

func (fssc *FakeSimpleStorageClient) BucketExists(bucketName string) (bool, error) {
	_, ok := fssc.files[bucketName]
	if !ok {
		return false, nil
	}
	return true, nil
}

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

func (fssc *FakeSimpleStorageClient) WriteBucketFile(bucketName, localFileNameAndPath, bucketFileName string) error {
	if localFileNameAndPath == "" {
		return fmt.Errorf("received empty localFileName")
	}

	if bucketFileName == "" {
		return fmt.Errorf("received empty bucketFileName")
	}

	f, err := os.Open(localFileNameAndPath)
	if err != nil {
		return fmt.Errorf("error opening local file %q - err: %v", localFileNameAndPath, err)
	}
	defer f.Close()

	bucketExists, err := fssc.BucketExists(bucketName)
	if err != nil {
		return fmt.Errorf("error checking if bucket exists - err: %v", err)
	}
	if !bucketExists {
		return fmt.Errorf("cannot insert file into non-existant bucket %q", bucketName)
	}
	fssc.files[bucketName][bucketFileName] = true
	return nil
}
