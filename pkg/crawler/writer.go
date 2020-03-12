package crawler

import (
	"io"
	"os"
	"path/filepath"
)

type DataStore struct {
	Directory string
}

func (d DataStore) Writer(path string) (io.WriteCloser, error) {
	path = filepath.Join(d.Directory, path)

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}
