package persistence

import (
	"context"
	"time"
)

type Record struct {
	Value string    `json:"value"`
	Exp   time.Time `json:"exp"`
}

type Persister interface {
	Save(ctx context.Context, data map[string]Record) error
	Load(ctx context.Context) (map[string]Record, error)
}
