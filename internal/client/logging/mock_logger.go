package logging

import "log"

// MockLogger is used for ensuring that certain log calls were made while testing code.
type MockLogger struct {
	LogMock      func(message string)
	WarningMock  func(message string)
	ErrorMock    func(message string)
	CriticalMock func(message string)
}

func (mlog *MockLogger) Log(message string) {
	if mlog.LogMock == nil {
		log.Fatalf("LogMock not set.")
	}
	mlog.LogMock(message)
}

func (mlog *MockLogger) Warning(message string) {
	if mlog.WarningMock == nil {
		log.Fatalf("WarningMock not set.")
	}
	mlog.WarningMock(message)
}

func (mlog *MockLogger) Error(message string) {
	if mlog.ErrorMock == nil {
		log.Fatalf("ErrorMock not set.")
	}
	mlog.ErrorMock(message)
}

func (mlog *MockLogger) Critical(message string) {
	if mlog.CriticalMock == nil {
		log.Fatalf("CriticalMock not set.")
	}
	mlog.CriticalMock(message)
}
