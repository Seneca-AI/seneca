package logging

import "fmt"

// LocalLogger is the logging client that simply print log messages to the console.
type LocalLogger struct{}

// NewLocalLogger returns an instance of the LocalLogger.
// Params:
// Returns:
//		*LocalLogger
func NewLocalLogger() *LocalLogger {
	return &LocalLogger{}
}

// Log prints the message to standard out with a "LOG: " prefix.
func (l *LocalLogger) Log(message string) {
	fmt.Printf("LOG: %s\n", message)
}

// Warning prints the message to standard out with a "WARNING: " prefix.
func (l *LocalLogger) Warning(message string) {
	fmt.Printf("WARNING: %s\n", message)
}

// Error prints the message to standard out with a "ERROR: " prefix.
func (l *LocalLogger) Error(message string) {
	fmt.Printf("ERROR: %s\n", message)
}

// Critical prints the message to standard out with a "CRITICAL: " prefix.
func (l *LocalLogger) Critical(message string) {
	critical := "********************\nCRITICAL\n********************\n"
	fmt.Printf("%sCRITICAL: %s\n", critical, message)
}
