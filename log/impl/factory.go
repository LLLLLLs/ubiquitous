package impl

import (
	"context"
	"github.com/LLLLLLs/ubiquitous/log"
	"github.com/LLLLLLs/ubiquitous/sign"
)

// Init 内置日志库的初始化
func Init(opts ...Option) {
	initLogger(opts...)
}

// New  外部构造的工厂接口 返回对应的logger日志生成器 交由外部进行管理
func New(opts ...Option) log.Logger {
	return newLogger(opts...)
}

// GetLogger returns logger 变量给外部管理
func GetLogger() log.Logger {
	return lg
}

func GetLoggerCtx(ctx context.Context) log.Logger {
	if lg := ctx.Value(sign.LOGGER); lg != nil {
		return lg.(log.Logger)
	}
	if lg := ctx.Value(sign.LOGGER.String()); lg != nil {
		return lg.(log.Logger)
	}
	return GetLogger()
}
