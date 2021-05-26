package tests

import (
	"fmt"
	st "seneca/api/type"
	"seneca/env"
	"seneca/internal/util/data"
	"seneca/test/integrationtest/testenv"
	"sort"
)

func E2ESyncer(testUserEmail string, testEnv *testenv.TestEnvironment) error {
	defer testEnv.Clean()

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
	}

	return nil
}
