package dumper

import (
	"bufio"
	"os"
	"sync"
)

//TODO add godoc
type SerialWriter interface {
	Write(b []byte) (int, error)
	Rewrite(b []byte) (int, error)
	ScanAll() ([]string, error)
	Clear() error
	Close() error
}

var (
	_ SerialWriter = (*File)(nil)
)

//TODO add godoc
type File struct {
	fd *os.File
	mx *sync.Mutex
}

//TODO add godoc
func NewFileStream(path string) (*File, error) {
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &File{
		fd: fd,
		mx: &sync.Mutex{},
	}, nil
}

//TODO add godoc
func (f *File) Write(b []byte) (int, error) {
	f.mx.Lock()
	defer f.mx.Unlock()
	return f.fd.Write(b)
}

//TODO add godoc
func (f *File) Rewrite(b []byte) (n int, err error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	f.fd.Seek(0, 0)

	if err = f.Clear(); err == nil {
		n, err = f.fd.Write(b)
	}

	return n, err
}

//TODO add godoc
func (f *File) ScanAll() ([]string, error) {
	f.fd.Seek(0, 0)
	sc := bufio.NewScanner(f.fd)
	rs := make([]string, 0)
	for sc.Scan() {
		rs = append(rs, sc.Text())
	}
	f.fd.Seek(0, 0)
	return rs, sc.Err()
}

//TODO add godoc
func (f *File) Clear() error {
	return f.fd.Truncate(0)
}

//TODO add godoc
func (f *File) Close() error {
	return f.fd.Close()
}
