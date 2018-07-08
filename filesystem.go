package ramfs

import (
	"io"
	"log"
	"os"
	"sync"
)

// Filesystem is used to hold all information about the filesystem.
type Filesystem struct {
	mu    sync.Mutex
	files map[string]*Node
}

// New creates a new Filesystem
func New() *Filesystem {
	return &Filesystem{
		files: make(map[string]*Node),
	}
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func (fs *Filesystem) Open(name string) (*File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile is the generalized open call; most users will use Open
// or Create instead. It opens the named file with specified flag
// (O_RDONLY etc.) and perm (before umask), if applicable. If successful,
// methods on the returned File can be used for I/O.
// If there is an error, it will be of type *PathError.
func (fs *Filesystem) OpenFile(name string, flag int, perm os.FileMode) (*File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	f, ok := fs.files[name]
	if !ok {
		if flag&os.O_CREATE == 0 {
			return nil, &os.PathError{
				Op:   "open",
				Err:  os.ErrNotExist,
				Path: name,
			}
		}
		f = &Node{
			Name: name,
			Mode: perm,
		}
		fs.files[name] = f
	}
	if (f.Mode.Perm() & perm.Perm()) != perm.Perm() {
		log.Printf("%x %x", f.Mode.Perm(), perm.Perm())
		// TODO is this check correct?
		return nil, &os.PathError{
			Op:   "open",
			Err:  os.ErrPermission,
			Path: name,
		}
	}
	file := &File{
		node: f,
	}
	if flag&os.O_TRUNC != 0 {
		file.Truncate(0)
	}

	return file, nil
}

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned
// File can be used for I/O; the associated file descriptor has mode
// O_RDWR.
// If there is an error, it will be of type *PathError.
func (fs *Filesystem) Create(name string) (*File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Chmod changes the mode of the named file to mode.
func (fs *Filesystem) Chmod(name string, mode os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	f, ok := fs.files[name]
	if !ok {
		return &os.PathError{
			Op:   "chmod",
			Err:  os.ErrExist,
			Path: name,
		}
	}
	f.Mode = mode
	return nil
}

// MapFile maps a file from the host system into the guest system.
func (fs *Filesystem) MapFile(hostname, guestname string) error {
	f, err := os.Open(hostname)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	fg, err := fs.Create(guestname)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fg, f); err != nil {
		return err
	}
	return fs.Chmod(guestname, stat.Mode())
}
