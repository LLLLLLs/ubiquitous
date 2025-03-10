package impl

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	rotls "github.com/lestrrat/go-file-rotatelogs"
)

// 文件日志异步写入 缓冲区超过阈值或倒计时结束时写入文件
type fileWriteASyncer struct {
	bufferLock sync.Mutex
	buffer     *bytes.Buffer
	fileWriter *rotls.RotateLogs
	ch         chan struct{}
	syncChan   chan struct{}
}

func newFileWriteASyncer(app, dir, subDir string, async bool) *fileWriteASyncer {
	fwa := &fileWriteASyncer{}
	fwa.init(app, dir, subDir)
	if async {
		go fwa.batchWriteLog()
	} else {
		go fwa.onceWriteLog()
	}
	return fwa
}

func (f *fileWriteASyncer) init(app, dir, subDir string) {
	f.bufferLock = sync.Mutex{}
	f.buffer = bytes.NewBuffer(make([]byte, 0, 4096))
	f.ch = make(chan struct{}, 1000)
	f.syncChan = make(chan struct{})
	f.initFileWriter(app, dir, subDir)
}

func (f *fileWriteASyncer) initFileWriter(app, dir, subDir string) {
	// 每小时一个文件
	logf, _ := rotls.New(f.buildFilePattern(dir, subDir),
		rotls.WithLinkName(f.buildFileLinkName(app, dir, subDir)),
		rotls.WithMaxAge(30*24*time.Hour),
		rotls.WithRotationTime(time.Minute),
	)
	f.fileWriter = logf
}

func (f *fileWriteASyncer) buildFilePattern(dir, subDir string) string {
	parentDir := filepath.Join(dir, subDir)
	err := os.MkdirAll(parentDir, os.ModePerm)
	checkErr(err)
	// windows不支持冒号作为文件名字 这边特殊处理
	timeFormat := "%Y-%m-%d-%H:00.log"
	if runtime.GOOS == "windows" {
		timeFormat = "%Y-%m-%d-%H-00.log"
	}

	return filepath.Join(parentDir, timeFormat)
}

func (f *fileWriteASyncer) buildFileLinkName(app, dir, subDir string) string {
	appPart := app
	if appPart != "" {
		appPart += "-"
	}
	return dir + fileSeparate + subDir + fileSeparate + appPart + subDir + ".log"
}

func (f *fileWriteASyncer) Write(data []byte) (n int, err error) {
	f.withBufferLock(func() {
		f.buffer.Write(data)
	})
	f.ch <- struct{}{} // 外部调用write后会修改data所以无法将data传递给channel
	return len(data), nil
}

func (f *fileWriteASyncer) Sync() error {
	f.syncChan <- struct{}{}
	return nil
}

// nolint 异步写
func (f *fileWriteASyncer) onceWriteLog() {
	defer func() {
		if err := recover(); err != nil {
			// 借用系统日志来记录错误 并 重启
			log.Printf("batch file write asyner panic:%+v", err)
			go f.onceWriteLog()
		}
	}()
	for {
		select {
		case <-f.ch:
			f.writeToFile()
		case <-f.syncChan:
			f.writeToFile()
		}
	}
}

// nolint 异步批量写
func (f *fileWriteASyncer) batchWriteLog() {
	defer func() {
		if err := recover(); err != nil {
			// 借用系统日志来记录错误 并 重启
			log.Printf("batch file write asyner panic:%+v", err)
			go f.batchWriteLog()
		}
	}()
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			if len(f.buffer.Bytes()) > 0 {
				f.writeToFile()
			}
		case <-f.ch:
			if len(f.buffer.Bytes()) >= 1024*4 {
				f.writeToFile()
			}
		case <-f.syncChan:
			if len(f.buffer.Bytes()) > 0 {
				f.writeToFile()
			}
		}
	}
}

func (f *fileWriteASyncer) writeToFile() {
	f.withBufferLock(func() {
		if len(f.buffer.Bytes()) == 0 {
			return
		}
		_, err := f.fileWriter.Write(f.buffer.Bytes())
		if err != nil {
			panic(err)
		}
		f.buffer.Reset()
	})
}

func (f *fileWriteASyncer) withBufferLock(h func()) {
	f.bufferLock.Lock()
	defer f.bufferLock.Unlock()
	h()
}
