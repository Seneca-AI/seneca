package syncer

import (
	"context"
	"fmt"
	st "seneca/api/type"
	"seneca/env"
	"seneca/internal/client/cloud"
	"seneca/internal/client/cloud/gcp"
	"seneca/internal/client/cloud/gcpdatastore"
	"seneca/internal/client/googledrive"
	"seneca/internal/client/logging"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
	"seneca/internal/dao/userdao"
	"seneca/internal/datagatherer/rawvideohandler"
	"seneca/internal/syncer"
	"seneca/internal/util/data"
	"seneca/internal/util/mp4"
	"sort"
	"time"
)

const testUserEmail = "itestuser000@senecacam.com"

func E2ESyncer(projectID string) error {
	wantRawVideos := []*st.RawVideo{
		{
			CreateTimeMs: 1617572239000,
			DurationMs:   60000,
			UserId:       "5642368648740864",
		},
		{
			CreateTimeMs: 1617574040000,
			DurationMs:   60000,
			UserId:       "5642368648740864",
		},
		{
			CreateTimeMs: 1617574444000,
			DurationMs:   60000,
			UserId:       "5642368648740864",
		},
	}

	// TODO(lucaloncar): test expected raw locations and motions

	if err := env.ValidateEnvironmentVariables(); err != nil {
		return fmt.Errorf("failed to validate environment variables: %w", err)
	}

	ctx := context.TODO()
	// Initialize clients.
	logger, err := logging.NewGCPLogger(ctx, "singleserver", projectID)
	if err != nil {
		return fmt.Errorf("logging.NewGCPLogger() returns - err: %v", err)
	}

	gcsc, err := gcp.NewGoogleCloudStorageClient(ctx, projectID, time.Second*10, time.Minute)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("cloud.NewGoogleCloudStorageClient() returns - err: %v", err))
	}

	sqlService, err := gcpdatastore.New(context.TODO(), projectID)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("error initializing sql service - err: %v", err))
	}

	userDAO := userdao.NewSQLUserDao(sqlService)
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlService, time.Second*5)
	rawLocationDAO := rawlocationdao.NewSQLRawLocationDAO(sqlService)
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(sqlService)

	mp4Tool, err := mp4.NewMP4Tool(logger)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("mp4.NewMP4Tool() returns - err: %v", err))
	}

	rawVideoHandler, err := rawvideohandler.NewRawVideoHandler(gcsc, mp4Tool, rawVideoDAO, rawLocationDAO, rawMotionDAO, logger, projectID)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("cloud.NewRawVideoHandler() returns - err: %v", err))
	}

	gDriveFactory := &googledrive.UserClientFactory{}

	syncer := syncer.New(rawVideoHandler, gDriveFactory, userDAO, logger)

	user, err := userDAO.GetUserByEmail(testUserEmail)
	if err != nil {
		return fmt.Errorf("GetUserByEmail(%s) returns err: %w", testUserEmail, err)
	}

	// Clean up.
	defer func() {
		rawVideoIDs, err := rawVideoDAO.ListUserRawVideoIDs(user.Id)
		if err != nil {
			logger.Error(fmt.Sprintf("ListUserRawVideoIDs(%s) returns err: %v", user.Id, err))
		}
		for _, rid := range rawVideoIDs {
			rawVideo, err := rawVideoDAO.GetRawVideoByID(rid)
			if err != nil {
				logger.Error(fmt.Sprintf("GetRawVideoByID(%s) for user %q returns err: %v", rid, user.Id, err))
			}

			if err := gcsc.DeleteBucketFile(cloud.RawVideoBucketName, rawVideo.CloudStorageFileName); err != nil {
				logger.Error(fmt.Sprintf("DeleteBucketFile(%s, %s) for user %q returns err: %v", cloud.RawVideoBucketName, rawVideo.CloudStorageFileName, user.Id, err))
			}

			if err := rawVideoDAO.DeleteRawVideoByID(rid); err != nil {
				logger.Error(fmt.Sprintf("DeleteRawVideoByID(%s) for user %q returns err: %v", rid, user.Id, err))
			}
		}

		rawLocationIDs, err := rawLocationDAO.ListUserRawLocationIDs(user.Id)
		if err != nil {
			logger.Error(fmt.Sprintf("ListUserRawLocationIDs(%s) returns err: %v", user.Id, err))
		}
		for _, rlid := range rawLocationIDs {
			if err := rawLocationDAO.DeleteRawLocationByID(rlid); err != nil {
				logger.Error(fmt.Sprintf("DeleteRawLocationByID(%s) for user %q returns err: %v", rlid, user.Id, err))
			}
		}

		rawMotionIDs, err := rawMotionDAO.ListUserRawMotionIDs(user.Id)
		if err != nil {
			logger.Error(fmt.Sprintf("ListUserRawMotionIDs(%s) returns err: %v", user.Id, err))
		}
		for _, rlid := range rawMotionIDs {
			if err := rawMotionDAO.DeleteRawMotionByID(rlid); err != nil {
				logger.Error(fmt.Sprintf("DeleteRawMotionByID(%s) for user %q returns err: %v", rlid, user.Id, err))
			}
		}

		gDrive, err := gDriveFactory.New(user)
		if err != nil {
			logger.Error(fmt.Sprintf("error initializing gdrive client for user %q", user.Id))
			return
		}
		fileIDs, err := gDrive.ListFileIDs(googledrive.AllMP4s)
		if err != nil {
			logger.Error(fmt.Sprintf("error listing all file IDs for user %q", user.Id))
		}
		prefixes := []googledrive.FilePrefix{googledrive.WorkInProgress, googledrive.Error}
		for _, fid := range fileIDs {
			for _, prefix := range prefixes {
				if err := gDrive.MarkFileByID(fid, prefix, true); err != nil {
					logger.Error(fmt.Sprintf("gDrive.MarkFileByID(%s, %s, true) for user %q returns err: %v", fid, prefix, user.Id, err))
				}
			}
		}
	}()

	if err := syncer.SyncUser(user.Id); err != nil {
		return fmt.Errorf("SyncUser() returns err: %w", err)
	}

	// Verify files exist.
	gotRawVideoIDs, err := rawVideoDAO.ListUserRawVideoIDs(user.Id)
	if err != nil {
		return fmt.Errorf("error list raw videos for user %q - err: %w", user.Id, err)
	}

	if len(gotRawVideoIDs) != len(wantRawVideos) {
		return fmt.Errorf("want %d raw videos for user %q, got %d", len(wantRawVideos), user.Id, len(gotRawVideoIDs))
	}

	gotRawVideos := []*st.RawVideo{}
	for _, gotID := range gotRawVideoIDs {
		gotRawVideo, err := rawVideoDAO.GetRawVideoByID(gotID)
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
