package logging

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
