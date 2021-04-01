package cloud

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"seneca/api/senecaerror"
	"seneca/internal/util/mp4"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

const (
	// RawVideoBucketName is the name of the GCS bucket for raw videos.
	RawVideoBucketName = "seneca_raw_videos"
	// CutVideoBucketName is the name of the GCS bucket for cut videos.
	CutVideoBucketName = "seneca_cut_videos"
	// QuickTimeOut is the time out used for operations that should be quick,
	// like reading metadata or creating a bucket.
	QuickTimeOut = time.Second * 10
	// LongTimeOut is the time out used for operations that may take some time,
	// like uploading a file.
	LongTimeOut = time.Minute
)

// GoogleCloudStorageClient implements SimpleStorageInterface with
// Google Cloud Storage.
type GoogleCloudStorageClient struct {
	client       *storage.Client
	projectID    string
	quickTimeOut time.Duration
	longTimeOut  time.Duration
}

// NewGoogleCloudStorageClient initializes a new Google storage.Client with the given parameters.
// Params:
// 		ctx context.Context
// 		projectID string: the project
// 		quickTimeOut time.Duration: the time out used for operations that should be quick, like reading metadata or creating a bucket.
// 		longTimeOut time.Duration: the time out used for operations that may take some time, like uploading a file.
// Returns:
//		*GoogleCloudStorageClient: the client
// 		error
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

// CreateBucket creates a bucket in the project with the given name.
func (gcsc *GoogleCloudStorageClient) CreateBucket(bucketName string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, gcsc.quickTimeOut)
	defer cancel()
	if err := gcsc.client.Bucket(bucketName).Create(ctx, gcsc.projectID, nil); err != nil {
		return senecaerror.NewCloudError(err)
	}
	return nil
}

// BucketExists checks if a bucket with the given name already exists.
func (gcsc *GoogleCloudStorageClient) BucketExists(bucketName string) (bool, error) {
	ctx := context.Background()

	var buckets []string
	ctx, cancel := context.WithTimeout(ctx, gcsc.quickTimeOut)
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
		if b == bucketName {
			return true, nil
		}
	}
	return false, nil
}

// BucketFileExists checks if a file with the given name exists in the given bucket.
// This is done by trying to read the attributes of the file, and if an error is
// returned, we assume the file does not exist.
func (gcsc *GoogleCloudStorageClient) BucketFileExists(bucketName, bucketFileName string) (bool, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, gcsc.quickTimeOut)
	defer cancel()

	object := gcsc.client.Bucket(bucketName).Object(bucketFileName)

	// If there is an error, we assume the file does not exist.
	if _, err := object.Attrs(ctx); err != nil {
		return false, senecaerror.NewCloudError(err)
	}

	return true, nil
}

// WriteBucketFile writes the given local file to the given bucket with the bucketFileName.
func (gcsc *GoogleCloudStorageClient) WriteBucketFile(bucketName, localFileNameAndPath, bucketFileName string) error {
	var err error
	if localFileNameAndPath == "" {
		return senecaerror.NewBadStateError(fmt.Errorf("received empty localFileName"))
	}

	if bucketFileName == "" {
		return senecaerror.NewBadStateError(fmt.Errorf("received empty bucketFileName"))
	}

	ctx := context.Background()
	f, err := os.Open(localFileNameAndPath)
	if err != nil {
		return senecaerror.NewBadStateError(fmt.Errorf("error opening local file %q - err: %v", localFileNameAndPath, err))
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, gcsc.longTimeOut)
	defer cancel()
	wc := gcsc.client.Bucket(bucketName).Object(bucketFileName).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return senecaerror.NewBadStateError(err)
	}
	if err := wc.Close(); err != nil {
		return senecaerror.NewBadStateError(err)
	}
	return nil
}

// GetBucketFile downloads the file with the given bucketFileName, stores it in a temp file, and returns bytes.
func (gcsc *GoogleCloudStorageClient) GetBucketFile(bucketName, bucketFileName string) (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, gcsc.longTimeOut)
	defer cancel()

	rc, err := gcsc.client.Bucket(bucketName).Object(bucketFileName).NewReader(ctx)
	if err != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("error reading from bucket %q for file %q - err: %w", bucketName, bucketFileName, err))
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", senecaerror.NewBadStateError(fmt.Errorf("error extracting bytes from file %q in bucket %q - err: %w", bucketFileName, bucketName, err))
	}

	tempFile, err := mp4.CreateTempMP4File(bucketFileName)
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
