package cache

import (
	"context"
	"errors"
	"io"
)

var (
	ErrNotFound = errors.New("cache: artifact not found")
	ErrExists   = errors.New("cache: artifact already exists")
)

type Store interface {
	Get(ctx context.Context, project string, hash string) (body io.ReadCloser, size int64, err error)
	Put(ctx context.Context, project string, hash string, r io.Reader) error
	Flush(ctx context.Context, project string) error
}
