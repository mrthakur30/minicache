package cache

import "errors"

var (
    ErrNotFound = errors.New("minicache: key not found")
)