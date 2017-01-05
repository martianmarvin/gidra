package config

import (
	"errors"
	"fmt"
)

var (
	// ErrParse is a generic parsing error
	ErrParse = errors.New("Could not parse config")
	// ErrRequired means a required field is missing
	ErrRequired = errors.New("Required field is missing")
)

// KeyError is a custom error for a key without a registered parser
type KeyError struct {
	Name string
	// The underlying error
	Err  error
	Line string
}

func (e KeyError) Error() string {
	return "KeyError: " + e.Name + e.Err.Error() + "\n" +
		e.Line
}

// ValueError is a custom error for any error parsing/validating a value
type ValueError struct {
	Name string
	// The underlying error
	Err  error
	Line string
}

func (e ValueError) Error() string {
	return fmt.Sprintf("ValueError: Error parsing value %s: %s\n%s", e.Name, e.Err, e.Line)
}

// NewValueError wraps a regular error into a ValueError, or returns nil if the
// error is nil
func NewValueError(name string, err error) error {
	if err == nil {
		return nil
	}
	return ValueError{Name: name, Err: err}
}
