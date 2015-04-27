package dat

import (
	"errors"
)

var (
	// ErrNotFound ...
	ErrNotFound = errors.New("not found")
	// ErrNotUTF8 ...
	ErrNotUTF8 = errors.New("invalid UTF-8")
	// ErrInvalidSliceLength ...
	ErrInvalidSliceLength = errors.New("length of slice is 0. length must be >= 1")
	// ErrInvalidSliceValue ...
	ErrInvalidSliceValue = errors.New("trying to interpolate invalid slice value into query")
	// ErrInvalidValue ...
	ErrInvalidValue = errors.New("trying to interpolate invalid value into query")
	// ErrArgumentMismatch ...
	ErrArgumentMismatch = errors.New("mismatch between ? (placeholders) and arguments")
)
