package syncer

import (
	"fmt"
	st "seneca/api/type"
	"seneca/env"
	"seneca/internal/client/cloud"
	"seneca/internal/client/googledrive"
	"seneca/internal/controller/syncer"
	"seneca/internal/datagatherer/rawvideohandler"
	"seneca/internal/util/data"
	"seneca/internal/util/mp4"
	"seneca/test/integrationtest/testenv"
	"sort"
)

func E2ESyncer(testUserEmail string, testEnv *testenv.TestEnvironment) error {
	wantRawVideos := []*st.RawVideo{
		{
			CreateTimeMs: 1617572239000,
			DurationMs:   60000,
			UserId:       "5642368648740864",
		},
	}

	// TODO(lucaloncar): test expected raw locations and motions

	if err := env.ValidateEnvironmentVariables(); err != nil {
		return fmt.Errorf("failed to validate environment variables: %w", err)
	}

	mp4Tool, err := mp4.NewMP4Tool(testEnv.Logger)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("mp4.NewMP4Tool() returns - err: %v", err))
	}

	rawVideoHandler, err := rawvideohandler.NewRawVideoHandler(testEnv.SimpleStorage, mp4Tool, testEnv.RawVideoDAO, testEnv.RawLocationDAO, testEnv.RawMotionDAO, testEnv.Logger, testEnv.ProjectID)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("cloud.NewRawVideoHandler() returns - err: %v", err))
	}

	gDriveFactory := &googledrive.UserClientFactory{}

	syncer := syncer.New(rawVideoHandler, gDriveFactory, testEnv.UserDAO, testEnv.Logger)

	user, err := testEnv.UserDAO.GetUserByEmail(testUserEmail)
	if err != nil {
		return fmt.Errorf("GetUserByEmail(%s) returns err: %w", testUserEmail, err)
	}

	// Clean up.
	defer func() {
		rawVideoIDs, err := testEnv.RawVideoDAO.ListUserRawVideoIDs(user.Id)
		if err != nil {
			testEnv.Logger.Error(fmt.Sprintf("ListUserRawVideoIDs(%s) returns err: %v", user.Id, err))
		}
		for _, rid := range rawVideoIDs {
			rawVideo, err := testEnv.RawVideoDAO.GetRawVideoByID(rid)
			if err != nil {
				testEnv.Logger.Error(fmt.Sprintf("GetRawVideoByID(%s) for user %q returns err: %v", rid, user.Id, err))
			}

			if err := testEnv.SimpleStorage.DeleteBucketFile(cloud.RawVideoBucketName, rawVideo.CloudStorageFileName); err != nil {
				testEnv.Logger.Error(fmt.Sprintf("DeleteBucketFile(%s, %s) for user %q returns err: %v", cloud.RawVideoBucketName, rawVideo.CloudStorageFileName, user.Id, err))
			}

			if err := testEnv.RawVideoDAO.DeleteRawVideoByID(rid); err != nil {
				testEnv.Logger.Error(fmt.Sprintf("DeleteRawVideoByID(%s) for user %q returns err: %v", rid, user.Id, err))
			}
		}

		rawLocationIDs, err := testEnv.RawLocationDAO.ListUserRawLocationIDs(user.Id)
		if err != nil {
			testEnv.Logger.Error(fmt.Sprintf("ListUserRawLocationIDs(%s) returns err: %v", user.Id, err))
		}
		for _, rlid := range rawLocationIDs {
			if err := testEnv.RawLocationDAO.DeleteRawLocationByID(rlid); err != nil {
				testEnv.Logger.Error(fmt.Sprintf("DeleteRawLocationByID(%s) for user %q returns err: %v", rlid, user.Id, err))
			}
		}

		rawMotionIDs, err := testEnv.RawMotionDAO.ListUserRawMotionIDs(user.Id)
		if err != nil {
			testEnv.Logger.Error(fmt.Sprintf("ListUserRawMotionIDs(%s) returns err: %v", user.Id, err))
		}
		for _, rlid := range rawMotionIDs {
			if err := testEnv.RawMotionDAO.DeleteRawMotionByID(rlid); err != nil {
				testEnv.Logger.Error(fmt.Sprintf("DeleteRawMotionByID(%s) for user %q returns err: %v", rlid, user.Id, err))
			}
		}

		gDrive, err := gDriveFactory.New(user)
		if err != nil {
			testEnv.Logger.Error(fmt.Sprintf("error initializing gdrive client for user %q", user.Id))
			return
		}
		fileIDs, err := gDrive.ListFileIDs(googledrive.AllMP4s)
		if err != nil {
			testEnv.Logger.Error(fmt.Sprintf("error listing all file IDs for user %q", user.Id))
		}
		prefixes := []googledrive.FilePrefix{googledrive.WorkInProgress, googledrive.Error}
		for _, fid := range fileIDs {
			for _, prefix := range prefixes {
				if err := gDrive.MarkFileByID(fid, prefix, true); err != nil {
					testEnv.Logger.Error(fmt.Sprintf("gDrive.MarkFileByID(%s, %s, true) for user %q returns err: %v", fid, prefix, user.Id, err))
				}
			}
		}
	}()

	if err := syncer.SyncUser(user.Id); err != nil {
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
	}

	return nil
}
