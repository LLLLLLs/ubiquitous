package impl

import (
	"go.uber.org/zap/zapcore"
	"time"
)

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
