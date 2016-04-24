package telebot

import "fmt"

type Logger interface {
	Warningf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

type StdoutLogger struct {
}

func (l *StdoutLogger) print(prefix string, format string, args ...interface{}) {
	fmt.Printf(prefix+format+"\n", args...)
}

func (l *StdoutLogger) Warningf(format string, args ...interface{}) {
	l.print("[WARNING] ", format, args...)
}
func (l *StdoutLogger) Debugf(format string, args ...interface{}) {
	l.print("[DEBUG] ", format, args...)
}

func (l *StdoutLogger) Errorf(format string, args ...interface{}) {
	l.print("[ERROR] ", format, args...)
}

func (l *StdoutLogger) Infof(format string, args ...interface{}) {
	l.print("[INFO] ", format, args...)
}

func NewStdoutLogger() Logger {
	return &StdoutLogger{}
}

