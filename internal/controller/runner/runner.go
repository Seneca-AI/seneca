package runner

import (
	"fmt"
	"seneca/internal/client/logging"
	"seneca/internal/dao"
	"seneca/internal/dataprocessor"
)

type Runner struct {
	dataprocessor *dataprocessor.DataProcessor
	userDAO       dao.UserDAO
	logger        logging.LoggingInterface
}

func New(userDAO dao.UserDAO, dataprocessor *dataprocessor.DataProcessor, logger logging.LoggingInterface) *Runner {
	return &Runner{
		dataprocessor: dataprocessor,
		userDAO:       userDAO,
		logger:        logger,
	}
}

func (rnr *Runner) Run() {
	userIDs, err := rnr.userDAO.ListAllUserIDs()
	if err != nil {
		rnr.logger.Error(fmt.Sprintf("ListAllUserIDs() returns err: %v", err))
	}

	for _, uid := range userIDs {
		rnr.dataprocessor.Run(uid)
	}
}
