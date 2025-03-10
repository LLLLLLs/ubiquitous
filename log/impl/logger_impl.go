package impl

import (
	"github.com/LLLLLLs/ubiquitous/log"
	"github.com/LLLLLLs/ubiquitous/log/field"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

type logger struct {
	base *zap.Logger
	fs   []field.Field
}

func (l *logger) With(fields ...field.Field) log.Logger {
	return &logger{
		base: l.base,
		fs:   append(l.fs, fields...),
	}
}

func (l *logger) fields(f ...field.Field) []field.Field {
	if len(l.fs) != 0 {
		return append(f, l.fs...)
	}
	return f
}

func (l *logger) Debug(msg string, field ...field.Field) {
	l.base.Debug(msg, l.fields(field...)...)
}

func (l *logger) Info(msg string, field ...field.Field) {
	l.base.Info(msg, l.fields(field...)...)
}

func (l *logger) Warn(msg string, field ...field.Field) {
	l.base.Warn(msg, l.fields(field...)...)
}

func (l *logger) Error(msg string, field ...field.Field) {
	l.base.Error(msg, l.fields(field...)...)
}

func (l *logger) Sync() {
	_ = l.base.Sync()
}

type builder struct {
	conf *config
}

func initLogger(opts ...Option) {
	lg = newLogger(opts...)
}

func newLogger(opts ...Option) log.Logger {
	build := &builder{
		conf: newDefaultConfig(),
	}

	for i := range opts {
		opts[i](build.conf)
	}

	lg := build.build()
	if build.conf.appName != "" {
		lg = lg.With(field.String("app", build.conf.appName))
	}
	if build.conf.regionId != 0 {
		lg = lg.With(field.Int32("region", build.conf.regionId))
	}
	lg.(*logger).Sync()
	lg.Debug("日志系统初始化完成")
	return lg
}

func (l *builder) build() log.Logger {
	lg, err := l.conf.Build(l.cores())
	checkErr(err)
	return &logger{base: lg.WithOptions(zap.AddCallerSkip(1))}
}

func (l *builder) cores() zap.Option {
	cores := make([]zapcore.Core, 0)
	cores = append(cores, l.fileoutCores()...)
	cores = append(cores, l.stdoutCores()...)

	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	})
}

func (l *builder) fileoutCores() []zapcore.Core {
	cores := make([]zapcore.Core, 0)
	if l.conf.toFile {
		// 文件输出去掉颜色装饰
		encode := l.conf.EncoderConfig
		encode.EncodeLevel = zapcore.LowercaseLevelEncoder
		fileEncoder := zapcore.NewJSONEncoder(encode)
		cores = append(cores,
			zapcore.NewCore(fileEncoder, newFileWriteASyncer(l.conf.appName, l.conf.fileDir, "error", l.conf.fileAsync), l.priority(zapcore.ErrorLevel)),
			zapcore.NewCore(fileEncoder, newFileWriteASyncer(l.conf.appName, l.conf.fileDir, "warn", l.conf.fileAsync), l.priority(zapcore.WarnLevel)),
			zapcore.NewCore(fileEncoder, newFileWriteASyncer(l.conf.appName, l.conf.fileDir, "info", l.conf.fileAsync), l.priority(zapcore.InfoLevel)),
			zapcore.NewCore(fileEncoder, newFileWriteASyncer(l.conf.appName, l.conf.fileDir, "debug", l.conf.fileAsync), l.priority(zapcore.DebugLevel)),
		)
	}
	return cores
}

func (l *builder) stdoutCores() []zapcore.Core {
	cores := make([]zapcore.Core, 0)
	if l.conf.stdout {
		stdoutEncoder := zapcore.NewConsoleEncoder(l.conf.EncoderConfig)
		if strings.ToLower(l.conf.stdoutType) == "json" {
			stdoutEncoder = zapcore.NewJSONEncoder(l.conf.EncoderConfig)
		}
		debugConsoleWS := zapcore.Lock(os.Stdout) // 控制台标准输出
		errorConsoleWS := zapcore.Lock(os.Stderr)
		cores = append(cores,
			zapcore.NewCore(stdoutEncoder, errorConsoleWS, l.priority(zapcore.ErrorLevel)),
			zapcore.NewCore(stdoutEncoder, debugConsoleWS, l.priority(zapcore.WarnLevel)),
			zapcore.NewCore(stdoutEncoder, debugConsoleWS, l.priority(zapcore.InfoLevel)),
			zapcore.NewCore(stdoutEncoder, debugConsoleWS, l.priority(zapcore.DebugLevel)),
		)
	}
	return cores
}

func (l *builder) priority(level zapcore.Level) zap.LevelEnablerFunc {
	return func(lvl zapcore.Level) bool {
		return lvl == level && level >= l.conf.Level.Level()
	}
}
