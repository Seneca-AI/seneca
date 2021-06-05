package gcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"seneca/api/senecaerror"
	"seneca/internal/client/cloud"
	mp4util "seneca/internal/util/mp4/util"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// TODO(lucaloncar): define a client over this service

const (
	// 	QuickTimeOut is the time out used for operations that should be quick,
	// 	like reading metadata or creating a bucket.
	QuickTimeOut = time.Second * 10
	// 	LongTimeOut is the time out used for operations that may take some time,
	// 	like uploading a file.
	LongTimeOut = time.Minute
)

// 	GoogleCloudStorageClient implements SimpleStorageInterface with
// 	Google Cloud Storage.
type GoogleCloudStorageClient struct {
	client       *storage.Client
	projectID    string
	quickTimeOut time.Duration
	longTimeOut  time.Duration
}

// 	NewGoogleCloudStorageClient initializes a new Google storage.Client with the given parameters.
// 	Params:
//		ctx context.Context
// 		projectID string
// 		quickTimeOut time.Duration: the time out used for operations that should be quick, like reading metadata or creating a bucket.
// 		longTimeOut time.Duration: the time out used for operations that may take some time, like uploading a file.
// 	Returns:
//		*GoogleCloudStorageClient
// 		senecaerror.CloudError
func NewGoogleCloudStorageClient(ctx context.Context, projectID string, quickTimeOut, longTimeOut time.Duration) (*GoogleCloudStorageClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, senecaerror.NewCloudError(fmt.Errorf("error initializing NewGoogleCloudStorageClient - err: %v", err))
	}
	return &GoogleCloudStorageClient{
		client:       client,
		projectID:    projectID,
		quickTimeOut: quickTimeOut,
		longTimeOut:  longTimeOut,
	}, nil
}

// 	CreateBucket creates a bucket in the project with the given name.
//	Params:
//		bucketName cloud.BucketName
//	Returns:
//		senecaerror.CloudError
func (gcsc *GoogleCloudStorageClient) CreateBucket(bucketName cloud.BucketName) error {
	ctx, cancel := context.WithTimeout(context.TODO(), gcsc.quickTimeOut)
	defer cancel()
	if err := gcsc.client.Bucket(bucketName.RealName(gcsc.projectID)).Create(ctx, gcsc.projectID, nil); err != nil {
		return senecaerror.NewCloudError(err)
	}
	return nil
}

// 	BucketExists checks if a bucket with the given name already exists.
//	Params:
//		bucketName cloud.BucketName
//	Returns:
//		senecaerror.CloudError
func (gcsc *GoogleCloudStorageClient) BucketExists(bucketName cloud.BucketName) (bool, error) {
	var buckets []string
	ctx, cancel := context.WithTimeout(context.TODO(), gcsc.quickTimeOut)
	defer cancel()
	it := gcsc.client.Buckets(ctx, gcsc.projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return false, senecaerror.NewCloudError(fmt.Errorf("failed to list buckets - err: %v", err))
		}
		buckets = append(buckets, battrs.Name)
	}

	for _, b := range buckets {
		if b == bucketName.RealName(gcsc.projectID) {
			return true, nil
		}
	}
	return false, nil
}

// 	BucketFileExists checks if a file with the given name exists in the given bucket.
// 	This is done by trying to read the attributes of the file.
//	Params:
//		bucketName cloud.BucketName
//		bucketFileName string
//	Returns:
//		bool
//		senecaerror.CloudError, senecaerror.BadStateError
func (gcsc *GoogleCloudStorageClient) BucketFileExists(bucketName cloud.BucketName, bucketFileName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), gcsc.quickTimeOut)
	defer cancel()

	object := gcsc.client.Bucket(bucketName.RealName(gcsc.projectID)).Object(bucketFileName)

	if _, err := object.Attrs(ctx); err != nil {
		dneErr := &storage.ErrObjectNotExist
		bucketDNEError := &storage.ErrBucketNotExist
		if errors.As(err, dneErr) {
			return false, nil
		} else if errors.As(err, bucketDNEError) {
			return false, senecaerror.NewBadStateError(fmt.Errorf("bucket %q does not exist - err: %w", bucketName, err))
		}
		return false, senecaerror.NewCloudError(err)
	}

	return true, nil
}

