package data

import (
	"context"
	"fmt"
	"seneca/api/constants"
	"seneca/internal/client/cloud"
	"seneca/internal/client/database"
	"seneca/internal/client/logging"
	"seneca/internal/dao/rawframedao"
	"seneca/internal/dao/rawlocationdao"
	"seneca/internal/dao/rawmotiondao"
	"seneca/internal/dao/rawvideodao"
)

func DeleteAllUserData(userID string, includeUser bool, sqlInterface database.SQLInterface, storageClient cloud.SimpleStorageInterface, logger logging.LoggingInterface) error {

	// Delete all CloudStorage data.
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(sqlInterface, logger, 0)
	rawVideoIDs, err := rawVideoDAO.ListUserRawVideoIDs(userID)
	if err != nil {
		return fmt.Errorf("ListUserRawVideoIDs(%s,) returns err: %v", userID, err)
	}
	for _, rvid := range rawVideoIDs {
		rawVideo, err := rawVideoDAO.GetRawVideoByID(rvid)
		if err != nil {
			return fmt.Errorf("GetRawVideoByID(%s) returns err: %v", rvid, err)
		}

		if rawVideo.CloudStorageFileName == "" {
			continue
		}

		bucketName, fileName, err := GCSURLToBucketNameAndFileName(rawVideo.CloudStorageFileName)
		if err != nil {
			return fmt.Errorf("GCSURLToBucketNameAndFileName(%s) returns err: %v", rawVideo.CloudStorageFileName, err)
		}

		storageClient.DeleteBucketFile(bucketName, fileName)
	}
	rawFrameDAO := rawframedao.NewSQLRawFrameDAO(sqlInterface)
	rawFrameIDs, err := rawFrameDAO.ListUserRawFrameIDs(userID)
	if err != nil {
		return fmt.Errorf("ListUserRawVideoIDs(%s,) returns err: %v", userID, err)
	}
	for _, rfid := range rawFrameIDs {
		rawFrame, err := rawFrameDAO.GetRawFrameByID(rfid)
		if err != nil {
			return fmt.Errorf("GetRawFrameByID(%s) returns err: %v", rfid, err)
		}

		if rawFrame.CloudStorageFileName == "" {
			continue
		}

		bucketName, fileName, err := GCSURLToBucketNameAndFileName(rawFrame.CloudStorageFileName)
		if err != nil {
			return fmt.Errorf("GCSURLToBucketNameAndFileName(%s) returns err: %v", rawFrame.CloudStorageFileName, err)
		}

		storageClient.DeleteBucketFile(bucketName, fileName)
	}

	for _, tableName := range constants.DataTableNames {
		if tableName == constants.UsersTable && includeUser {
			if err := sqlInterface.DeleteByID(tableName, userID); err != nil {
				return fmt.Errorf("DeleteByID(%s, %s) returns err: %w", tableName, userID, err)
			}
		}

		ids, err := sqlInterface.ListIDs(tableName, []*database.QueryParam{{FieldName: constants.UserIDFieldName, Operand: "=", Value: userID}})
		if err != nil {
			return fmt.Errorf("ListIDs(%s, %s) returns err: %w", tableName, userID, err)
		}
		for _, id := range ids {
			if err := sqlInterface.DeleteByID(tableName, id); err != nil {
				return fmt.Errorf("DeleteByID(%s, %s) returns err: %w", tableName, id, err)
			}
		}
	}

	return nil
}

func RemoveAllUserAlgoTagsInDB(userID string, dbClient database.SQLInterface, logger logging.LoggingInterface) error {
	rawVideoDAO := rawvideodao.NewSQLRawVideoDAO(dbClient, logger, 0)
	rawLocationDAO := rawlocationdao.NewSQLRawLocationDAO(dbClient)
	rawMotionDAO := rawmotiondao.NewSQLRawMotionDAO(dbClient, logger)
	rawFrameDAO := rawframedao.NewSQLRawFrameDAO(dbClient)

	unprocessedRawVideoIDs, err := rawVideoDAO.ListUserRawVideoIDs(userID)
	if err != nil {
		return fmt.Errorf("ListUserRawVideoIDs(%s,) returns err: %v", userID, err)
	}

	for _, urvid := range unprocessedRawVideoIDs {
		rawVideo, err := rawVideoDAO.GetRawVideoByID(urvid)
		if err != nil {
			logger.Error(fmt.Sprintf("GetRawVideoByID(%s) returns err: %v", urvid, err))
		}
		if rawVideo.AlgosVersion != 0 {
			rawVideo.AlgosVersion = 0
			if err := rawVideoDAO.PutRawVideoByID(context.TODO(), rawVideo.Id, rawVideo); err != nil {
				return fmt.Errorf("PutRawVideoByID() returns err: %w", err)
			}
		}
	}

	unprocessedRawMotionIDs, err := rawMotionDAO.ListUserRawMotionIDs(userID)
	if err != nil {
		logger.Error(fmt.Sprintf("ListUserRawMotionIDs(%s,) returns err: %v", userID, err))
	}

	for _, urmid := range unprocessedRawMotionIDs {
		rawMotion, err := rawMotionDAO.GetRawMotionByID(urmid)
		if err != nil {
			logger.Error(fmt.Sprintf("GetRawMotionByID(%s) returns err: %v", urmid, err))
		}
		if rawMotion.AlgosVersion != 0 {
			rawMotion.AlgosVersion = 0
			if err := rawMotionDAO.PutRawMotionByID(context.TODO(), rawMotion.Id, rawMotion); err != nil {
				return fmt.Errorf("PutRawMotionByID() returns err: %w", err)
			}
		}
	}

	locationIDs, err := rawLocationDAO.ListUserRawLocationIDs(userID)
	if err != nil {
		logger.Error(fmt.Sprintf("ListUserRawLocationIDs(%s) returns err: %v", userID, err))
	}

	for _, lid := range locationIDs {
		rawLocation, err := rawLocationDAO.GetRawLocationByID(lid)
		if err != nil {
			logger.Error(fmt.Sprintf("GetRawLocationByID(%s) returns err: %v", lid, err))
			continue
		}
		if rawLocation.AlgosVersion != 0 {
			rawLocation.AlgosVersion = 0
			if err := rawLocationDAO.PutRawLocationByID(context.TODO(), rawLocation.Id, rawLocation); err != nil {
				return fmt.Errorf("PutRawLocationByID() returns err: %w", err)
			}
		}
	}

	unprocessedRawFrameIDs, err := rawFrameDAO.ListUserRawFrameIDs(userID)
	if err != nil {
		logger.Error(fmt.Sprintf("ListUserRawFrameIDs(%s) returns err: %v", userID, err))
	}

	for _, rfid := range unprocessedRawFrameIDs {
		rawFrame, err := rawFrameDAO.GetRawFrameByID(rfid)
		if err != nil {
			logger.Error(fmt.Sprintf("GetRawFrameByID(%s) returns err: %v", rfid, err))
			continue
		}
		if rawFrame.AlgosVersion != 0 {
			rawFrame.AlgosVersion = 0
			if err := rawFrameDAO.PutRawFrameByID(context.TODO(), rawFrame.Id, rawFrame); err != nil {
				return fmt.Errorf("PutRawFrameByID() returns err: %w", err)
			}
		}
	}

	return nil
}
