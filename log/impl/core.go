package impl

import (
	"github.com/LLLLLLs/ubiquitous/log"
	"path/filepath"
)

// Level  开启的日志等级
type Level string

const (
	DEBUG Level = "debug"
	INFO  Level = "info"
	WARN  Level = "warn"
	ERROR Level = "error"
)

var (
	lg           = log.NewDefaultLogger()
	fileSeparate = string(filepath.Separator)
)
