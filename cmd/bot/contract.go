package main

import "io"

// NamedCloser wraps io.Closer with Name() method
type NamedCloser interface {
	io.Closer
	Name() string
}
