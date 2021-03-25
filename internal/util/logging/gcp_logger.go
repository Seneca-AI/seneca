package logging

import (
	"context"
	"fmt"

	"cloud.google.com/go/logging"
)

type GCPLogger struct {
	client *logging.Client
	logger *logging.Logger
}

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

func (l *GCPLogger) Log(message string) {
	l.logger.Log(logging.Entry{
		Payload: struct{ Anything string }{
			Anything: message,
		},
		Severity: logging.Default,
	})
}

func (l *GCPLogger) Warning(message string) {
	l.logger.Log(logging.Entry{
		Payload: struct{ Anything string }{
			Anything: message,
		},
		Severity: logging.Warning,
	})
}

func (l *GCPLogger) Error(message string) {
	l.logger.Log(logging.Entry{
		Payload: struct{ Anything string }{
			Anything: message,
		},
		Severity: logging.Error,
	})
}

func (l *GCPLogger) Critical(message string) {
	l.logger.Log(logging.Entry{
		Payload: struct{ Anything string }{
			Anything: message,
		},
		Severity: logging.Critical,
	})
}
