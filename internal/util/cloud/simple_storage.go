package cloud

// SimpleStorageInterface is the interface used for interacting with
// S3 like files across Seneca.
type SimpleStorageInterface interface {
	// CreateBucket creates a bucket in the base project with the given name.
	// Params:
	// 		bucketName string: the name of the bucket
	// Returns:
	//		error
	CreateBucket(bucketName string) error
	// BucketExists checks if a bucket with the given name already exists.
	// Params:
	// 		bucketName string: the name of the bucket
	// Returns:
	//		bool: true if the bucket exists, false otherwise
	//		error
	BucketExists(bucketName string) (bool, error)
	// BucketFileExists checks if a file with the given name exists in the given bucket.
	// Params:
	// 		bucketName string: the name of the bucket
	//		bucketFileName string: the name of the file
	// Returns:
	//		bool: true if the file exists, false otherwise
	// 		error
	BucketFileExists(bucketName, bucketFileName string) (bool, error)
	// WriteBucketFile writes the given local file to the given bucket with the bucketFileName.
	// Params:
	// 		bucketName string: the name of the bucket
	// 		localFileNameAndPath string: path to the local file to upload
	//		bucketFileName string: the name of the file
	// Returns:
	//		error
	WriteBucketFile(bucketName, localFileNameAndPath, bucketFileName string) error
	// GetBucketFile downloads the file with the given bucketFileName, stores it in a temp file, and returns that file's path.
	// Params:
	//		bucketName string
	//		bucketFileName string
	// Returns:
	//		string: the path to the temp file
	//		error
	GetBucketFile(bucketName, bucketFileName string) (string, error)
}
