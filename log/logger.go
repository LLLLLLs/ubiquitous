package log

import (
	"github.com/LLLLLLs/ubiquitous/log/field"
	"log"
)

type Logger interface {
	With(field ...field.Field) Logger
	Debug(msg string, field ...field.Field)
	Info(msg string, field ...field.Field)
	Warn(msg string, field ...field.Field)
	Error(msg string, field ...field.Field)
}

func NewDefaultLogger() Logger {
	return &defaultLogger{def: log.Default()}
}

func NewNopLogger() Logger {
	return &nopLogger{}
}
