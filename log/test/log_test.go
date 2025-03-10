package test

import (
	"github.com/LLLLLLs/ubiquitous/log"
	"github.com/LLLLLLs/ubiquitous/log/field"
	"github.com/LLLLLLs/ubiquitous/log/impl"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	d := log.NewDefaultLogger()
	d.With(field.Any("slice", []int{12345})).Debug("debug")
	d.Info("info", field.Int("test_id", 123))
	d.Warn("warn")
	d.Error("error")
}

func TestNop(t *testing.T) {
	nop := log.NewNopLogger()
	nop.Debug("debug")
	nop.Info("info")
	nop.Warn("warn")
	nop.Error("error")
}

func TestLogger(t *testing.T) {
	lg := impl.New(
		impl.WithStdout(true, "json"),
		impl.WithFileOut(true, "/Users/huajian/Workspace/huajian/ubiquitous/log/test/output"),
		impl.WithAppName("ubiquitous"),
		impl.WithLevel(impl.INFO),
		impl.WithRegionId(1234),
	)
	playerLg := lg.With(field.Int64("player_id", 1181947777))
	playerLg.Debug("debug")
	playerLg.Info("info")
	playerLg.Warn("warn", field.String("name", "test"))
	time.Sleep(time.Second * 10)
}
