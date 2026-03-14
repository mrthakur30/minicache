package cache

import "errors"

var (
    ErrNotFound = errors.New("minicache: key not found")
	ErrExpired = errors.New("minicache: key expired")
)

