// Package files stores uploaded documents. Bytes go to a Store (local volume today, object
// storage later); ownership and access rules live in Postgres under RLS.
package files

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Store is the byte store behind the files table. Keys are "<tenant_id>/<file_id>".
type Store interface {
	Put(ctx context.Context, key string, r io.Reader) error
	Open(ctx context.Context, key string) (io.ReadSeekCloser, error)
	Delete(ctx context.Context, key string) error
}

// LocalStore keeps files under Root (a Docker volume in production).
type LocalStore struct{ Root string }

func (s LocalStore) path(key string) (string, error) {
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return "", errors.New("bad key")
	}
	return filepath.Join(s.Root, filepath.FromSlash(key)), nil
}

func (s LocalStore) Put(_ context.Context, key string, r io.Reader) error {
	p, err := s.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return err
	}
	tmp := p + ".part"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, p)
}

func (s LocalStore) Open(_ context.Context, key string) (io.ReadSeekCloser, error) {
	p, err := s.path(key)
	if err != nil {
		return nil, err
	}
	return os.Open(p)
}

func (s LocalStore) Delete(_ context.Context, key string) error {
	p, err := s.path(key)
	if err != nil {
		return err
	}
	err = os.Remove(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// FromEnv picks the store: TAXIOS_FILES_DIR (default ./data/files for local runs).
func FromEnv() Store {
	dir := os.Getenv("TAXIOS_FILES_DIR")
	if dir == "" {
		dir = "data/files"
	}
	return LocalStore{Root: dir}
}
