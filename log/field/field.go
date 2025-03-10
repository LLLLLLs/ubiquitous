package field

import (
	"encoding/json"
	"go.uber.org/zap"
)

type M map[string]interface{}

type Field = zap.Field

type Fields []Field

func NewFields(fields ...Field) *Fields {
	f := Fields(fields)
	return &f
}

func (fs *Fields) With(fields ...Field) *Fields {
	*fs = append(*fs, fields...)
	return fs
}

func (fs *Fields) List() []Field {
	return *fs
}

// 有需要新的Field可以在这里引入
var (
	String    = zap.String
	Bool      = zap.Bool
	Error     = zap.Error
	StackSkip = zap.StackSkip
	Any       = zap.Any
	AnyString = func(key string, value interface{}) Field {
		valData, _ := json.Marshal(value)
		return zap.String(key, string(valData))
	}
	// int
	Int    = zap.Int
	Int32  = zap.Int32
	Int64  = zap.Int64
	Int64s = zap.Int64s
	// uint
	Binary = zap.Binary
	Uint8  = zap.Uint8
	Uint64 = zap.Uint64
	Uint32 = zap.Uint32
	// time
	Duration = zap.Duration
)
