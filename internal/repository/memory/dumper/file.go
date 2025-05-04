package dumper

import (
	"bufio"
	"io"
	"os"
	"sync"
)

// SerialWrite is the interface that group method for Writing, Rewriting and Scanning from source
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

// File is wrapper over simple file and implements SerialWriter interface
type File struct {
	fd *os.File
	mx *sync.Mutex
	io.Closer
}

// NewFileStream opens or create file and return pointer to entity or error if can not open file
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

// Write - writes data to file
func (f *File) Write(b []byte) (int, error) {
	f.mx.Lock()
	defer f.mx.Unlock()
	return f.fd.Write(b)
}

// Rewrite - deletes all data from file and writes new data
func (f *File) Rewrite(b []byte) (n int, err error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	f.fd.Seek(0, 0)

	if err = f.Clear(); err == nil {
		n, err = f.fd.Write(b)
	}

	return n, err
}

// ScanAll - scans all data from file and return slice of strings
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

// Clear - delete data in file and switch offset to 0
func (f *File) Clear() error {
	return f.fd.Truncate(0)
}

// Close - closes file
func (f *File) Close() error {
	return f.fd.Close()
}
