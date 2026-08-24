package cache

import (
	"context"
	"errors"
	"io"
)

type Disk struct {
	root string
}

var _ Store = (*Disk)(nil)

func NewDisk(dir string) (*Disk, error) {
	return &Disk{root: dir}, nil
}

func (d *Disk) Root() string { return d.root }

func (d *Disk) Get(ctx context.Context, hash string) (io.ReadCloser, int64, error) {
	return nil, 0, errors.New("cache: Get not implemented")
}

func (d *Disk) Put(ctx context.Context, hash string, r io.Reader) error {
	return errors.New("cache: Put not implemented")
}
