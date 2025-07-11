package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"
)

type FilePersister struct {
	mu   sync.Mutex
	path string
}

func NewFilePersister(path string) *FilePersister {
	return &FilePersister{
		path: path,
	}
}

func (fp *FilePersister) Save(_ context.Context, data map[string]Record) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	f, err := os.CreateTemp("", "kayvee-*.json")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(data); err != nil {
		return err
	}

	return os.Rename(f.Name(), fp.path)
}

func (fp *FilePersister) Load(_ context.Context) (map[string]Record, error) {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	f, err := os.Open(fp.path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]Record{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data map[string]Record
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
