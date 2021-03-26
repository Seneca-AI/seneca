package cloud

// SimpleStorageInterface is the interface used for interacting with
// S3 like files across Seneca.
type SimpleStorageInterface interface {
	CreateBucket(bucketName string) error
	BucketExists(bucketName string) (bool, error)
	BucketFileExists(bucketName, bucketFileName string) (bool, error)
	WriteBucketFile(bucketName, localFileNameAndPath, bucketFileName string) error
}
