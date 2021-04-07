package googledrive

import "fmt"

type FakeGoogleDriveUserClient struct {
	ListFileIDsMock      func() ([]string, error)
	DownloadFileByIDMock func(fileID string) (string, error)
	MarkFileByIDMock     func(fileID string, failure bool) error
}

func (fgduc *FakeGoogleDriveUserClient) ListFileIDs() ([]string, error) {
	if fgduc.ListFileIDsMock == nil {
		return nil, fmt.Errorf("ListFileIDsMock not set")
	}
	return fgduc.ListFileIDsMock()
}

func (fgduc *FakeGoogleDriveUserClient) DownloadFileByID(fileID string) (string, error) {
	if fgduc.DownloadFileByIDMock == nil {
		return "", fmt.Errorf("DownloadFileByIDMock not set")
	}
	return fgduc.DownloadFileByIDMock(fileID)
}

func (fgduc *FakeGoogleDriveUserClient) MarkFileByID(fileID string, failure bool) error {
	if fgduc.MarkFileByIDMock == nil {
		return fmt.Errorf("MarkFileByIDMock not set")
	}
	return fgduc.MarkFileByIDMock(fileID, failure)
}
