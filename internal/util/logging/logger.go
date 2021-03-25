package logging

type LoggingInterface interface {
	Log(message string)
	Warning(message string)
	Error(message string)
	Critical(message string)
}
