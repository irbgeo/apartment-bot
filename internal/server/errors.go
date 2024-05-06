package server

import "errors"

var (
	errLimitExceeded = errors.New("the limit on the number of filters is 1")
)
