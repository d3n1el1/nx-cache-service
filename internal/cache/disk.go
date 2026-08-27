package cache

import (
	"context"
	"io"
	"os"
	"path"
)

type Disk struct {
	root string
}

var _ Store = (*Disk)(nil)

func NewDisk(dir string) (*Disk, error) {
	return &Disk{root: dir}, nil
}

func (d *Disk) Root() string { return d.root }

func (d *Disk) Get(ctx context.Context, project string, hash string) (io.ReadCloser, int64, error) {
	filePath := path.Join(d.root, project, hash)
	f, err := os.Open(filePath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, ErrNotFound
		}

		return nil, 0, err
	}

	fileInfo, err := f.Stat()

	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}

	return f, fileInfo.Size(), nil
}

func (d *Disk) Put(ctx context.Context, project string, hash string, r io.Reader) error {
	dir := path.Join(d.root, project)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filePath := path.Join(dir, hash)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0444)

	if err != nil {
		if os.IsExist(err) {
			return ErrExists
		}

		return err
	}

	if _, copyErr := io.Copy(f, r); copyErr != nil {
		_ = f.Close()
		_ = os.Remove(filePath)
		return copyErr
	}

	return f.Close()
}
