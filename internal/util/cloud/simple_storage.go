package cloud

// GoogleCloudStorageClient is the interface used for interacting with
// Google Cloud Storage across Seneca.
type SimpleStorageInterface interface {
	CreateBucket(bucketName string) error
	BucketExists(bucketName string) (bool, error)
	BucketFileExists(bucketName, bucketFileName string) (bool, error)
	WriteBucketFile(bucketName, localFileNameAndPath, bucketFileName string) error
}
