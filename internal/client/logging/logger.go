package logging

import (
	"errors"
	"seneca/api/senecaerror"
)

// LoggingInterface is the interface for logging across Seneca.
//nolint
type LoggingInterface interface {
	// Log logs the message at default severity.
	// Params:
	//		message string: the message to log
	Log(message string)
	// Warning logs the message at warning severity.
	// Params:
	//		message string: the message to log
	Warning(message string)
	// Error logs the message at error severity.
	// Params:
	//		message string: the message to log
	Error(message string)
	// Critical logs the message at critical severity.
	// Params:
	//		message string: the message to log
	Critical(message string)
}

// LogSenecaError chooses the error severity based on the senecaerror type.
func LogSenecaError(logger LoggingInterface, err error) {
	var ue *senecaerror.UserError
	if errors.As(err, &ue) {
		logger.Warning(err.Error())
	} else {
		logger.Error(err.Error())
	}
}
