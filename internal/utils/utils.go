package utils

import (
	"context"
	"encoding/json"
)

// PackVar packs a variable into a context.
type contextKey string

var (
	IDKey = contextKey("id")
)

func PackVar(ctx context.Context, key contextKey, value interface{}) context.Context {
	data, err := json.Marshal(value)
	if err != nil {
		return ctx
	}

	return context.WithValue(ctx, key, data)
}

// UnpackVar unpacks a variable from a context.
func UnpackVar(ctx context.Context, key contextKey, value interface{}) error {
	data, ok := ctx.Value(key).([]byte)
	if !ok {
		return nil
	}

	err := json.Unmarshal(data, value)
	if err != nil {
		return err
	}

	return nil
}
