package log

import (
	"github.com/LLLLLLs/ubiquitous/log/field"
)

type nopLogger struct{}

func (n nopLogger) With(_ ...field.Field) Logger { return n }

func (n nopLogger) Debug(_ string, _ ...field.Field) {}

func (n nopLogger) Info(_ string, _ ...field.Field) {}

func (n nopLogger) Warn(_ string, _ ...field.Field) {}

func (n nopLogger) Error(_ string, _ ...field.Field) {}
