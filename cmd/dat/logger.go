package main

import (
	"fmt"
	"os"
)

// Logger ...
type Logger struct{}

// Info writes string to stdout
func (Logger) Info(s string, args ...interface{}) {
	if len(args) > 0 {
		os.Stdout.WriteString(fmt.Sprintf(s, args...))
		return
	}

	os.Stdout.WriteString(s)
}

// Error writes string to stderr
func (Logger) Error(s string) {
	os.Stderr.WriteString(s)
}

var logger = Logger{}
