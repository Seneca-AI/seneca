package tests

import (
	"fmt"
	st "seneca/api/type"
	"seneca/env"
	"seneca/internal/client/googledrive"
	"seneca/internal/util/data"
	"seneca/test/integrationtest/testenv"
	"sort"
	"strings"
)

func E2ESyncer(testUserEmail string, testEnv *testenv.TestEnvironment) error {
	defer testEnv.Clean()

	wantRawVideos := []*st.RawVideo{
		{
			CreateTimeMs:     1617554180000,
			DurationMs:       60000,
			UserId:           "5642368648740864",
			OriginalFileName: "three.mp4",
		},
	}

	// TODO(lucaloncar): test expected raw locations and motions
	if err := env.ValidateEnvironmentVariables(); err != nil {
		return fmt.Errorf("failed to validate environment variables: %w", err)
	}

	user, err := testEnv.UserDAO.GetUserByEmail(testUserEmail)
	if err != nil {
		return fmt.Errorf("GetUserByEmail(%s) returns err: %w", testUserEmail, err)
	}

	if err := testEnv.Syncer.SyncUser(user.Id); err != nil {
		return fmt.Errorf("SyncUser() returns err: %w", err)
	}

	// Verify files exist.
	gotRawVideoIDs, err := testEnv.RawVideoDAO.ListUserRawVideoIDs(user.Id)
	if err != nil {
		return fmt.Errorf("error list raw videos for user %q - err: %w", user.Id, err)
	}

	if len(gotRawVideoIDs) != len(wantRawVideos) {
		return fmt.Errorf("want %d raw videos for user %q, got %d", len(wantRawVideos), user.Id, len(gotRawVideoIDs))
	}

	gotRawVideos := []*st.RawVideo{}
	for _, gotID := range gotRawVideoIDs {
		gotRawVideo, err := testEnv.RawVideoDAO.GetRawVideoByID(gotID)
		if err != nil {
			return fmt.Errorf("error getting raw video with id %q for user %q", gotID, user.Id)
		}

		gotRawVideos = append(gotRawVideos, gotRawVideo)
	}

	sort.Slice(gotRawVideos, func(i, j int) bool { return gotRawVideos[i].CreateTimeMs < gotRawVideos[j].CreateTimeMs })

	for i := range gotRawVideos {
		if err := data.RawVideosEqual(gotRawVideos[i], wantRawVideos[i]); err != nil {
			return fmt.Errorf("raw videos not equal (got != want): %w", err)
		}

		bucketName, fileName, err := data.GCSURLToBucketNameAndFileName(gotRawVideos[i].CloudStorageFileName)
		if err != nil {
			return fmt.Errorf("GCSURLToBucketNameAndFileName() returns err: %w", err)
		}

		exists, err := testEnv.SimpleStorage.BucketFileExists(bucketName, fileName)
		if err != nil {
			return fmt.Errorf("BucketFileExists(%s, %s) returns err: %w", bucketName, fileName, err)
		}

		if !exists {
			return fmt.Errorf("BucketFileExists(%s, %s) returns false", bucketName, fileName)
		}
	}

	// Verify derivative data exists.
	rawLocationIDs, err := testEnv.RawLocationDAO.ListUserRawLocationIDs(user.Id)
	if err != nil {
		return fmt.Errorf("ListUserRawLocationIDs() returns err: %w", err)
	}
	if len(rawLocationIDs) != 60 {
		return fmt.Errorf("want %d rawLocationIDs, got %d", 60, len(rawLocationIDs))
	}
	rawMotionIDs, err := testEnv.RawMotionDAO.ListUserRawMotionIDs(user.Id)
	if err != nil {
		return fmt.Errorf("ListUserRawMotionIDs() returns err: %w", err)
	}
	if len(rawMotionIDs) != 60 {
		return fmt.Errorf("want %d rawMotionIDs, got %d", 60, len(rawMotionIDs))
	}
	rawFrameIDs, err := testEnv.RawFrameDAO.ListUserRawFrameIDs(user.Id)
	if err != nil {
		return fmt.Errorf("ListUserRawFrameIDs() returns err: %w", err)
	}
	if len(rawFrameIDs) != 60 {
		return fmt.Errorf("want %d rawFrameIDs, got %d", 60, len(rawFrameIDs))
	}
	for _, rfid := range rawFrameIDs {
		rawFrame, err := testEnv.RawFrameDAO.GetRawFrameByID(rfid)
		if err != nil {
			return fmt.Errorf("GetRawFrameByID(%s) returns err: %w", rfid, err)
		}

		if rawFrame.Source.SourceType != st.Source_RAW_VIDEO {
			return fmt.Errorf("want %q for rawFrame.Source.SourceType, got %q", st.Source_RAW_VIDEO, rawFrame.Source.SourceType)
		}

		bucketName, fileName, err := data.GCSURLToBucketNameAndFileName(rawFrame.CloudStorageFileName)
		if err != nil {
			return fmt.Errorf("GCSURLToBucketNameAndFileName() returns err: %w", err)
		}

		exists, err := testEnv.SimpleStorage.BucketFileExists(bucketName, fileName)
		if err != nil {
			return fmt.Errorf("BucketFileExists(%s, %s) returns err: %w", bucketName, fileName, err)
		}

		if !exists {
			return fmt.Errorf("BucketFileExists(%s, %s) returns false", bucketName, fileName)
		}
	}

	// Make sure all files are marked SUCCESS.
	userGDriveClient, err := testEnv.GDriveFactory.New(user)
	if err != nil {
		return fmt.Errorf("GDriveFactory.New() returns err: %w", err)
	}

	fileIDs, err := userGDriveClient.ListFileIDs(googledrive.AllMP4s)
	if err != nil {
		return fmt.Errorf("userGDriveClient.ListFileIDs() returns err: %w", err)
	}

	for _, fid := range fileIDs {
		fileInfo, err := userGDriveClient.GetFileInfo(fid)
		if err != nil {
			return fmt.Errorf("userGDriveClient.GetFileInfo() returns err: %w", err)
		}

		if !strings.HasPrefix(fileInfo.FileName, googledrive.Success.String()) {
			return fmt.Errorf("file with name %q for user with email %q not marked with SUCECSS_ prefix", fileInfo.FileName, user.Email)
		}
	}

	if testEnv.Logger.Failures() > 0 {
		return fmt.Errorf("got %d logging failures", testEnv.Logger.Failures())
	}

	return nil
}
