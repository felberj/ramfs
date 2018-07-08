package ramfs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Node represents a single File in the filesystem
type Node struct {
	Mu      sync.Mutex
	Data    bytes.Buffer
	Name    string
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

// FileInfo holds information about the file
type FileInfo struct {
	name    string
	len     int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

// Name of the file
func (f *FileInfo) Name() string {
	return f.name
}

// Size of the file
func (f *FileInfo) Size() int64 {
	return f.len
}

// Mode of the file
func (f *FileInfo) Mode() os.FileMode {
	return f.mode
}

// ModTime of the file
func (f *FileInfo) ModTime() time.Time {
	return f.modTime
}

// IsDir whether the file is a directory
func (f *FileInfo) IsDir() bool {
	return f.isDir
}

// Sys returns nil
func (f *FileInfo) Sys() interface{} {
	return nil
}

// Stat returns the FileInfo of the file
func (n *Node) Stat() os.FileInfo {
	n.Mu.Lock()
	defer n.Mu.Unlock()
	return &FileInfo{
		name:    n.Name,
		len:     int64(n.Data.Len()),
		isDir:   n.IsDir,
		modTime: n.ModTime,
		mode:    n.Mode,
	}
}

// File is used to read and write to. The API should mirror the one for the os.File.
type File struct {
	node   *Node
	offset int
}

// Truncate truncates the file
func (f *File) Truncate(n int64) error {
	f.node.Mu.Lock()
	defer f.node.Mu.Unlock()
	f.node.Data.Truncate(int(n))
	return nil
}

// Write writes the content of the array into the file.
func (f *File) Write(p []byte) (int, error) {
	f.node.Mu.Lock()
	defer f.node.Mu.Unlock()
	d := f.node.Data.Bytes()
	wrote := 0
	for ; f.offset < len(d); f.offset++ {
		if wrote >= len(p) {
			break
		}
		d[f.offset] = p[wrote]
		wrote++
	}
	n, err := f.node.Data.Write(p[wrote:])
	f.offset += n
	return n + wrote, err
}

// Read reads the content from the file.
func (f *File) Read(p []byte) (int, error) {
	f.node.Mu.Lock()
	defer f.node.Mu.Unlock()
	d := f.node.Data.Bytes()
	if f.offset >= len(d) {
		return 0, &os.PathError{
			Op:   "read",
			Path: f.node.Name,
			Err:  io.EOF,
		}
	}
	n := len(p)
	if f.offset+n > len(d) {
		n = len(d) - f.offset
	}
	for i := range p {
		if i >= n {
			break
		}
		p[i] = d[f.offset]
		f.offset++
	}
	if len(p) != n {
		return n, &os.PathError{
			Op:   "read",
			Path: f.node.Name,
			Err:  io.EOF,
		}
	}
	return n, nil
}

// Seek sets the offset for the next Read or Write on file to offset,
// interpreted according to whence: 0 means relative to the origin of the file,
// 1 means relative to the current offset, and 2 means relative to the end.
// It returns the new offset and an error, if any.
func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case 0:
		f.offset = int(offset)
	case 1:
		f.offset += int(offset)
	default:
		return int64(f.offset), fmt.Errorf("seek %d not implemented", whence)
	}
	return int64(f.offset), nil
}

// Stat returns the FileInfo structure describing file.
// If there is an error, it will be of type *PathError.
func (f *File) Stat() (os.FileInfo, error) {
	return f.node.Stat(), nil
}

// Close closes the file
func (f *File) Close() error {
	return nil
}
