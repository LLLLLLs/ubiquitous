package log

import (
	"github.com/LLLLLLs/ubiquitous/log/field"
	"log"
)

type defaultLogger struct {
	def    *log.Logger
	fields []field.Field
}

func (d defaultLogger) With(field ...field.Field) Logger {
	return defaultLogger{
		def:    d.def,
		fields: append(d.fields, field...),
	}
}

func (d defaultLogger) Debug(msg string, field ...field.Field) {
	d.print("Debug", msg, field...)
}

func (d defaultLogger) Info(msg string, field ...field.Field) {
	d.print("Info", msg, field...)
}

func (d defaultLogger) Warn(msg string, field ...field.Field) {
	d.print("Warn", msg, field...)
}

func (d defaultLogger) Error(msg string, field ...field.Field) {
	d.print("Error", msg, field...)
}

func (d defaultLogger) print(level, msg string, field ...field.Field) {
	d.def.Printf("level:%s,msg:%s,fields:%v\n", level, msg, append(d.fields, field...))
}