// 	WriteBucketFile writes the given local file to the given bucket with the bucketFileName.
//	Params:
//		bucketName cloud.BucketName
//		localFileNameAndPath string
//		bucketFileName string
//	Returns:
//		senecaerror.BadStateError
func (gcsc *GoogleCloudStorageClient) WriteBucketFile(bucketName cloud.BucketName, localFileNameAndPath, bucketFileName string) error {
	var err error
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

	ctx, cancel := context.WithTimeout(context.TODO(), gcsc.longTimeOut)
	defer cancel()
	wc := gcsc.client.Bucket(bucketName.RealName(gcsc.projectID)).Object(bucketFileName).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error uploading file: %w", err))
	}
	if err := wc.Close(); err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error closing writer: %w", err))
	}

	// TODO(lucaloncar): remove localFileNameAndPath

	return nil
}

// 	GetBucketFile downloads the file with the given bucketFileName, stores it in a temp file, and returns bytes.
//	Params:
//		bucketName cloud.BucketName
//		bucketFileName string
//	Returns:
//		string: the name of the temporary file downloaded
//		error, senecaerror.CloudError, senecaerror.BadStateError, senecaerror.NotFoundError
func (gcsc *GoogleCloudStorageClient) GetBucketFile(bucketName cloud.BucketName, bucketFileName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), gcsc.longTimeOut)
	defer cancel()

	rc, err := gcsc.client.Bucket(bucketName.RealName(gcsc.projectID)).Object(bucketFileName).NewReader(ctx)
	if err != nil {
		dneErr := &storage.ErrObjectNotExist
		bucketDNEError := &storage.ErrBucketNotExist
		if errors.As(err, dneErr) {
			return "", senecaerror.NewNotFoundError(fmt.Errorf("bucketFile %q does not exist: %w", bucketFileName, err))
		} else if errors.As(err, bucketDNEError) {
			return "", senecaerror.NewBadStateError(fmt.Errorf("bucket %q does not exist - err: %w", bucketName, err))
		}

		return "", senecaerror.NewCloudError(fmt.Errorf("error reading from bucket %q for file %q - err: %w", bucketName, bucketFileName, err))
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("error extracting bytes from file %q in bucket %q - err: %w", bucketFileName, bucketName, err)
	}

	tempFile, err := mp4util.CreateTempMP4File(bucketFileName)
	if err != nil {
		return "", fmt.Errorf("error creating temp file for file %q in bucket %q - err: %w", bucketFileName, bucketName, err)
	}

	if _, err := tempFile.Write(data); err != nil {
		return "", fmt.Errorf("error writing bytes to temp file for file %q in bucket %q - err: %w", bucketFileName, bucketName, err)
	}
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("error closing temp file for file %q in bucket %q - err: %w", bucketFileName, bucketName, err)
	}

	return tempFile.Name(), nil
}

// 	DeleteBucketFile deletes the bucket file from remote storage.
//	Params:
//		bucketName cloud.BucketName
//		bucketFileName string
//	Returns:
//		senecaerror.CloudError, senecaerror.BadStateError, senecaerror.NotFoundError
func (gcsc *GoogleCloudStorageClient) DeleteBucketFile(bucketName cloud.BucketName, bucketFileName string) error {
	if err := gcsc.client.Bucket(bucketName.RealName(gcsc.projectID)).Object(bucketFileName).Delete(context.TODO()); err != nil {
		dneErr := &storage.ErrObjectNotExist
		bucketDNEError := &storage.ErrBucketNotExist
		if errors.As(err, dneErr) {
			return senecaerror.NewNotFoundError(fmt.Errorf("bucketFile %q does not exist: %w", bucketFileName, err))
		} else if errors.As(err, bucketDNEError) {
			return senecaerror.NewBadStateError(fmt.Errorf("bucket %q does not exist - err: %w", bucketName, err))
		}
		return senecaerror.NewCloudError(fmt.Errorf("error deleting file %q in bucket %q: %w", bucketFileName, bucketName, err))
	}
	return nil
}
