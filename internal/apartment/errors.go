package apartment

import "errors"

var (
	ErrNilProvider     = errors.New("provider is nil")
	ErrInvalidPageSize = errors.New("max fetch pages must be positive")
)
