package field

import (
	"fmt"
	"time"
)

// 常用的自定义通用字段

var (
	ErrorAny = func(err interface{}) Field {
		if e, ok := err.(error); ok {
			return Error(e)
		}
		return String("error", fmt.Sprintf("%v", err))
	}
)

var (
	TraceId   = func(traceId string) Field { return String("trace_id", traceId) }
	Session   = func(sessionId uint64) Field { return Uint64("session", sessionId) }
	CostMicro = func(start time.Time) Field {
		return Int64("cost_micro", time.Since(start).Microseconds())
	}
)
