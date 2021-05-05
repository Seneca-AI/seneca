package googledrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"seneca/api/constants"
	st "seneca/api/type"
	mp4util "seneca/internal/util/mp4/util"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

const (
	googleDriveFolderName = "Senecacam"
	fileSuccessPrefix     = "SUCCESS_"
	fileErrorPrefix       = "ERROR_"
)

//nolint
type GoogleDriveUserInterface interface {
	// 	ListFileIDs lists all of the relevant files from the user's Google Drive.
	ListFileIDs() ([]string, error)
	// 	DownloadFileByID downloads the file with the given ID and returns a path to the tmp file saved to disk.
	DownloadFileByID(fileID string) (string, error)
	// 	MarkFileByID marks the file with the given ID as a success or failure.
	MarkFileByID(fileID string, failure bool) error
}

//nolint
type GoogleDriveUserClient struct {
	user     *st.User
	service  *drive.Service
	folderID string
}

// 	NewGoogleDriveUserClient returns a new GoogleDriveUserClient using the given path to oauth credentials.
//	Params:
//		user *st.User: the user to create this client for
//		pathToOAuthCredentials string: path to oauth credentials file
//	Returns:
//		*GoogleDriveUserClient
//		error
func NewGoogleDriveUserClient(user *st.User, pathToOAuthCredentials string) (*GoogleDriveUserClient, error) {
	b, err := ioutil.ReadFile(pathToOAuthCredentials)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadFile(%s) returns err: %w", pathToOAuthCredentials, err)
	}

	oauthConfig, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return nil, fmt.Errorf("google.ConfigFromJSON(%s, drive.DriveMetadataScope) returns err: %w", pathToOAuthCredentials, err)
	}

	if len(user.OauthToken) == 0 {
		return nil, fmt.Errorf("user with ID %q does not have OauthToken set", user.Id)
	}

	tok := &oauth2.Token{}
	err = json.NewDecoder(bytes.NewReader(user.OauthToken)).Decode(tok)
	if err != nil {
		return nil, fmt.Errorf("error reading oauth token for user with ID %s - err: %w", user.Id, err)
	}

	client := oauthConfig.Client(context.TODO(), tok)
	service, err := drive.New(client)
	if err != nil {
		return nil, fmt.Errorf("error initializing Drive Client for user with ID %s - err: %w", user.Id, err)
	}

	queryString := fmt.Sprintf("title = '%s'", googleDriveFolderName)
	r, err := service.Files.List().Q(queryString).Do()
	if err != nil {
		return nil, fmt.Errorf("service.Files.List().Q(%s).Do() returns err: %w", queryString, err)
	}
	if len(r.Items) != 1 {
		return nil, fmt.Errorf("got %d files when searching for %q , want 1", len(r.Items), queryString)
	}

	return &GoogleDriveUserClient{
		user:     user,
		service:  service,
		folderID: r.Items[0].Id,
	}, nil
}

// 	ListFileIDs returns a list of the IDs of all of the files in the Senecacam googleDriveFolderName
// 	without the fileSuccessPrefix and fileErrorPrefix but with a ".mp4" suffix.
// 	Params:
// 	Returns:
//		[]string: a list of the file IDs
//		error
func (gduc *GoogleDriveUserClient) ListFileIDs() ([]string, error) {
	queryString := fmt.Sprintf("parents in '%s' and not title contains '%s' and not title contains '%s' and title contains '.mp4'", gduc.folderID, fileSuccessPrefix, fileErrorPrefix)
	var fileIDs []string
	pageToken := ""
	for {
		q := gduc.service.Files.List().Q(queryString)
		// If we have a pageToken set, apply it to the query
		if pageToken != "" {
			q = q.PageToken(pageToken)
		}
		r, err := q.Do()
		if err != nil {
			fmt.Printf("An error occurred: %v\n", err)
			return nil, err
		}

		for _, item := range r.Items {
			fileIDs = append(fileIDs, item.Id)
		}
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return fileIDs, nil
}

// 	DownloadFileByID downloads the file with the given ID to a temp file on disk and returns its path.
//	Params:
//		fileID string: the ID of the file to download
//	Returns:
//		string: the path to the local temp video file
//		error
func (gduc *GoogleDriveUserClient) DownloadFileByID(fileID string) (string, error) {
	response, err := gduc.service.Files.Get(fileID).Download()
	if err != nil {
		return "", fmt.Errorf("service.Files.Get().Download() for fileID %q returns err: %w", fileID, err)
	}
	if response.ContentLength >= constants.MaxVideoFileSizeMB*1024*1024 {
		return "", fmt.Errorf("file with ID %q exceed maximum video file size of %d MB", fileID, constants.MaxVideoFileSizeMB)
	}
	videoBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error readying bytes from file with ID %q - err: %w", fileID, err)
	}

	tempFile, err := mp4util.CreateTempMP4File(fmt.Sprintf("%s.mp4", fileID))
	if err != nil {
		return "", fmt.Errorf("error creating tmp file with name %s.mp4", fileID)
	}
	if _, err := tempFile.Write(videoBytes); err != nil {
		return "", fmt.Errorf("error writing bytes to temp file %q: %w", tempFile.Name(), err)
	}

	return tempFile.Name(), nil
}

// 	MarkFileByID marks the file with the given ID with the fileSuccessPrefix or fileErrorPrefix.
//	Params:
//		fileID string: the ID of the file to mark
//		failure bool: if true, prefix file name with fileSuccessPrefix, else prefix file name with fileErrorPrefix
//	Returns:
//		error
func (gduc *GoogleDriveUserClient) MarkFileByID(fileID string, failure bool) error {
	file, err := gduc.service.Files.Get(fileID).Do()
	if err != nil {
		return fmt.Errorf("gduc.service.Files.Get(%s).Do() returns err: %w", fileID, err)
	}

	prefix := "SUCCESS_"
	if failure {
		prefix = "ERROR_"
	}
	originalName := file.Title
	file.Title = fmt.Sprintf("%s%s", prefix, originalName)
	if _, err := gduc.service.Files.Update(file.Id, file).Do(); err != nil {
		return fmt.Errorf("error patching file with original name %s - err: %w", originalName, err)
	}

	return nil
}
