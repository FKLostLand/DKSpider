package FKLog

import (
	"FKBase"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileLogWriter struct {
	lg                *log.Logger
	mw                *FKBase.MuxWriter
	startLock         sync.Mutex // 保证同时只会有一个log进行写入的锁
	Filename          string     `json:"filename"` // 文件名
	Maxlines          int        `json:"maxlines"` // 最大行数
	Maxsize           int        `json:"maxsize"`  // 文件最大大小
	Daily             bool       `json:"daily"`    // 是否每日更新日志文件
	Maxdays           int64      `json:"maxdays"`  // 如果按日更新文件，则自动更新最大天数
	Rotate            bool       `json:"rotate"`   // 是否开启自动文件更新（将影响行数达成后，文件大小达成后，日期更变后的自动处理）
	Level             int        `json:"level"`    // 写入文件的日志层级
	maxlines_curlines int
	maxsize_cursize   int
	daily_opendate    int
}

func createFileLogger() ILogger {
	w := &FileLogWriter{
		Filename: "",
		Maxlines: 1000000,
		Maxsize:  1 << 28, //256 MB
		Daily:    true,
		Maxdays:  365,
		Rotate:   true,
		Level:    FKBase.LevelDebug,
	}
	w.mw = new(FKBase.MuxWriter)
	w.lg = log.New(w.mw, "", log.Ldate|log.Ltime)
	return w
}

// Init file logger with json config.
// config like:
//	{
//	"filename":"logs/FK.log",
//	"maxlines":10000,
//	"maxsize":1<<30,
//	"daily":true,
//	"maxdays":15,
//	"rotate":true,
//	"level":3
//	}
func (w *FileLogWriter) Init(config map[string]interface{}) error {
	if config == nil {
		return errors.New("config can not be empty")
	}
	if filename, ok := config["filename"]; !ok || len(filename.(string)) == 0 {
		return errors.New("config must have filename")
	}
	conf, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = json.Unmarshal(conf, w)
	if err != nil {
		return err
	}
	return w.startLogger()
}

func (w *FileLogWriter) WriteMsg(msg string, level int) error {
	if level > w.Level {
		return nil
	}
	n := 28 + len(msg) // 24 stand for the length "2013/06/23 21:00:22 [DEBUG] "
	w.docheck(n)
	w.lg.Println(msg)
	return nil
}

func (w *FileLogWriter) Destroy() {
	w.mw.GetFd().Close()
}

func (w *FileLogWriter) Flush() {
	w.mw.GetFd().Sync()
}

func (w *FileLogWriter) startLogger() error {
	fd, err := w.createLogFile()
	if err != nil {
		return err
	}
	w.mw.SetFd(fd)
	return w.initFd()
}

func (w *FileLogWriter) docheck(size int) {
	w.startLock.Lock()
	defer w.startLock.Unlock()
	if w.Rotate && ((w.Maxlines > 0 && w.maxlines_curlines >= w.Maxlines) ||
		(w.Maxsize > 0 && w.maxsize_cursize >= w.Maxsize) ||
		(w.Daily && time.Now().Day() != w.daily_opendate)) {
		if err := w.doRotate(); err != nil {
			fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.Filename, err)
			return
		}
	}
	w.maxlines_curlines++
	w.maxsize_cursize += size
}

func (w *FileLogWriter) createLogFile() (*os.File, error) {
	fd, err := os.OpenFile(w.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	return fd, err
}

func (w *FileLogWriter) initFd() error {
	fd := w.mw.GetFd()
	finfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s\n", err)
	}
	w.maxsize_cursize = int(finfo.Size())
	w.daily_opendate = time.Now().Day()
	w.maxlines_curlines = 0
	if finfo.Size() > 0 {
		count, err := w.lines()
		if err != nil {
			return err
		}
		w.maxlines_curlines = count
	}
	return nil
}

func (w *FileLogWriter) lines() (int, error) {
	fd, err := os.Open(w.Filename)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	buf := make([]byte, 32768) // 32k
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

func (w *FileLogWriter) doRotate() error {
	_, err := os.Lstat(w.Filename)
	if err == nil {
		// 文件存在，查找新的可用数字
		num := 1
		fname := ""
		for ; err == nil && num <= 999; num++ {
			fname = w.Filename + fmt.Sprintf(".%s.%03d", time.Now().Format("2006-01-02"), num)
			_, err = os.Lstat(fname)
		}
		if err == nil {
			return fmt.Errorf("Rotate: Cannot find free log number to rename %s\n", w.Filename)
		}

		// 锁死Logger的io.Writer
		w.mw.Lock()
		defer w.mw.Unlock()

		fd := w.mw.GetFd()
		fd.Close()

		// 文件重命名之前，关闭Fd
		err = os.Rename(w.Filename, fname)
		if err != nil {
			return fmt.Errorf("Rotate: %s\n", err)
		}

		// 重启Logger
		err = w.startLogger()
		if err != nil {
			return fmt.Errorf("Rotate StartLogger: %s\n", err)
		}

		go w.deleteOldLog()
	}

	return nil
}

func (w *FileLogWriter) deleteOldLog() {
	dir := filepath.Dir(w.Filename)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				returnErr = fmt.Errorf("Unable to delete old log '%s', error: %+v", path, r)
				fmt.Println(returnErr)
			}
		}()

		if !info.IsDir() && info.ModTime().Unix() < (time.Now().Unix()-60*60*24*w.Maxdays) {
			if strings.HasPrefix(filepath.Base(path), filepath.Base(w.Filename)) {
				os.Remove(path)
			}
		}
		return
	})
}
