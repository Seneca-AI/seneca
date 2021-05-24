package logging

import "fmt"

// LocalLogger is the logging client that simply print log messages to the console.
type LocalLogger struct {
	silent bool
}

// TODO(lucaloncar): instead of 'silent', take min log level as input
// 	NewLocalLogger returns an instance of the LocalLogger.
// 	Params:
//		silent bool: whether the logger shouldn't log anything
// 	Returns:
//		*LocalLogger
func NewLocalLogger(silent bool) *LocalLogger {
	return &LocalLogger{
		silent: silent,
	}
}

// Log prints the message to standard out with a "LOG: " prefix.
func (l *LocalLogger) Log(message string) {
	if !l.silent {
		fmt.Printf("LOG: %s\n", message)
	}
}

// Warning prints the message to standard out with a "WARNING: " prefix.
func (l *LocalLogger) Warning(message string) {
	if !l.silent {
		fmt.Printf("WARNING: %s\n", message)
	}
}

// Error prints the message to standard out with a "ERROR: " prefix.
func (l *LocalLogger) Error(message string) {
	if !l.silent {
		fmt.Printf("ERROR: %s\n", message)
	}
}

// Critical prints the message to standard out with a "CRITICAL: " prefix.
func (l *LocalLogger) Critical(message string) {
	if !l.silent {
		critical := "********************\nCRITICAL\n********************\n"
		fmt.Printf("%sCRITICAL: %s\n", critical, message)
	}
}
