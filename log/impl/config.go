package impl

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"path/filepath"
	"time"
)

type config struct {
	appName    string // 标识日志归属(game_server/mdb/...,会作为一个field打印出来)
	regionId   int32  // 区服id
	toFile     bool   // 是否输出到文件
	fileDir    string // 日志路径
	fileAsync  bool   // 日志是否异步写入(异步可减少文件读写次数)
	stdout     bool   // 是否打印到控制台
	stdoutType string // 标准输出格式[json,console]
	// maxSize    int    // 一个文件多少Ｍ大于该数字开始切分文件
	// maxBackups int    // MaxBackups是要保留的最大旧日志文件数
	// maxAge     int    // MaxAge是根据日期保留旧日志文件的最大天数
	zap.Config
}

type Option func(conf *config)

func newDefaultConfig() *config {
	return &config{
		appName:   "",
		regionId:  0,
		toFile:    false,
		fileDir:   "",
		fileAsync: false,
		stdout:    false,
		// maxSize:    100,
		// maxBackups: 30,
		// maxAge:     30,
		Config: zap.Config{
			Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
			Encoding:         "json",
			EncoderConfig:    DefaultEncoder(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		},
	}
}

func DefaultEncoder() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "date",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func WithAppName(appName string) Option {
	return func(conf *config) {
		conf.appName = appName
	}
}

func WithRegionId(regionId int32) Option {
	return func(conf *config) {
		conf.regionId = regionId
	}
}

// WithFileOut 输出到文件,默认是同步
func WithFileOut(fileout bool, dir string, async ...bool) Option {
	return func(conf *config) {
		conf.toFile = fileout
		if !fileout {
			return
		}
		if dir == "" {
			dir, _ = filepath.Abs(filepath.Dir(filepath.Join(".")))
			dir += fileSeparate + "logs" + fileSeparate
		}
		conf.fileDir = dir
		if len(async) != 0 {
			conf.fileAsync = async[0]
		}
	}
}

// WithStdout 输出到控制台
func WithStdout(stdout bool, stdoutType string) Option {
	return func(conf *config) {
		conf.stdout = stdout
		conf.stdoutType = stdoutType
	}
}

var levelToZapLv = map[Level]zapcore.Level{
	DEBUG: zapcore.DebugLevel,
	INFO:  zapcore.InfoLevel,
	WARN:  zapcore.WarnLevel,
	ERROR: zapcore.ErrorLevel,
}

// WithLevel 若等级不匹配，默认使用INFO
func WithLevel(level Level) Option {
	return func(conf *config) {
		targetLevel, has := levelToZapLv[level]
		if !has {
			targetLevel = zapcore.InfoLevel
		}
		conf.Config.Level = zap.NewAtomicLevelAt(targetLevel)
	}
}

func WithEncodeTime(format func(t time.Time) string) Option {
	return func(conf *config) {
		conf.Config.EncoderConfig.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(format(t))
		}
	}
}
