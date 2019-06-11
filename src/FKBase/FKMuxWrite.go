package FKBase

import (
	"os"
	"sync"
)

type MuxWriter struct {
	sync.Mutex
	Fd *os.File
}

func (l *MuxWriter) Write(b []byte) (int, error) {
	l.Lock()
	defer l.Unlock()
	return l.Fd.Write(b)
}

func (l *MuxWriter) SetFd(fd *os.File) {
	if l.Fd != nil {
		l.Fd.Close()
	}
	l.Fd = fd
}

func (l *MuxWriter) GetFd() *os.File {
	if l != nil {
		return l.Fd
	}
	return nil
}
