package logging

import (
	"context"
	"fmt"

	"cloud.google.com/go/logging"
)

// GCPLogger is the client for logging to Google Cloud Platform.
type GCPLogger struct {
	client *logging.Client
	logger *logging.Logger
}

// NewGCPLogger initializes the GCPLogger.
// Params:
//		ctx context.Context
//		logName string: what the log will be named in the logs explorer, like a key prefix
//		projectID string
// Returns:
//		*GCPLogger
//		error
func NewGCPLogger(ctx context.Context, logName, projectID string) (*GCPLogger, error) {
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error initializing new GCP logging client - err: %v", err)
	}
	return &GCPLogger{
		client: client,
		logger: client.Logger(logName),
	}, nil
}

// Log logs the given message at log level logging.Default.
func (l *GCPLogger) Log(message string) {
	l.logger.Log(logging.Entry{
		Payload: struct{ Anything string }{
			Anything: message,
		},
		Severity: logging.Default,
	})
}

// Warning logs the given message at the log level logging.Warning.
func (l *GCPLogger) Warning(message string) {
	l.logger.Log(logging.Entry{
		Payload: struct{ Anything string }{
			Anything: message,
		},
		Severity: logging.Warning,
	})
}

// Error logs the given message at the log level logging.Error.
func (l *GCPLogger) Error(message string) {
	l.logger.Log(logging.Entry{
		Payload: struct{ Anything string }{
			Anything: message,
		},
		Severity: logging.Error,
	})
}

// Critical logs the given message at the log level logging.Critical.
func (l *GCPLogger) Critical(message string) {
	l.logger.Log(logging.Entry{
		Payload: struct{ Anything string }{
			Anything: message,
		},
		Severity: logging.Critical,
	})
}
