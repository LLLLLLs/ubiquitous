# Logger

> 基于``` go.uber.org/zap ```封装实现的日志库

## Quick Start

### 日志配置

目录：``` config/conf.yaml ```

```yaml
服务配置文件层级
...
...
...
logger:
  # 日志级别
  level: "debug"
  # 是否输出到文件
  fileout: true
  # 日志文件地址，当fileout=true时生效
  logDir: "~/data/log/game"
  # 是否输出到控制台
  stdout: true

```

### 一、获取日志实例

```go
package logger_test

import (
	logger "github.com/LLLLLLs/ubiquitous/log/impl"
)

// 1.新建日志实例
func newLogger() {
	_ = logger.New(
		logger.WithAppName("app_name"),        // 应用名称
		logger.WithLevel(logger.DEBUG),        // 日志等级,大于等于该等级的日志将被打印
		logger.WithFileOut(true, "/data/log"), // 日志是否输出到文件以及文件目录
		logger.WithStdout(false, "console"),   // 日志是否输出到控制台
	)
}

// 2.获取logger包预设的全局日志实例(app_name=campfire)
func getLogger() {
	_ = logger.GetLogger()
}

// 3.会为每次请求都创建一个实例(包含了本次请求的trace_id)
func msgHandler(sess *link.Session, in interface{}) {
	state := service.State(sess)
	_ = state.Logger()
}

```

### 二、打印日志

```go
package logger_test


func initLogger() {
	log := logger.GetLogger()
	// 携带字段
	log = log.With(
		field.Int64("player_id", 123456),
		field.String("player_name", "peter"),
	)
	// 打印日志
	log.Debug("this is a debug message")
	log.Info("this is a info message")
	log.Warn("this is a warn message")
	log.Error("this is a error message") // Error级别会携带 stack 信息
}
```

### 三、搭配kibana使用

#### 1、获取基础数据

- 左边选择index pattern (game-log)
- 右上角展开设置日志时间范围
- 获取基础数据信息```_source```
- 左边可以选择希望显示的字段

### 拓展 - 日志等级

- DEBUG: 指出细粒度信息事件对调试应用程序是非常有帮助的，主要用于开发过程中打印一些运行信息。
- INFO: 消息在粗粒度级别上突出强调应用程序的运行过程。打印一些你感兴趣的或者重要的信息，这个可以用于生产环境中输出程序运行的一些重要信息，但是不能滥用，避免打印过多的日志。
- WARN: 表明会出现潜在错误的情形，有些信息不是错误信息，但是也要给程序员的一些提示。
- ERROR: 指出虽然发生错误事件，但仍然不影响系统的继续运行。打印错误和异常信息，如果不想输出太多的日志，可以使用这个级别。(此级别日志会默认携带```stack```信息)