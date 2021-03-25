package logging

import "fmt"

type LocalLogger struct{}

func NewLocalLogger() *LocalLogger {
	return &LocalLogger{}
}

func (l *LocalLogger) Log(message string) {
	fmt.Printf("LOG: %s\n", message)
}

func (l *LocalLogger) Warning(message string) {
	fmt.Printf("WARNING: %s\n", message)
}

func (l *LocalLogger) Error(message string) {
	fmt.Printf("ERROR: %s\n", message)
}

func (l *LocalLogger) Critical(message string) {
	critical := "********************\nCRITICAL\n********************\n"
	fmt.Printf("%sCRITICAL: %s\n", critical, message)
}
