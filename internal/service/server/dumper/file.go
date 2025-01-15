package dumper

import (
	"bufio"
	"os"
	"sync"
)

type FileStream struct {
	fd *os.File
	mx *sync.Mutex
}

func NewFileStream(path string) (*FileStream, error) {
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &FileStream{
		fd: fd,
		mx: &sync.Mutex{},
	}, nil
}

func (f *FileStream) Write(b []byte) (int, error) {
	f.mx.Lock()
	defer f.mx.Unlock()
	return f.fd.Write(b)
}

func (f *FileStream) Rewrite(b []byte) (int, error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	if err := f.fd.Truncate(0); err != nil {
		return 0, err
	}
	return f.fd.Write(b)
}

func (f *FileStream) Flush() error {
	f.mx.Lock()
	defer f.mx.Unlock()
	return f.fd.Sync()
}

func (f *FileStream) ScanAll() ([]string, error) {
	f.fd.Seek(0, 0)
	sc := bufio.NewScanner(f.fd)
	rs := make([]string, 0)
	for sc.Scan() {
		rs = append(rs, sc.Text())
	}
	f.fd.Seek(0, 0)
	return rs, sc.Err()
}

func (f *FileStream) Clear() error {
	return f.fd.Truncate(0)
}

func (f *FileStream) Close() error {
	return f.fd.Close()
}